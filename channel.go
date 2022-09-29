package natschannel

import (
	"net"

	"github.com/nats-io/nats.go"
)

const (
	DefaultInboxSize = 256

	statusHeader       = "Status"
	noRespondersStatus = "503"
	reqTimeoutStatus   = "408"
)

// Option is a builder function for modifying Options.
type Option = func(opts *Options) error

// Options represents a collection of options used for controlling how nats.Conn are created and how the resultant
// Channel is bound to the nats.Conn.
type Options struct {
	Group       string
	NatsOptions []nats.Option
	InboxSize   int
}

// DefaultOptions returns the default options for the Channel.
func DefaultOptions() Options {
	return Options{
		InboxSize: DefaultInboxSize,
	}
}

// InboxSize specifies the size of the msg inbox channel used for receiving responses.
func InboxSize(size int) Option {
	return func(opts *Options) error {
		// TODO some simple validation
		opts.InboxSize = size
		return nil
	}
}

// Group specifies a NATS work queue group that the Channel will be bound to.
func Group(group string) Option {
	return func(opts *Options) error {
		opts.Group = group
		return nil
	}
}

// NatsOptions allows for passing nats.Option to the nats.Conn that is being dialed.
func NatsOptions(options ...nats.Option) Option {
	return func(opts *Options) error {
		opts.NatsOptions = options
		return nil
	}
}

// Channel implements the jrpc2 Channel interface over a NATS connection.
type Channel struct {
	subject string
	group   string

	conn  *nats.Conn
	inbox chan *nats.Msg
}

// Send implements the corresponding method of the jrpc2 Channel interface.
// Data is sent in a single NATS message.
func (c *Channel) Send(data []byte) error {
	// create a unique inbox subject for receiving the response
	replyTo := c.conn.NewRespInbox()
	// subscribe to the inbox subject
	if err := c.subscribe(replyTo); err != nil {
		return err
	}
	// publish the request
	return c.conn.PublishRequest(c.subject, replyTo, data)
}

// Recv implements the corresponding method of the Channel interface.
// The last message to have been received is read and it's payload returned.
func (c *Channel) Recv() ([]byte, error) {
	msg, ok := <-c.inbox
	// check if the channel has been closed
	if !ok {
		return nil, nats.ErrConnectionClosed
	}
	// check for a status response
	if len(msg.Data) == 0 && msg.Header.Get(statusHeader) == noRespondersStatus {
		return nil, nats.ErrNoResponders
	}
	// otherwise return the msg data
	return msg.Data, nil
}

// Close implements the corresponding method of the Channel interface.
// Any active subscriptions are drained, the inbox channel closed and then the connection closed.
func (c *Channel) Close() error {
	// check if closes was already called
	if c.conn.IsDraining() || c.conn.IsClosed() {
		return net.ErrClosed
	}

	// drain any interest first
	if err := c.conn.Drain(); err != nil {
		return err
	}

	// close inbox channel and the nats connection
	// TODO listen with close handler?
	close(c.inbox)
	c.conn.Close()
	return nil
}

// subscribe takes care of subscribing to a response subject based on whether a work group has been configured.
func (c *Channel) subscribe(replyTo string) error {
	var err error
	var sub *nats.Subscription

	if c.group != "" {
		sub, err = c.conn.ChanQueueSubscribe(replyTo, c.group, c.inbox)
	} else {
		sub, err = c.conn.ChanSubscribe(replyTo, c.inbox)
	}

	if err != nil {
		return err
	}
	// cleanup automatically after we receive one response
	return sub.AutoUnsubscribe(1)
}

// New wraps the given nats.Conn to implement the Channel interface.
func New(conn *nats.Conn, subject string, options ...Option) (*Channel, error) {
	opts, err := buildOptions(options...)
	if err != nil {
		return nil, err
	}
	return &Channel{
		subject: subject,
		group:   opts.Group,
		conn:    conn,
		inbox:   make(chan *nats.Msg, opts.InboxSize),
	}, nil
}

// Dial dials the specified nats url ("nats://...") and binds it to the provided subject with the given options.
func Dial(url string, subject string, options ...Option) (*Channel, error) {
	opts, err := buildOptions(options...)
	if err != nil {
		return nil, err
	}
	conn, err := nats.Connect(url, opts.NatsOptions...)
	if err != nil {
		return nil, err
	}

	return &Channel{
		subject: subject,
		group:   opts.Group,
		conn:    conn,
		inbox:   make(chan *nats.Msg, opts.InboxSize),
	}, nil
}

func buildOptions(options ...Option) (*Options, error) {
	opts := DefaultOptions()
	for _, opt := range options {
		if err := opt(&opts); err != nil {
			return nil, err
		}
	}
	return &opts, nil
}
