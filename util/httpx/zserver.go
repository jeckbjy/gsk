package httpx

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"reflect"
	"runtime"
	"sync"
	"time"
)

// Binder is the interface that wraps the Bind method.
type Binder interface {
	Bind(i interface{}, c Context) error
}

// BindUnmarshaler is the interface used to wrap the UnmarshalParam method.
// Types that don't implement this, but do implement encoding.TextUnmarshaler
// will use that interface instead.
type BindUnmarshaler interface {
	// UnmarshalParam decodes and assigns a value from an form or query param.
	UnmarshalParam(param string) error
}

// Validator is the interface that wraps the Validate function.
type Validator interface {
	Validate(i interface{}) error
}

// Renderer is the interface that wraps the Render function.
type Renderer interface {
	Render(io.Writer, string, interface{}, Context) error
}

// Context represents the context of the current HTTP request. It holds request and
// response objects, path, path parameters, data and registered handler.
type Context interface {
	// Reset resets the context after request completes. It must be called along
	// with `Echo#AcquireContext()` and `Echo#ReleaseContext()`.
	// See `Echo#ServeHTTP()`
	Reset(r *http.Request, w http.ResponseWriter)

	// Request returns `*http.Request`.
	Request() *http.Request

	// SetRequest sets `*http.Request`.
	SetRequest(r *http.Request)

	// Response returns `*Response`.
	Response() *Response

	// IsTLS returns true if HTTP connection is TLS otherwise false.
	IsTLS() bool

	// IsWebSocket returns true if HTTP connection is WebSocket otherwise false.
	IsWebSocket() bool
	// Scheme returns the HTTP protocol scheme, `http` or `https`.
	Scheme() string
	// RealIP returns the client's network address based on `X-Forwarded-For`
	// or `X-Real-IP` request header.
	RealIP() string

	// Path returns the registered path for the handler.
	Path() string

	// SetPath sets the registered path for the handler.
	SetPath(p string)

	// Param returns path parameter by name.
	Param(name string) string

	// ParamNames returns path parameter names.
	ParamNames() []string

	// SetParamNames sets path parameter names.
	SetParamNames(names ...string)

	// ParamValues returns path parameter values.
	ParamValues() []string

	// SetParamValues sets path parameter values.
	SetParamValues(values ...string)

	// QueryParam returns the query param for the provided name.
	QueryParam(name string) string

	// QueryParams returns the query parameters as `url.Values`.
	QueryParams() url.Values

	// QueryString returns the URL query string.
	QueryString() string

	// FormValue returns the form field value for the provided name.
	FormValue(name string) string

	// FormParams returns the form parameters as `url.Values`.
	FormParams() (url.Values, error)

	// FormFile returns the multipart form file for the provided name.
	FormFile(name string) (*multipart.FileHeader, error)

	// MultipartForm returns the multipart form.
	MultipartForm() (*multipart.Form, error)

	// Cookie returns the named cookie provided in the request.
	Cookie(name string) (*http.Cookie, error)

	// SetCookie adds a `Set-Cookie` header in HTTP response.
	SetCookie(cookie *http.Cookie)

	// Cookies returns the HTTP cookies sent with the request.
	Cookies() []*http.Cookie

	// Get retrieves data from the context.
	Get(key string) interface{}

	// Set saves data in the context.
	Set(key string, val interface{})

	// Bind binds the request body into provided type `i`. The default binder
	// does it based on Content-Type header.
	Bind(i interface{}) error

	// Validate validates provided `i`. It is usually called after `Context#Bind()`.
	// Validator must be registered using `Echo#Validator`.
	Validate(i interface{}) error

	// Render renders a template with data and sends a text/html response with status
	// code. Renderer must be registered using `Echo.Renderer`.
	Render(code int, name string, data interface{}) error

	// HTML sends an HTTP response with status code.
	HTML(code int, html string) error

	// HTMLBlob sends an HTTP blob response with status code.
	HTMLBlob(code int, b []byte) error

	// String sends a string response with status code.
	String(code int, s string) error

	// JSON sends a JSON response with status code.
	JSON(code int, i interface{}) error

	// JSONPretty sends a pretty-print JSON with status code.
	JSONPretty(code int, i interface{}, indent string) error

	// JSONBlob sends a JSON blob response with status code.
	JSONBlob(code int, b []byte) error

	// JSONP sends a JSONP response with status code. It uses `callback` to construct
	// the JSONP payload.
	JSONP(code int, callback string, i interface{}) error

	// JSONPBlob sends a JSONP blob response with status code. It uses `callback`
	// to construct the JSONP payload.
	JSONPBlob(code int, callback string, b []byte) error

	// XML sends an XML response with status code.
	XML(code int, i interface{}) error

	// XMLPretty sends a pretty-print XML with status code.
	XMLPretty(code int, i interface{}, indent string) error

	// XMLBlob sends an XML blob response with status code.
	XMLBlob(code int, b []byte) error

	// Blob sends a blob response with status code and content type.
	Blob(code int, contentType string, b []byte) error

	// Stream sends a streaming response with status code and content type.
	Stream(code int, contentType string, r io.Reader) error

	// File sends a response with the content of the file.
	File(file string) error

	// Attachment sends a response as attachment, prompting client to save the
	// file.
	Attachment(file string, name string) error

	// Inline sends a response as inline, opening the file in the browser.
	Inline(file string, name string) error

	// NoContent sends a response with no body and a status code.
	NoContent(code int) error

	// Redirect redirects the request to a provided URL with status code.
	Redirect(code int, url string) error

	// Error invokes the registered HTTP error handler. Generally used by middleware.
	Error(err error)

	// Handler returns the matched handler by router.
	Handler() HandlerFunc

	// SetHandler sets the matched handler by router.
	SetHandler(h HandlerFunc)

	// Server returns the `Server` instance.
	Server() *Server
}

