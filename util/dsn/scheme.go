package dsn

import (
	"fmt"
	stdpath "path"
	"strings"
)

var schemeMap map[string]*Scheme

func init() {
	schemes := []*Scheme{
		// core databases
		{"mssql", "", genSQLServer, 0, false, []string{"sqlserver"}},
		{"mysql", "", genMySQL, ProtoAll, false, []string{"mariadb", "maria", "percona", "aurora"}},
		{"godror", "", genOracle, 0, false, []string{"or", "ora", "oracle", "oci", "oci8", "odpi", "odpi-c"}},
		{"postgres", "", genPostgres, ProtoUnix, false, []string{"pg", "postgresql", "pgsql"}},
		{"sqlite3", "", genOpaque, 0, true, []string{"sqlite", "file"}},
	}

	schemeMap = make(map[string]*Scheme, len(schemes))
	for _, s := range schemes {
		Register(s)
	}
}

// Register registers a Scheme.
func Register(scheme *Scheme) {
	if scheme.Generator == nil {
		panic("must specify Generator when registering Scheme")
	}
	if scheme.Opaque && scheme.Proto&ProtoUnix != 0 {
		panic("scheme must support only Opaque or Unix protocols, not both")
	}
	if _, ok := schemeMap[scheme.Driver]; ok {
		panic(fmt.Sprintf("scheme %s already registered", scheme.Driver))
	}

	schemeMap[scheme.Driver] = scheme
	for _, alias := range scheme.Aliases {
		if scheme.Driver != alias {
			if _, ok := schemeMap[alias]; ok {
				panic(fmt.Sprintf("scheme %s not registered", alias))
			}
			schemeMap[alias] = scheme
		}
	}
}

// Unregister unregisters a Scheme and all associated aliases.
func Unregister(name string) *Scheme {
	scheme, ok := schemeMap[name]
	if ok {
		for _, alias := range scheme.Aliases {
			delete(schemeMap, alias)
		}
		delete(schemeMap, name)
		return scheme
	}

	return nil
}

// Proto are the allowed transport protocol types in a database URL scheme.
type Proto uint

// Proto types.
const (
	ProtoNone Proto = 0
	ProtoTCP  Proto = 1
	ProtoUDP  Proto = 2
	ProtoUnix Proto = 4
	ProtoAny  Proto = 8
	ProtoAll        = ProtoTCP | ProtoUDP | ProtoUnix
)

func (p Proto) Verify(proto string) bool {
	if p == ProtoNone {
		return false
	}
	switch {
	case p&ProtoAny != 0 && proto != "":
	case p&ProtoTCP != 0 && proto == "tcp":
	case p&ProtoUDP != 0 && proto == "udp":
	case p&ProtoUnix != 0 && proto == "unix":
	default:
		return false
	}

	return true
}

type Scheme struct {
	// Driver is the name of the SQL driver that will set as the Scheme in
	// Parse'd URLs, and is the driver name expected by the standard sql.Open
	// calls.
	Driver string

	// Override is the Go SQL driver to use instead of Driver.
	Override string

	// Generator is the func responsible for generating a URL based on parsed
	// URL information.
	//
	// Note: this func should not modify the passed URL.
	Generator func(u *URL) (string, error)

	// Proto are allowed protocol types for the scheme.
	Proto Proto

	// Opaque toggles Parse to not re-process URLs with an "opaque" component.
	Opaque bool

	// Aliases are any additional aliases for the scheme.
	Aliases []string
}

func (s *Scheme) getDriver() string {
	if s.Override != "" {
		return s.Override
	}

	return s.Driver
}

func genSQLServer(u *URL) (string, error) {
	host, port, dbname := hostname(u.Host), hostport(u.Host), u.Database

	// add instance name to host if present
	if i := strings.Index(dbname, "/"); i != -1 {
		host = host + `\` + dbname[:i]
		dbname = dbname[i+1:]
	}

	q := u.Options
	q.Set("Server", host)
	q.Set("Port", port)
	q.Set("Database", dbname)

	// add user/pass
	if u.Username != "" {
		q.Set("User ID", u.Username)
		q.Set("Password", u.Password)
	}

	// save host, port, dbname
	//if u.hostPortDB == nil {
	//	u.hostPortDB = []string{host, port, dbname}
	//}

	return genOptionsODBC(q, true), nil
}

func genMySQL(u *URL) (string, error) {
	host, port, dbname := hostname(u.Host), hostport(u.Host), u.Database

	builder := strings.Builder{}

	// build user/pass
	if u.Username != "" {
		builder.WriteString(u.Username)
		if u.Password != "" {
			builder.WriteString(":")
			builder.WriteString(u.Password)
		}
		builder.WriteString("@")
	}

	// resolve path
	if u.Proto == "unix" {
		if host == "" {
			dbname = "/" + dbname
		}
		host, dbname = resolveSocket(stdpath.Join(host, dbname))
		port = ""
	}

	// if host or proto is not empty
	if u.Proto != "unix" {
		if host == "" {
			host = "127.0.0.1"
		}
		if port == "" {
			port = "3306"
		}
	}
	if port != "" {
		port = ":" + port
	}

	builder.WriteString(u.Proto + "(" + host + port + ")")
	builder.WriteString("/" + dbname)
	builder.WriteString(genQueryOptions(u.Options))

	return builder.String(), nil
}

func genOracle(u *URL) (string, error) {
	// Easy Connect Naming method enables clients to connect to a database server
	// without any configuration. Clients use a connect string for a simple TCP/IP
	// address, which includes a host name and optional port and service name:
	// CONNECT username[/password]@[//]host[:port][/service_name][:server][/instance_name]

	host, port, service := hostname(u.Host), hostport(u.Host), u.Database
	var instance string

	// grab instance name from service name
	if i := strings.LastIndex(service, "/"); i != -1 {
		instance = service[i+1:]
		service = service[:i]
	}

	// build dsn
	var builder strings.Builder

	// build user/pass
	if u.Username != "" {
		builder.WriteString(u.Username)
		if u.Password != "" {
			builder.WriteString("/")
			builder.WriteString(u.Password)
		}
		builder.WriteString("@//")
	}

	builder.WriteString(host)
	if port != "" {
		builder.WriteString(":")
		builder.WriteString(port)
	}

	if service != "" {
		builder.WriteString("/" + service)
	}

	if instance != "" {
		builder.WriteString("/" + instance)
	}

	return builder.String(), nil
}

func genPostgres(u *URL) (string, error) {
	host, port, dbname := hostname(u.Host), hostport(u.Host), u.Database
	if host == "." {
		return "", ErrRelativePathNotSupported
	}

	// resolve path
	if u.Proto == "unix" {
		if host == "" {
			dbname = "/" + dbname
		}

		host, port, dbname = resolveDir(stdpath.Join(host, dbname))
	}

	q := u.Options
	q.Set("host", host)
	q.Set("port", port)
	q.Set("dbname", dbname)

	// add user/pass
	if u.Username != "" {
		q.Set("user", u.Username)
		q.Set("password", u.Password)
	}

	// save host, port, dbname
	//if u.hostPortDB == nil {
	//	u.hostPortDB = []string{host, port, dbname}
	//}

	return genOptions(q, "", "=", " ", ",", true), nil
}

func genOpaque(u *URL) (string, error) {
	if u.Opaque == "" {
		return "", ErrMissingPath
	}

	return u.Opaque + genQueryOptions(u.Options), nil
}
