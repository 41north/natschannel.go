package natschannel

import (
	"testing"

	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats-server/v2/test"
	"github.com/nats-io/nats.go"
)

func runBasicServer(t *testing.T) *server.Server {
	t.Helper()
	opts := test.DefaultTestOptions
	opts.Port = -1
	return test.RunServer(&opts)
}

func client(t *testing.T, s *server.Server, opts ...nats.Option) *nats.Conn {
	t.Helper()
	nc, err := nats.Connect(s.ClientURL(), opts...)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	return nc
}

func pingPongTestServer(t *testing.T, s *server.Server, subject string, group string) *nats.Subscription {
	t.Helper()
	conn := client(t, s)

	var err error
	var sub *nats.Subscription

	if group != "" {
		sub, err = conn.QueueSubscribe(subject, group, func(msg *nats.Msg) {
			// respond with the data received
			if err := msg.Respond(msg.Data); err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
		})
	} else {
		sub, err = conn.Subscribe(subject, func(msg *nats.Msg) {
			// respond with the data received
			if err := msg.Respond(msg.Data); err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
		})
	}

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	return sub
}

func shutdownServer(t *testing.T, s *server.Server) {
	t.Helper()
	s.Shutdown()
	s.WaitForShutdown()
}