// Map defines a generic map of type `map[string]interface{}`.
type Map map[string]interface{}

// HandlerFunc defines a function to serve HTTP requests.
type HandlerFunc func(Context) error

// MiddlewareFunc defines a function to process middleware.
type MiddlewareFunc func(HandlerFunc) HandlerFunc

// ErrorHandler is a centralized HTTP error handler.
type ErrorHandler func(error, Context)

func NewServer() *Server {
	e := &Server{
		Server:    new(http.Server),
		TLSServer: new(http.Server),
		maxParam:  new(int),
	}
	e.pool.New = func() interface{} {
		return e.NewContext(nil, nil)
	}

	e.Server.Handler = e
	e.TLSServer.Handler = e
	e.ErrorHandler = e.DefaultHTTPErrorHandler
	e.Binder = &DefaultBinder{}
	e.router = NewRouter(e)
	e.routers = map[string]*Router{}
	return e
}

type Server struct {
	common
	premiddleware   []MiddlewareFunc
	middleware      []MiddlewareFunc
	maxParam        *int
	router          *Router
	routers         map[string]*Router
	notFoundHandler HandlerFunc
	pool            sync.Pool
	Server          *http.Server
	TLSServer       *http.Server
	Listener        net.Listener
	TLSListener     net.Listener
	DisableHTTP2    bool
	Debug           bool
	ErrorHandler    ErrorHandler
	Binder          Binder
	Validator       Validator
	Renderer        Renderer
}

// NewContext returns a Context instance.
func (e *Server) NewContext(r *http.Request, w http.ResponseWriter) Context {
	return &_Context{
		request:  r,
		response: NewResponse(w, e),
		store:    make(Map),
		server:   e,
		pvalues:  make([]string, *e.maxParam),
		handler:  NotFoundHandler,
	}
}

// Router returns the default router.
func (e *Server) Router() *Router {
	return e.router
}

// Routers returns the map of host => router.
func (e *Server) Routers() map[string]*Router {
	return e.routers
}

// DefaultHTTPErrorHandler is the default HTTP error handler. It sends a JSON response
// with status code.
func (e *Server) DefaultHTTPErrorHandler(err error, c Context) {
	var (
		code = http.StatusInternalServerError
		msg  interface{}
	)

	if he, ok := err.(*HTTPError); ok {
		code = he.Code
		msg = he.Message
		if he.Internal != nil {
			err = fmt.Errorf("%v, %v", err, he.Internal)
		}
	} else if e.Debug {
		msg = err.Error()
	} else {
		msg = http.StatusText(code)
	}
	if _, ok := msg.(string); ok {
		msg = Map{"message": msg}
	}

	// Send response
	if !c.Response().Committed {
		if c.Request().Method == http.MethodHead { // Issue #608
			err = c.NoContent(code)
		} else {
			err = c.JSON(code, msg)
		}
		//if err != nil {
		//	e.Logger.Error(err)
		//}
	}
}

