package httpx

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

const (
	defaultMemory = 32 << 20 // 32 MB
	indexPage     = "index.html"
	defaultIndent = "  "
)

type _Context struct {
	request  *http.Request
	response *Response
	path     string
	pnames   []string
	pvalues  []string
	query    url.Values
	handler  HandlerFunc
	store    Map
	server   *Server
	lock     sync.RWMutex
}

func (c *_Context) writeContentType(value string) {
	header := c.Response().Header()
	if header.Get(HeaderContentType) == "" {
		header.Set(HeaderContentType, value)
	}
}

func (c *_Context) Request() *http.Request {
	return c.request
}

func (c *_Context) SetRequest(r *http.Request) {
	c.request = r
}

func (c *_Context) Response() *Response {
	return c.response
}

func (c *_Context) IsTLS() bool {
	return c.request.TLS != nil
}

func (c *_Context) IsWebSocket() bool {
	upgrade := c.request.Header.Get(HeaderUpgrade)
	return strings.ToLower(upgrade) == "websocket"
}

func (c *_Context) Scheme() string {
	// Can't use `r.Request.URL.Scheme`
	// See: https://groups.google.com/forum/#!topic/golang-nuts/pMUkBlQBDF0
	if c.IsTLS() {
		return "https"
	}
	if scheme := c.request.Header.Get(HeaderXForwardedProto); scheme != "" {
		return scheme
	}
	if scheme := c.request.Header.Get(HeaderXForwardedProtocol); scheme != "" {
		return scheme
	}
	if ssl := c.request.Header.Get(HeaderXForwardedSsl); ssl == "on" {
		return "https"
	}
	if scheme := c.request.Header.Get(HeaderXUrlScheme); scheme != "" {
		return scheme
	}
	return "http"
}

func (c *_Context) RealIP() string {
	if ip := c.request.Header.Get(HeaderXForwardedFor); ip != "" {
		return strings.Split(ip, ", ")[0]
	}
	if ip := c.request.Header.Get(HeaderXRealIP); ip != "" {
		return ip
	}
	ra, _, _ := net.SplitHostPort(c.request.RemoteAddr)
	return ra
}

func (c *_Context) Path() string {
	return c.path
}

func (c *_Context) SetPath(p string) {
	c.path = p
}

func (c *_Context) Param(name string) string {
	for i, n := range c.pnames {
		if i < len(c.pvalues) {
			if n == name {
				return c.pvalues[i]
			}
		}
	}
	return ""
}

func (c *_Context) ParamNames() []string {
	return c.pnames
}

func (c *_Context) SetParamNames(names ...string) {
	c.pnames = names
}

func (c *_Context) ParamValues() []string {
	return c.pvalues[:len(c.pnames)]
}

func (c *_Context) SetParamValues(values ...string) {
	c.pvalues = values
}

func (c *_Context) QueryParam(name string) string {
	if c.query == nil {
		c.query = c.request.URL.Query()
	}
	return c.query.Get(name)
}

func (c *_Context) QueryParams() url.Values {
	if c.query == nil {
		c.query = c.request.URL.Query()
	}
	return c.query
}

func (c *_Context) QueryString() string {
	return c.request.URL.RawQuery
}

func (c *_Context) FormValue(name string) string {
	return c.request.FormValue(name)
}

func (c *_Context) FormParams() (url.Values, error) {
	if strings.HasPrefix(c.request.Header.Get(HeaderContentType), MIMEMultipartForm) {
		if err := c.request.ParseMultipartForm(defaultMemory); err != nil {
			return nil, err
		}
	} else {
		if err := c.request.ParseForm(); err != nil {
			return nil, err
		}
	}
	return c.request.Form, nil
}

func (c *_Context) FormFile(name string) (*multipart.FileHeader, error) {
	_, fh, err := c.request.FormFile(name)
	return fh, err
}

func (c *_Context) MultipartForm() (*multipart.Form, error) {
	err := c.request.ParseMultipartForm(defaultMemory)
	return c.request.MultipartForm, err
}

func (c *_Context) Cookie(name string) (*http.Cookie, error) {
	return c.request.Cookie(name)
}

func (c *_Context) SetCookie(cookie *http.Cookie) {
	http.SetCookie(c.Response(), cookie)
}

func (c *_Context) Cookies() []*http.Cookie {
	return c.request.Cookies()
}

func (c *_Context) Get(key string) interface{} {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.store[key]
}

func (c *_Context) Set(key string, val interface{}) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.store == nil {
		c.store = make(Map)
	}
	c.store[key] = val
}

func (c *_Context) Bind(i interface{}) error {
	return c.server.Binder.Bind(i, c)
}

func (c *_Context) Validate(i interface{}) error {
	if c.server.Validator == nil {
		return ErrValidatorNotRegistered
	}
	return c.server.Validator.Validate(i)
}

func (c *_Context) Render(code int, name string, data interface{}) (err error) {
	if c.server.Renderer == nil {
		return ErrRendererNotRegistered
	}
	buf := new(bytes.Buffer)
	if err = c.server.Renderer.Render(buf, name, data, c); err != nil {
		return
	}
	return c.HTMLBlob(code, buf.Bytes())
}

func (c *_Context) HTML(code int, html string) (err error) {
	return c.HTMLBlob(code, []byte(html))
}

func (c *_Context) HTMLBlob(code int, b []byte) (err error) {
	return c.Blob(code, MIMETextHTMLCharsetUTF8, b)
}

func (c *_Context) String(code int, s string) (err error) {
	return c.Blob(code, MIMETextPlainCharsetUTF8, []byte(s))
}

