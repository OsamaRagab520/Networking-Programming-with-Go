package connections

import (
	"context"
	"net"
	"syscall"
	"testing"
	"time"
)

func DialTimeout(network string, address string, timeout time.Duration) (net.Conn, error) {

	d := net.Dialer{
		Control: func(_, address string, _ syscall.RawConn) error {
			return &net.DNSError{
				Err:         "connection timed out",
				Name:        address,
				Server:      "localhost:8080",
				IsTimeout:   true,
				IsTemporary: true}
		},
		Timeout: timeout}

	return d.Dial(network, address)
}
func TestDialTimeout(t *testing.T) {
	c, err := DialTimeout("tcp", "10.0.0.1:http", 5*time.Second)
	if err == nil {
		c.Close()
		t.Fatal("Connection did not time out")
	}

	nErr, ok := err.(net.Error)
	if !ok {
		t.Fatal(err)
	}
	if !nErr.Timeout() {
		t.Fatal("Error is not a timeout")
	}
}

func TestDialContext(t *testing.T) {
	// Create a deadline to wait for.
	dl := time.Now().Add(5 * time.Second)

	// Create a context that is both manually cancellable and will signal
	ctx, cancel := context.WithDeadline(context.Background(), dl)
	defer cancel()

	//
	d := net.Dialer{
		Control: func(_, _ string, _ syscall.RawConn) error {
			// Sleep long enough to reach the context's deadline.
			time.Sleep(5*time.Second + time.Millisecond)
			return nil
		},
	}

	conn, err := d.DialContext(ctx, "tcp", "10.0.0.0:http")
	if err == nil {
		conn.Close()
		t.Fatal("Connection did not time out")
	}

	// Check that the error is a timeout.
	nErr, ok := err.(net.Error)
	if !ok {
		t.Error(err)
	} else {
		if !nErr.Timeout() {
			t.Errorf("Error is not a timeout: %v", err)
		}
	}

	// Check that the context has timed out.
	if ctx.Err() != context.DeadlineExceeded {
		t.Errorf("Expected deadline exceeded; actual: %v", ctx.Err())
	}
}