// Pre adds middleware to the chain which is run before router.
//func (e *Server) Pre(middleware ...MiddlewareFunc) {
//	e.premiddleware = append(e.premiddleware, middleware...)
//}

// Use adds middleware to the chain which is run after router.
func (e *Server) Use(middleware ...MiddlewareFunc) {
	e.middleware = append(e.middleware, middleware...)
}

// CONNECT registers a new CONNECT route for a path with matching handler in the
// router with optional route-level middleware.
func (e *Server) CONNECT(path string, h HandlerFunc, m ...MiddlewareFunc) *Route {
	return e.Add(http.MethodConnect, path, h, m...)
}

// DELETE registers a new DELETE route for a path with matching handler in the router
// with optional route-level middleware.
func (e *Server) DELETE(path string, h HandlerFunc, m ...MiddlewareFunc) *Route {
	return e.Add(http.MethodDelete, path, h, m...)
}

// GET registers a new GET route for a path with matching handler in the router
// with optional route-level middleware.
func (e *Server) GET(path string, h HandlerFunc, m ...MiddlewareFunc) *Route {
	return e.Add(http.MethodGet, path, h, m...)
}

// HEAD registers a new HEAD route for a path with matching handler in the
// router with optional route-level middleware.
func (e *Server) HEAD(path string, h HandlerFunc, m ...MiddlewareFunc) *Route {
	return e.Add(http.MethodHead, path, h, m...)
}

// OPTIONS registers a new OPTIONS route for a path with matching handler in the
// router with optional route-level middleware.
func (e *Server) OPTIONS(path string, h HandlerFunc, m ...MiddlewareFunc) *Route {
	return e.Add(http.MethodOptions, path, h, m...)
}

// PATCH registers a new PATCH route for a path with matching handler in the
// router with optional route-level middleware.
func (e *Server) PATCH(path string, h HandlerFunc, m ...MiddlewareFunc) *Route {
	return e.Add(http.MethodPatch, path, h, m...)
}

// POST registers a new POST route for a path with matching handler in the
// router with optional route-level middleware.
func (e *Server) POST(path string, h HandlerFunc, m ...MiddlewareFunc) *Route {
	return e.Add(http.MethodPost, path, h, m...)
}

// PUT registers a new PUT route for a path with matching handler in the
// router with optional route-level middleware.
func (e *Server) PUT(path string, h HandlerFunc, m ...MiddlewareFunc) *Route {
	return e.Add(http.MethodPut, path, h, m...)
}

// TRACE registers a new TRACE route for a path with matching handler in the
// router with optional route-level middleware.
func (e *Server) TRACE(path string, h HandlerFunc, m ...MiddlewareFunc) *Route {
	return e.Add(http.MethodTrace, path, h, m...)
}

// Any registers a new route for all HTTP methods and path with matching handler
// in the router with optional route-level middleware.
func (e *Server) Any(path string, handler HandlerFunc, middleware ...MiddlewareFunc) []*Route {
	routes := make([]*Route, len(methods))
	for i, m := range methods {
		routes[i] = e.Add(m, path, handler, middleware...)
	}
	return routes
}

// Match registers a new route for multiple HTTP methods and path with matching
// handler in the router with optional route-level middleware.
func (e *Server) Match(methods []string, path string, handler HandlerFunc, middleware ...MiddlewareFunc) []*Route {
	routes := make([]*Route, len(methods))
	for i, m := range methods {
		routes[i] = e.Add(m, path, handler, middleware...)
	}
	return routes
}

// Static registers a new route with path prefix to serve static files from the
// provided root directory.
func (e *Server) Static(prefix, root string) *Route {
	if root == "" {
		root = "." // For security we want to restrict to CWD.
	}
	return e.static(prefix, root, e.GET)
}

// File registers a new route with path to serve a static file with optional route-level middleware.
func (e *Server) File(path, file string, m ...MiddlewareFunc) *Route {
	return e.file(path, file, e.GET, m...)
}

func (e *Server) add(host, method, path string, handler HandlerFunc, middleware ...MiddlewareFunc) *Route {
	name := handlerName(handler)
	router := e.findRouter(host)
	router.Add(method, path, func(c Context) error {
		h := handler
		// Chain middleware
		for i := len(middleware) - 1; i >= 0; i-- {
			h = middleware[i](h)
		}
		return h(c)
	})
	r := &Route{
		Method: method,
		Path:   path,
		Name:   name,
	}
	e.router.routes[method+path] = r
	return r
}

