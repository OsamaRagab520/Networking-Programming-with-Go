package connections

import (
	"io"
	"net"
	"testing"
	"time"
)

func TestDeadLine(t *testing.T) {
	sync := make(chan struct{})

	listener, err := net.Listen("tcp", "localhost:")

	if err != nil {
		t.Fatal(err)
	}
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			t.Log(err)
			return
		}

		defer func() {
			conn.Close()
			close(sync)
		}()

		// Set a deadline on the connection.
		err = conn.SetDeadline(time.Now().Add(5 * time.Second))
		if err != nil {
			t.Error(err)
			return
		}

		// Read a single byte, blocks to trigger the deadline.
		buf := make([]byte, 1)
		_, err = conn.Read(buf)

		nErr, ok := err.(net.Error)
		if !ok || !nErr.Timeout() {
			t.Errorf("Expected time out error; actual: %v", err)
		}
		sync <- struct{}{}

		// Reset the deadline.
		err = conn.SetDeadline(time.Now().Add(5 * time.Second))
		if err != nil {
			t.Error(err)
			return
		}

		_, err = conn.Read(buf)
		if err != nil {
			t.Error(err)
		}
	}()

	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}

	defer conn.Close()

	<-sync
	_, err = conn.Write([]byte("1"))
	if err != nil {
		t.Fatal(err)
	}

	// Wait for the server to terminate the connection.
	buf := make([]byte, 1)
	_, err = conn.Read(buf)
	if err != io.EOF {
		t.Errorf("expected server termination; actual: %v", err)
	}
}
