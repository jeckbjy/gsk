package broker

type IBroker interface {
	Connect() error
	Disconnect() error
	Init(...Option) error
	Publish(string, *Message, ...PublishOption) error
	Subscribe(string, Handler, ...SubscribeOption) (Subscriber, error)
	String() string
}

// Handler is used to process messages via a subscription of a topic.
// The handler is passed a publication interface which contains the
// message and optional Ack method to acknowledge receipt of the message.
type Handler func(Publication) error

type Message struct {
	Header map[string]string
	Body   []byte
}

// Publication is given to a subscription handler for processing
type Publication interface {
	Topic() string
	Message() *Message
	Ack() error
}

// Subscriber is a convenience return type for the Subscribe method
type Subscriber interface {
	Options() SubscribeOptions
	Topic() string
	Unsubscribe() error
}
