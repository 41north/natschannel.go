package natschannel

import (
	"crypto/rand"
	"fmt"
	"net"
	"testing"

	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
)

func ExampleNew() {
	conn, err := nats.Connect("nats://localhost:4222")
	if err != nil {
		panic(err)
	}
	channel, err := New(conn, "foo.bar")
	if err != nil {
		panic(err)
	}

	err = channel.Close()
}

func ExampleDial() {
	channel, err := Dial("nats://localhost:4222", "foo.bar")
	if err != nil {
		panic(err)
	}

	err = channel.Close()
}

func TestChannel_New(t *testing.T) {
	s := runBasicServer(t)
	defer shutdownServer(t, s)

	conn := client(t, s)

	channel, err := New(conn, "foo.bar", InboxSize(128))
	assert.Nil(t, err)

	assert.Nil(t, channel.Close())
	assert.Equal(t, 128, cap(channel.inbox))
}

func TestChannel_Dial(t *testing.T) {
	s := runBasicServer(t)
	defer shutdownServer(t, s)

	channel, err := Dial(s.ClientURL(), "foo.bar")
	assert.Nil(t, err)
	assert.Equal(t, DefaultInboxSize, cap(channel.inbox))

	assert.Nil(t, channel.Close())
}

func TestChannel_SendAndReceive(t *testing.T) {
	s := runBasicServer(t)
	defer shutdownServer(t, s)

	subject := "foo.bar"

	// create a client
	conn := client(t, s)
	channel, err := New(conn, subject, NatsOptions(nats.RetryOnFailedConnect(false)))
	assert.Nil(t, err)

	// attempt a send and recv before there is a responder
	assert.Error(t, nats.ErrNoResponders, channel.Send([]byte("hello")))
	_, err = channel.Recv()
	assert.Error(t, nats.ErrNoResponders, err)

	// subsequent send and recv result in a closed error
	assert.Error(t, net.ErrClosed, channel.Send([]byte("world")))
	_, err = channel.Recv()
	assert.Error(t, net.ErrClosed, err)

	// create a responder
	pingPongTestResponder(t, s, subject, "")

	// re-init the channel
	channel, err = New(conn, subject)
	assert.Nil(t, err)

	// send some random data and check that we received it back
	for i := 0; i < 1000; i++ {
		bytes := []byte(fmt.Sprintf("data: %d", i+1))
		// send
		assert.Nil(t, channel.Send(bytes))
		// receive
		received, err := channel.Recv()
		assert.Nil(t, err)
		assert.False(t, len(received) == 0)
		assert.Equal(t, bytes, received)
	}

	// close the channel
	assert.Nil(t, channel.Close())
	// subsequent calls to close should return an error
	assert.Error(t, net.ErrClosed, channel.Close())
}

func TestChannel_GroupSendAndReceive(t *testing.T) {
	s := runBasicServer(t)
	defer shutdownServer(t, s)

	subject := "foo.bar"
	group := "test_group"

	// create a client
	conn := client(t, s)
	channel, err := New(conn, subject, Group(group))
	assert.Nil(t, err)
	assert.Equal(t, group, channel.group)

	// create a few test servers
	for i := 0; i < 3; i++ {
		pingPongTestResponder(t, s, subject, group)
	}

	// send some random data and check that we received it back
	for i := 0; i < 1000; i++ {
		// generate some random bytes
		bytes := make([]byte, 128)
		rand.Read(bytes)
		// send
		assert.Nil(t, channel.Send(bytes))
		// receive
		received, err := channel.Recv()
		assert.Nil(t, err)
		assert.Equal(t, bytes, received)
	}
}

func TestChannel_Close(t *testing.T) {
	s := runBasicServer(t)
	defer shutdownServer(t, s)

	conn := client(t, s)
	channel, err := New(conn, "foo.bar")
	assert.Nil(t, err)

	// close should happen without issue
	assert.Nil(t, channel.Close())
	// subsequent calls to close should return a closed error
	assert.Error(t, net.ErrClosed, channel.Close())
	// calls to send and receive should also return a closed error
	assert.Error(t, net.ErrClosed, channel.Send([]byte{}))
	_, err = channel.Recv()
	assert.Error(t, net.ErrClosed, err)
}