// Add registers a new route for an HTTP method and path with matching handler
// in the router with optional route-level middleware.
func (e *Server) Add(method, path string, handler HandlerFunc, middleware ...MiddlewareFunc) *Route {
	return e.add("", method, path, handler, middleware...)
}

// Host creates a new router group for the provided host and optional host-level middleware.
func (e *Server) Host(name string, m ...MiddlewareFunc) (g *Group) {
	e.routers[name] = NewRouter(e)
	g = &Group{host: name, server: e}
	g.Use(m...)
	return
}

// Group creates a new router group with prefix and optional group-level middleware.
func (e *Server) Group(prefix string, m ...MiddlewareFunc) (g *Group) {
	g = &Group{prefix: prefix, server: e}
	g.Use(m...)
	return
}

// URI generates a URI from handler.
func (e *Server) URI(handler HandlerFunc, params ...interface{}) string {
	name := handlerName(handler)
	return e.Reverse(name, params...)
}

// URL is an alias for `URI` function.
func (e *Server) URL(h HandlerFunc, params ...interface{}) string {
	return e.URI(h, params...)
}

// Reverse generates an URL from route name and provided parameters.
func (e *Server) Reverse(name string, params ...interface{}) string {
	uri := new(bytes.Buffer)
	ln := len(params)
	n := 0
	for _, r := range e.router.routes {
		if r.Name == name {
			for i, l := 0, len(r.Path); i < l; i++ {
				if r.Path[i] == ':' && n < ln {
					for ; i < l && r.Path[i] != '/'; i++ {
					}
					uri.WriteString(fmt.Sprintf("%v", params[n]))
					n++
				}
				if i < l {
					uri.WriteByte(r.Path[i])
				}
			}
			break
		}
	}
	return uri.String()
}

// Routes returns the registered routes.
func (e *Server) Routes() []*Route {
	routes := make([]*Route, 0, len(e.router.routes))
	for _, v := range e.router.routes {
		routes = append(routes, v)
	}
	return routes
}

// AcquireContext returns an empty `Context` instance from the pool.
// You must return the context by calling `ReleaseContext()`.
func (e *Server) AcquireContext() Context {
	return e.pool.Get().(Context)
}

// ReleaseContext returns the `Context` instance back to the pool.
// You must call it after `AcquireContext()`.
func (e *Server) ReleaseContext(c Context) {
	e.pool.Put(c)
}

// ServeHTTP implements `http.Handler` interface, which serves HTTP requests.
func (e *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Acquire context
	c := e.pool.Get().(Context)
	c.Reset(r, w)

	h := NotFoundHandler

	if e.premiddleware == nil {
		e.findRouter(r.Host).Find(r.Method, getPath(r), c)
		h = c.Handler()
		h = applyMiddleware(h, e.middleware...)
	} else {
		h = func(c Context) error {
			e.findRouter(r.Host).Find(r.Method, getPath(r), c)
			h := c.Handler()
			h = applyMiddleware(h, e.middleware...)
			return h(c)
		}
		h = applyMiddleware(h, e.premiddleware...)
	}

	// Execute chain
	if err := h(c); err != nil {
		e.ErrorHandler(err, c)
	}

	// Release context
	e.pool.Put(c)
}

// Start starts an HTTP server.
func (e *Server) Start(address string) error {
	e.Server.Addr = address
	return e.StartServer(e.Server)
}

// StartTLS starts an HTTPS server.
// If `certFile` or `keyFile` is `string` the values are treated as file paths.
// If `certFile` or `keyFile` is `[]byte` the values are treated as the certificate or key as-is.
func (e *Server) StartTLS(address string, certFile, keyFile interface{}) (err error) {
	var cert []byte
	if cert, err = filepathOrContent(certFile); err != nil {
		return
	}

	var key []byte
	if key, err = filepathOrContent(keyFile); err != nil {
		return
	}

	s := e.TLSServer
	s.TLSConfig = new(tls.Config)
	s.TLSConfig.Certificates = make([]tls.Certificate, 1)
	if s.TLSConfig.Certificates[0], err = tls.X509KeyPair(cert, key); err != nil {
		return
	}

	return e.startTLS(address)
}

