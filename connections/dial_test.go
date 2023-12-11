package connections

import (
	"io"
	"net"
	"testing"
)

func TestDial(t *testing.T) {
	listener, err := net.Listen("tcp", "localhost:8080") // Create a TCP listener on localhost:8080

	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()

	t.Logf("bound to %q", listener.Addr()) // Log the address the listener is bound to

	done := make(chan struct{}) // Create a channel to signal completion
	go func() {
		defer func() {
			done <- struct{}{} // Signal completion by sending a struct{}{} to the channel
		}()

		for {
			conn, err := listener.Accept() // Accept incoming connections
			if err != nil {
				t.Log(err)
				return
			}

			go func(c net.Conn) {
				defer func() {
					c.Close()
					done <- struct{}{} // Signal completion by sending a struct{}{} to the channel
				}()

				buf := make([]byte, 1024)
				for {
					n, err := c.Read(buf) // Read data from the connection
					if err != nil {
						if err != io.EOF {
							t.Error(err)
						}
						return
					}

					t.Logf("received: %q", buf[:n]) // Log the received data
				}
			}(conn)
		}
	}()

	conn, err := net.Dial("tcp", listener.Addr().String()) // Dial the listener's address
	if err != nil {
		t.Fatal(err)
	}
	conn.Close()
	<-done // Wait for the first completion signal
	listener.Close()
	<-done // Wait for the second completion signal
}
