package dsn

import (
	"errors"
	"net/url"
	"reflect"
	"strconv"
	"strings"
)

var (
	ErrInvalidScheme            = errors.New("invalid scheme")
	ErrInvalidTransportProtocol = errors.New("invalid transport protocol")
	ErrMissingPath              = errors.New("missing path")
	ErrRelativePathNotSupported = errors.New("relative path not supported")
	ErrNotSupportOptionType     = errors.New("not support option type")
	ErrInvalidOption            = errors.New("invalid option")
)

func Parse(urlstr string) (*URL, error) {
	u, err := url.Parse(urlstr)
	if err != nil {
		return nil, err
	}

	if u.Scheme == "" {
		return nil, ErrInvalidScheme
	}

	var scheme, proto string
	var checkProto bool
	// parse scheme and proto
	if i := strings.IndexRune(u.Scheme, '+'); i != -1 {
		scheme = u.Scheme[:i]
		proto = u.Scheme[i+1:]
		checkProto = true
	} else {
		scheme = u.Scheme
		proto = "tcp"
	}

	sz, ok := schemeMap[scheme]
	if ok {
		if !sz.Opaque && u.Opaque != "" {
			return Parse(toOpaqueURL(u))
		}

		if sz.Opaque && u.Opaque == "" {
			// force Opaque
			u.Opaque, u.Host, u.Path, u.RawPath = u.Host+u.Path, "", "", ""
		} else if u.Host == "." || (u.Host == "" && strings.TrimPrefix(u.Path, "/") != "") {
			// force unix proto
			proto = "unix"
		}

		if checkProto || proto != "tcp" {
			if !sz.Proto.Verify(proto) {
				return nil, ErrInvalidTransportProtocol
			}
		}
	}

	dsn := &URL{}

	if u.User != nil {
		dsn.Username = u.User.Username()
		dsn.Password, _ = u.User.Password()
	}

	dsn.Driver = scheme
	dsn.Scheme = scheme
	dsn.Proto = proto
	dsn.Host = u.Host
	dsn.Hosts = strings.Split(u.Host, ",")
	dsn.Database = strings.Trim(u.Path, "/")
	dsn.Options = u.Query()
	dsn.Opaque = u.Opaque

	if sz != nil {
		sqlDSN, err := sz.Generator(dsn)
		if err != nil {
			return nil, err
		}
		dsn.Driver = sz.getDriver()
		dsn.DSN = sqlDSN
	}

	return dsn, nil
}

func toOpaqueURL(u *url.URL) string {
	q := ""
	if u.RawQuery != "" {
		q = "?" + u.RawQuery
	}
	f := ""
	if u.Fragment != "" {
		f = "#" + u.Fragment
	}

	return u.Scheme + "://" + u.Opaque + q + f
}

// URL is "data source name"
type URL struct {
	Driver   string
	DSN      string
	Scheme   string
	Proto    string
	Username string
	Password string
	Host     string
	Hosts    []string
	Database string
	Options  url.Values
	Opaque   string
	raw      string
}

func (dsn *URL) String() string {
	return dsn.raw
}

// Bind 绑定Option
func (dsn *URL) Bind(value interface{}) error {
	v := reflect.ValueOf(value)
	t := v.Type()
	if t.Kind() != reflect.Ptr {
		return ErrInvalidOption
	}
	v = v.Elem()
	t = t.Elem()

	for i := 0; i < t.NumField(); i++ {
		ft := t.Field(i)
		tag := ft.Tag.Get("dsn")
		if tag == "-" {
			continue
		}
		if tag == "" {
			tag = ft.Name
		}
		opt := dsn.Options.Get(tag)
		if opt == "" {
			continue
		}

		fv := v.Field(i)
		if err := setValue(&fv, ft.Type, opt); err != nil {
			return err
		}
	}
	return nil
}

func setValue(v *reflect.Value, t reflect.Type, opt string) error {
	switch t.Kind() {
	case reflect.Bool:
		x, err := strconv.ParseBool(opt)
		if err != nil {
			return err
		}
		v.SetBool(x)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		x, err := strconv.ParseInt(opt, 10, 64)
		if err != nil {
			return err
		}
		v.SetInt(x)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		x, err := strconv.ParseUint(opt, 10, 64)
		if err != nil {
			return err
		}
		v.SetUint(x)
	case reflect.Float32, reflect.Float64:
		x, err := strconv.ParseFloat(opt, 64)
		if err != nil {
			return err
		}
		v.SetFloat(x)
	case reflect.String:
		v.SetString(opt)
	default:
		return ErrNotSupportOptionType
	}

	return nil
}