func (c *_Context) jsonPBlob(code int, callback string, i interface{}) (err error) {
	enc := json.NewEncoder(c.response)
	_, pretty := c.QueryParams()["pretty"]
	if c.server.Debug || pretty {
		enc.SetIndent("", "  ")
	}
	c.writeContentType(MIMEApplicationJavaScriptCharsetUTF8)
	c.response.WriteHeader(code)
	if _, err = c.response.Write([]byte(callback + "(")); err != nil {
		return
	}
	if err = enc.Encode(i); err != nil {
		return
	}
	if _, err = c.response.Write([]byte(");")); err != nil {
		return
	}
	return
}

func (c *_Context) json(code int, i interface{}, indent string) error {
	enc := json.NewEncoder(c.response)
	if indent != "" {
		enc.SetIndent("", indent)
	}
	c.writeContentType(MIMEApplicationJSONCharsetUTF8)
	c.response.Status = code
	return enc.Encode(i)
}

func (c *_Context) JSON(code int, i interface{}) (err error) {
	indent := ""
	if _, pretty := c.QueryParams()["pretty"]; c.server.Debug || pretty {
		indent = defaultIndent
	}
	return c.json(code, i, indent)
}

func (c *_Context) JSONPretty(code int, i interface{}, indent string) (err error) {
	return c.json(code, i, indent)
}

func (c *_Context) JSONBlob(code int, b []byte) (err error) {
	return c.Blob(code, MIMEApplicationJSONCharsetUTF8, b)
}

func (c *_Context) JSONP(code int, callback string, i interface{}) (err error) {
	return c.jsonPBlob(code, callback, i)
}

func (c *_Context) JSONPBlob(code int, callback string, b []byte) (err error) {
	c.writeContentType(MIMEApplicationJavaScriptCharsetUTF8)
	c.response.WriteHeader(code)
	if _, err = c.response.Write([]byte(callback + "(")); err != nil {
		return
	}
	if _, err = c.response.Write(b); err != nil {
		return
	}
	_, err = c.response.Write([]byte(");"))
	return
}

func (c *_Context) xml(code int, i interface{}, indent string) (err error) {
	c.writeContentType(MIMEApplicationXMLCharsetUTF8)
	c.response.WriteHeader(code)
	enc := xml.NewEncoder(c.response)
	if indent != "" {
		enc.Indent("", indent)
	}
	if _, err = c.response.Write([]byte(xml.Header)); err != nil {
		return
	}
	return enc.Encode(i)
}

func (c *_Context) XML(code int, i interface{}) (err error) {
	indent := ""
	if _, pretty := c.QueryParams()["pretty"]; c.server.Debug || pretty {
		indent = defaultIndent
	}
	return c.xml(code, i, indent)
}

func (c *_Context) XMLPretty(code int, i interface{}, indent string) (err error) {
	return c.xml(code, i, indent)
}

func (c *_Context) XMLBlob(code int, b []byte) (err error) {
	c.writeContentType(MIMEApplicationXMLCharsetUTF8)
	c.response.WriteHeader(code)
	if _, err = c.response.Write([]byte(xml.Header)); err != nil {
		return
	}
	_, err = c.response.Write(b)
	return
}

func (c *_Context) Blob(code int, contentType string, b []byte) (err error) {
	c.writeContentType(contentType)
	c.response.WriteHeader(code)
	_, err = c.response.Write(b)
	return
}

func (c *_Context) Stream(code int, contentType string, r io.Reader) (err error) {
	c.writeContentType(contentType)
	c.response.WriteHeader(code)
	_, err = io.Copy(c.response, r)
	return
}

func (c *_Context) File(file string) (err error) {
	f, err := os.Open(file)
	if err != nil {
		return NotFoundHandler(c)
	}
	defer f.Close()

	fi, _ := f.Stat()
	if fi.IsDir() {
		file = filepath.Join(file, indexPage)
		f, err = os.Open(file)
		if err != nil {
			return NotFoundHandler(c)
		}
		defer f.Close()
		if fi, err = f.Stat(); err != nil {
			return
		}
	}
	http.ServeContent(c.Response(), c.Request(), fi.Name(), fi.ModTime(), f)
	return
}

func (c *_Context) Attachment(file, name string) error {
	return c.contentDisposition(file, name, "attachment")
}

func (c *_Context) Inline(file, name string) error {
	return c.contentDisposition(file, name, "inline")
}

func (c *_Context) contentDisposition(file, name, dispositionType string) error {
	c.response.Header().Set(HeaderContentDisposition, fmt.Sprintf("%s; filename=%q", dispositionType, name))
	return c.File(file)
}

func (c *_Context) NoContent(code int) error {
	c.response.WriteHeader(code)
	return nil
}

func (c *_Context) Redirect(code int, url string) error {
	if code < 300 || code > 308 {
		return ErrInvalidRedirectCode
	}
	c.response.Header().Set(HeaderLocation, url)
	c.response.WriteHeader(code)
	return nil
}

func (c *_Context) Error(err error) {
	c.server.ErrorHandler(err, c)
}

func (c *_Context) Server() *Server {
	return c.server
}

func (c *_Context) Handler() HandlerFunc {
	return c.handler
}

func (c *_Context) SetHandler(h HandlerFunc) {
	c.handler = h
}

func (c *_Context) Reset(r *http.Request, w http.ResponseWriter) {
	c.request = r
	c.response.reset(w)
	c.query = nil
	c.handler = NotFoundHandler
	c.store = nil
	c.path = ""
	c.pnames = nil
	// NOTE: Don't reset because it has to have length c.echo.maxParam at all times
	// c.pvalues = nil
}
