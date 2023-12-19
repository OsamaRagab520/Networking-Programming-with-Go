package transmission

import (
	"bufio"
	"net"
	"reflect"
	"testing"
)

const payload = "Go is pretty cool"

func TestScanner(t *testing.T) {
	listener, err := net.Listen("tcp", "localhost:")
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		conn, err := listener.Accept()
		if err != nil {
			t.Error(err)
			return
		}
		defer conn.Close()

		_, err = conn.Write([]byte(payload))
		if err != nil {
			t.Error(err)
		}
	}()

	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	scanner := bufio.NewScanner(conn)
	scanner.Split(bufio.ScanWords)

	var words []string

	for scanner.Scan() {
		words = append(words, scanner.Text())
	}

	err = scanner.Err()
	if err != nil {
		t.Error(err)
	}

	expected := []string{"Go", "is", "pretty", "cool"}

	if !reflect.DeepEqual(words, expected) {
		t.Fatal("incorrect word list")
	}
	t.Logf("Scanned words: %#v", words)
}
