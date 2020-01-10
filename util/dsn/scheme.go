// Package dburl provides a standard, URL style mechanism for parsing and
// opening SQL database connection strings for Go. Provides standardized way to
// parse and open URLs for popular databases PostgreSQL, MySQL, SQLite3, Oracle
// Database, Microsoft SQL Server, in addition to most other SQL databases with
// a publicly available Go driver.
//
// Database Connection URL Overview
//
// Supported database connection URLs are of the form:
//
//   protocol+transport://user:pass@host/dbname?opt1=a&opt2=b
//   protocol:/path/to/file
//
// Where:
//
//   protocol  - driver name or alias (see below)
//   transport - "tcp", "udp", "unix" or driver name (odbc/oleodbc)                                  |
//   user      - username
//   pass      - password
//   host      - host
//   dbname*   - database, instance, or service name/id to connect to
//   ?opt1=... - additional database driver options
//                 (see respective SQL driver for available options)
//
// * for Microsoft SQL Server, /dbname can be /instance/dbname, where /instance
// is optional. For Oracle Database, /dbname is of the form /service/dbname
// where /service is the service name or SID, and /dbname is optional. Please
// see below for examples.
//
// Quickstart
//
// Database connection URLs in the above format can be parsed with Parse as such:
//
//   import (
//       "github.com/xo/dburl"
//   )
//   u, err := dburl.Parse("postgresql://user:pass@localhost/mydatabase/?sslmode=disable")
//   if err != nil { /* ... */ }
//
// Additionally, a simple helper, Open, is provided that will parse, open, and
// return a standard sql.DB database connection:
//
//   import (
//       "github.com/xo/dburl"
//   )
//   db, err := dburl.Open("sqlite:mydatabase.sqlite3?loc=auto")
//   if err != nil { /* ... */ }
//
// Example URLs
//
// The following are example database connection URLs that can be handled by
// Parse and Open:
//
//   postgres://user:pass@localhost/dbname
//   pg://user:pass@localhost/dbname?sslmode=disable
//   mysql://user:pass@localhost/dbname
//   mysql:/var/run/mysqld/mysqld.sock
//   sqlserver://user:pass@remote-host.com/dbname
//   mssql://user:pass@remote-host.com/instance/dbname
//   ms://user:pass@remote-host.com:port/instance/dbname?keepAlive=10
//   oracle://user:pass@somehost.com/sid
//   sap://user:pass@localhost/dbname
//   sqlite:/path/to/file.db
//   file:myfile.sqlite3?loc=auto
//   odbc+postgres://user:pass@localhost:port/dbname?option1=
//
// Protocol Schemes and Aliases
//
// The following protocols schemes (ie, driver) and their associated aliases
// are supported out of the box:
//
//   Database (scheme/driver)     | Protocol Aliases [real driver]
//   -----------------------------|--------------------------------------------
//   Microsoft SQL Server (mssql) | ms, sqlserver
//   MySQL (mysql)                | my, mariadb, maria, percona, aurora
//   Oracle Database (godror)     | or, ora, oci, oci8, odpi, odpi-c
//   PostgreSQL (postgres)        | pg, postgresql, pgsql
//   SQLite3 (sqlite3)            | sq, sqlite, file
//   -----------------------------|--------------------------------------------
//   Amazon Redshift (redshift)   | rs [postgres]
//   CockroachDB (cockroachdb)    | cr, cockroach, crdb, cdb [postgres]
//   MemSQL (memsql)              | me [mysql]
//   TiDB (tidb)                  | ti [mysql]
//   Vitess (vitess)              | vt [mysql]
//   -----------------------------|--------------------------------------------
//   Google Spanner (spanner)     | gs, google, span (not yet public)
//   -----------------------------|--------------------------------------------
//   MySQL (mymysql)              | zm, mymy
//   PostgreSQL (pgx)             | px
//   -----------------------------|--------------------------------------------
//   Apache Avatica (avatica)     | av, phoenix
//   Apache Ignite (ignite)       | ig, gridgain
//   Cassandra (cql)              | ca, cassandra, datastax, scy, scylla
//   ClickHouse (clickhouse)      | ch
//   Couchbase (n1ql)             | n1, couchbase
//   Cznic QL (ql)                | ql, cznic, cznicql
//   Firebird SQL (firebirdsql)   | fb, firebird
//   Microsoft ADODB (adodb)      | ad, ado
//   ODBC (odbc)                  | od
//   OLE ODBC (oleodbc)           | oo, ole, oleodbc [adodb]
//   Presto (presto)              | pr, prestodb, prestos, prs, prestodbs
//   SAP ASE (tds)                | ax, ase, sapase
//   SAP HANA (hdb)               | sa, saphana, sap, hana
//   Snowflake (snowflake)        | sf
//   Vertica (vertica)            | ve
//   VoltDB (voltdb)              | vo, volt, vdb
//
// Any protocol scheme alias:// can be used in place of protocol://, and will
// work identically with Parse and Open.
//
// Using
//
// Please note that the dburl package does not import actual SQL drivers, and
// only provides a standard way to parse/open respective database connection URLs.
//
// For reference, these are the following "expected" SQL drivers that would need
// to be imported:
//
//   Database (scheme/driver)     | Package
//   -----------------------------|-------------------------------------------------
//   Microsoft SQL Server (mssql) | github.com/denisenkom/go-mssqldb
//   MySQL (mysql)                | github.com/go-sql-driver/mysql
//   Oracle Database (godror)     | github.com/godror/godror
//   PostgreSQL (postgres)        | github.com/lib/pq
//   SQLite3 (sqlite3)            | github.com/mattn/go-sqlite3
//   -----------------------------|-------------------------------------------------
//   Amazon Redshift (redshift)   | github.com/lib/pq
//   CockroachDB (cockroachdb)    | github.com/lib/pq
//   MemSQL (memsql)              | github.com/go-sql-driver/mysql
//   TiDB (tidb)                  | github.com/go-sql-driver/mysql
//   Vitess (vitess)              | github.com/go-sql-driver/mysql
//   -----------------------------|-------------------------------------------------
//   Google Spanner (spanner)     | github.com/xo/spanner (not yet public)
//   -----------------------------|-------------------------------------------------
//   MySQL (mymysql)              | github.com/ziutek/mymysql/godrv
//   PostgreSQL (pgx)             | github.com/jackc/pgx/stdlib
//   -----------------------------|-------------------------------------------------
//   Apache Avatica (avatica)     | github.com/Boostport/avatica
//   Apache Ignite (ignite)       | github.com/amsokol/ignite-go-client/sql
//   Cassandra (cql)              | github.com/MichaelS11/go-cql-driver
//   ClickHouse (clickhouse)      | github.com/ClickHouse/clickhouse-go
//   Couchbase (n1ql)             | github.com/couchbase/go_n1ql
//   Cznic QL (ql)                | modernc.org/ql
//   Firebird SQL (firebirdsql)   | github.com/nakagami/firebirdsql
//   Microsoft ADODB (adodb)      | github.com/mattn/go-adodb
//   ODBC (odbc)                  | github.com/alexbrainman/odbc
//   OLE ODBC (oleodbc)*          | github.com/mattn/go-adodb
//   Presto (presto)              | github.com/prestodb/presto-go-client
//   SAP ASE (tds)                | github.com/thda/tds
//   SAP HANA (hdb)               | github.com/SAP/go-hdb/driver
//   Snowflake (snowflake)        | github.com/snowflakedb/gosnowflake
//   Vertica (vertica)            | github.com/vertica/vertica-sql-go
//   VoltDB (voltdb)              | github.com/VoltDB/voltdb-client-go/voltdbclient
//
// * OLE ODBC is a special alias for using the "MSDASQL.1" OLE provider with the
// ADODB driver on Windows. oleodbc:// URLs will be converted to the equivalent
// ADODB URL with "Extended Properties" having the respective ODBC parameters,
// including the underlying transport prootocol. As such, oleodbc+protocol://user:pass@host/dbname
// URLs are equivalent to adodb://MSDASQL.1/?Extended+Properties=.... on
// Windows. See GenOLEODBC for information regarding how URL components are
// mapped and passed to ADODB's Extended Properties parameter.
//
// URL Parsing Rules
//
// Parse and Open rely heavily on the standard net/url.URL type, as such
// parsing rules have the same conventions/semantics as any URL parsed by the
// standard library's net/url.Parse.
//
// Related Projects
//
// This package was written mainly to support xo (https://github.com/xo/xo)
// and usql (https://github.com/xo/usql).
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
