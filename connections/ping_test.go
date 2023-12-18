package connections

import (
	"context"
	"net"
	"testing"
	"time"
)

func TestPingerAdvancedDeadline(t *testing.T) {
	done := make(chan struct{})

	listener, err := net.Listen("tcp", "localhost:")
	if err != nil {
		t.Fatal(err)
	}

	begin := time.Now()
	go func() {
		defer close(done)

		conn, err := listener.Accept()
		if err != nil {
			t.Log(err)
			return
		}
		ctx, cancel := context.WithCancel(context.Background())
		defer func() {
			cancel()
			conn.Close()
		}()
		resetTimer := make(chan time.Duration, 1)
		resetTimer <- 1 * time.Second

		go Pinger(ctx, conn, resetTimer)

		err = conn.SetDeadline(time.Now().Add(5 * time.Second))
		if err != nil {
			t.Error(err)
			return
		}

		buf := make([]byte, 1024)
		for {
			n, err := conn.Read(buf)
			if err != nil {
				return
			}
			t.Logf("[%s] %s", time.Since(begin).Truncate(time.Second), buf[:n])
			resetTimer <- 0
			err = conn.SetDeadline(time.Now().Add(5 * time.Second))
			t.Log("Advancing the deadline")
			if err != nil {
				t.Error(err)
				return
			}

		}
	}()

	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	buf := make([]byte, 1024)
	for i := 0; i < 4; i++ {
		n, err := conn.Read(buf)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("[%s] %s", time.Since(begin).Truncate(time.Second), buf[:n])
	}
	_, err = conn.Write([]byte("PONG!!!"))
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 4; i++ {
		n, err := conn.Read(buf)
		if err != nil {
			if err != nil {
				t.Fatal(err)
			}
			break
		}
		t.Logf("[%s] %s", time.Since(begin).Truncate(time.Second), buf[:n])
	}
	<-done
	end := time.Since(begin).Truncate(time.Second)
	t.Logf("[%s] done", end)
	if end != 9*time.Second {
		t.Fatalf("expected EOF at 9 seconds; actual %s", end)
	}
}