func filepathOrContent(fileOrContent interface{}) (content []byte, err error) {
	switch v := fileOrContent.(type) {
	case string:
		return ioutil.ReadFile(v)
	case []byte:
		return v, nil
	default:
		return nil, ErrInvalidCertOrKeyType
	}
}

// StartAutoTLS starts an HTTPS server using certificates automatically installed from https://letsencrypt.org.
func (e *Server) StartAutoTLS(address string) error {
	s := e.TLSServer
	s.TLSConfig = new(tls.Config)
	//s.TLSConfig.GetCertificate = e.AutoTLSManager.GetCertificate
	//s.TLSConfig.NextProtos = append(s.TLSConfig.NextProtos, acme.ALPNProto)
	return e.startTLS(address)
}

func (e *Server) startTLS(address string) error {
	s := e.TLSServer
	s.Addr = address
	if !e.DisableHTTP2 {
		s.TLSConfig.NextProtos = append(s.TLSConfig.NextProtos, "h2")
	}
	return e.StartServer(e.TLSServer)
}

// StartServer starts a custom http server.
func (e *Server) StartServer(s *http.Server) (err error) {
	// Setup
	s.Handler = e

	if s.TLSConfig == nil {
		if e.Listener == nil {
			e.Listener, err = newListener(s.Addr)
			if err != nil {
				return err
			}
		}
		//if !e.HidePort {
		//	e.Logger.Printf("⇨ http server started on %s\n", e.Listener.Addr())
		//}
		return s.Serve(e.Listener)
	}
	if e.TLSListener == nil {
		l, err := newListener(s.Addr)
		if err != nil {
			return err
		}
		e.TLSListener = tls.NewListener(l, s.TLSConfig)
	}
	//if !e.HidePort {
	//	e.Logger.Printf("⇨ https server started on %s\n", e.TLSListener.Addr())
	//}
	return s.Serve(e.TLSListener)
}

// Close immediately stops the server.
// It internally calls `http.Server#Close()`.
func (e *Server) Close() error {
	if err := e.TLSServer.Close(); err != nil {
		return err
	}
	return e.Server.Close()
}

// Shutdown stops the server gracefully.
// It internally calls `http.Server#Shutdown()`.
func (e *Server) Shutdown(ctx context.Context) error {
	if err := e.TLSServer.Shutdown(ctx); err != nil {
		return err
	}
	return e.Server.Shutdown(ctx)
}

func (e *Server) findRouter(host string) *Router {
	if len(e.routers) > 0 {
		if r, ok := e.routers[host]; ok {
			return r
		}
	}
	return e.router
}

func getPath(r *http.Request) string {
	path := r.URL.RawPath
	if path == "" {
		path = r.URL.Path
	}
	return path
}

func handlerName(h HandlerFunc) string {
	t := reflect.ValueOf(h).Type()
	if t.Kind() == reflect.Func {
		return runtime.FuncForPC(reflect.ValueOf(h).Pointer()).Name()
	}
	return t.String()
}

// // PathUnescape is wraps `url.PathUnescape`
// func PathUnescape(s string) (string, error) {
// 	return url.PathUnescape(s)
// }

// tcpKeepAliveListener sets TCP keep-alive timeouts on accepted
// connections. It's used by ListenAndServe and ListenAndServeTLS so
// dead TCP connections (e.g. closing laptop mid-download) eventually
// go away.
type tcpKeepAliveListener struct {
	*net.TCPListener
}

func (ln tcpKeepAliveListener) Accept() (c net.Conn, err error) {
	if c, err = ln.AcceptTCP(); err != nil {
		return
	} else if err = c.(*net.TCPConn).SetKeepAlive(true); err != nil {
		return
	} else if err = c.(*net.TCPConn).SetKeepAlivePeriod(3 * time.Minute); err != nil {
		return
	}
	return
}

func newListener(address string) (*tcpKeepAliveListener, error) {
	l, err := net.Listen("tcp", address)
	if err != nil {
		return nil, err
	}
	return &tcpKeepAliveListener{l.(*net.TCPListener)}, nil
}

func applyMiddleware(h HandlerFunc, middleware ...MiddlewareFunc) HandlerFunc {
	for i := len(middleware) - 1; i >= 0; i-- {
		h = middleware[i](h)
	}
	return h
}
