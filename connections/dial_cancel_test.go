package connections

import (
	"context"
	"net"
	"sync"
	"syscall"
	"testing"
	"time"
)

func TestDialContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	sync := make(chan struct{})

	go func() {
		defer func() {
			sync <- struct{}{}
		}()

		d := net.Dialer{
			Control: func(_, _ string, _ syscall.RawConn) error {
				time.Sleep(time.Second)
				return nil
			},
		}

		conn, err := d.DialContext(ctx, "tcp", "10.0.0.0:http")

		if err != nil {
			t.Log(err)
			return
		}

		conn.Close()
		t.Error("connection did not time out")
	}()

	cancel()
	<-sync

	if ctx.Err() != context.Canceled {
		t.Error("expected canceled context")
	}
}

func TestDialContextCancelFanOut(t *testing.T) {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(10*time.Second))

	listener, err := net.Listen("tcp", "localhost:")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()

	// Accept a single connection and close it immediately.
	go func() {
		conn, err := listener.Accept()
		if err == nil {
			conn.Close()
		}
	}()

	// `dial` is a helper function that dials the listener's address and sends the dialer's ID to the response channel.
	dial := func(ctx context.Context, address string, response chan int, id int, wg *sync.WaitGroup) {
		defer wg.Done()

		var d net.Dialer

		conn, err := d.DialContext(ctx, "tcp", address)
		if err != nil {
			t.Log(err)
			return
		}

		conn.Close()

		// Send the dialer's ID to the response channel if the context has not been canceled.
		select {
		case <-ctx.Done():
		case response <- id:
		}
	}

	// Create a channel to receive the dialer's ID.
	res := make(chan int)

	// Create a WaitGroup to wait for all dialers to complete.
	var wg sync.WaitGroup

	// Create 10 dialers.
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go dial(ctx, listener.Addr().String(), res, i, &wg)
	}

	// Wait for the first dialer to complete.
	response := <-res

	cancel()
	wg.Wait()
	close(res)

	if ctx.Err() != context.Canceled {
		t.Error("expected canceled context; actual:", ctx.Err())
	}

	t.Logf("dialer %d retrieved the resource", response)
}
