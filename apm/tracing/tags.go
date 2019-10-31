package tracing

const (
	// SpanTypeWeb marks a span as an HTTP server request.
	SpanTypeWeb = "web"

	// SpanTypeHTTP marks a span as an HTTP client request.
	SpanTypeHTTP = "http"

	// SpanTypeSQL marks a span as an SQL operation. These spans may
	// have an "sql.command" tag.
	SpanTypeSQL = "sql"

	// SpanTypeCassandra marks a span as a Cassandra operation. These
	// spans may have an "sql.command" tag.
	SpanTypeCassandra = "cassandra"

	// SpanTypeRedis marks a span as a Redis operation. These spans may
	// also have a "redis.raw_command" tag.
	SpanTypeRedis = "redis"

	// SpanTypeMemcached marks a span as a memcached operation.
	SpanTypeMemcached = "memcached"

	// SpanTypeMongoDB marks a span as a MongoDB operation.
	SpanTypeMongoDB = "mongodb"

	// SpanTypeElasticSearch marks a span as an ElasticSearch operation.
	// These spans may also have an "elasticsearch.body" tag.
	SpanTypeElasticSearch = "elasticsearch"

	// SpanTypeLevelDB marks a span as a leveldb operation
	SpanTypeLevelDB = "leveldb"

	// SpanTypeDNS marks a span as a DNS operation.
	SpanTypeDNS = "dns"

	// SpanTypeMessageConsumer marks a span as a queue operation
	SpanTypeMessageConsumer = "queue"

	// SpanTypeMessageProducer marks a span as a queue operation.
	SpanTypeMessageProducer = "queue"

	// SpanTypeConsul marks a span as a Consul operation.
	SpanTypeConsul = "consul"
)

// tags
const (
	// TargetHost sets the target host address.
	TargetHost = "out.host"

	// TargetPort sets the target host port.
	TargetPort = "out.port"

	// SamplingPriority is the tag that marks the sampling priority of a span.
	SamplingPriority = "sampling.priority"

	// SQLType sets the sql type tag.
	SQLType = "sql"

	// SQLQuery sets the sql query tag on a span.
	SQLQuery = "sql.query"

	// HTTPMethod specifies the HTTP method used in a span.
	HTTPMethod = "http.method"

	// HTTPCode sets the HTTP status code as a tag.
	HTTPCode = "http.status_code"

	// HTTPURL sets the HTTP URL for a span.
	HTTPURL = "http.url"

	// TODO: In the next major version, prefix these constants (SpanType, etc)
	// with "Key*" (KeySpanType, etc) to more easily differentiate between
	// constants representing tag values and constants representing keys.

	// SpanName is a pseudo-key for setting a span's operation name by means of
	// a tag. It is mostly here to facilitate vendor-agnostic frameworks like Opentracing
	// and OpenCensus.
	SpanName = "span.name"

	// SpanType defines the Span type (web, db, cache).
	SpanType = "span.type"

	// ServiceName defines the Service name for this Span.
	ServiceName = "service.name"

	// ResourceName defines the Resource name for the Span.
	ResourceName = "resource.name"

	// Error specifies the error tag. It's value is usually of type "error".
	Error = "error"

	// ErrorMsg specifies the error message.
	ErrorMsg = "error.msg"

	// ErrorType specifies the error type.
	ErrorType = "error.type"

	// ErrorStack specifies the stack dump.
	ErrorStack = "error.stack"

	// ErrorDetails holds details about an error which implements a formatter.
	ErrorDetails = "error.details"

	// Environment specifies the environment to use with a trace.
	Environment = "env"

	// EventSampleRate specifies the rate at which this span will be sampled
	// as an APM event.
	EventSampleRate = "_dd1.sr.eausr"

	// AnalyticsEvent specifies whether the span should be recorded as a Trace
	// Search & Analytics event.
	AnalyticsEvent = "analytics.event"

	// ManualKeep is a tag which specifies that the trace to which this span
	// belongs to should be kept when set to true.
	ManualKeep = "manual.keep"

	// ManualDrop is a tag which specifies that the trace to which this span
	// belongs to should be dropped when set to true.
	ManualDrop = "manual.drop"
)

// db
const (
	// DBApplication indicates the application using the database.
	DBApplication = "db.application"
	// DBName indicates the database name.
	DBName = "db.name"
	// DBType indicates the type of Database.
	DBType = "db.type"
	// DBInstance indicates the instance name of Database.
	DBInstance = "db.instance"
	// DBUser indicates the user name of Database, e.g. "readonly_user" or "reporting_user".
	DBUser = "db.user"
	// DBStatement records a database statement for the given database type.
	DBStatement = "db.statement"
)

// peer
const (
	// PeerHostIPV4 records IPv4 host address of the peer.
	PeerHostIPV4 = "peer.ipv4"
	// PeerHostIPV6 records the IPv6 host address of the peer.
	PeerHostIPV6 = "peer.ipv6"
	// PeerService records the service name of the peer service.
	PeerService = "peer.service"
	// PeerHostname records the host name of the peer.
	PeerHostname = "peer.hostname"
	// PeerPort records the port number of the peer.
	PeerPort = "peer.port"
)

// Standard system metadata names
const (
	// The pid of the traced process
	Pid = "system.pid"
)
