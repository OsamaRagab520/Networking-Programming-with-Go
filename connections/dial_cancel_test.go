package connections

import (
	"context"
	"net"
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
