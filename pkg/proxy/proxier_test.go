package proxy

import (
	"fmt"
	"io"
	"net"
	"testing"

	"k8s-firstcommit/pkg/api"
)

// a simple echoServer that only accept one connection
func echoServer(addr string) error {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to start echo service: %v", err)
	}
	defer l.Close()
	conn, err := l.Accept()
	if err != nil {
		return fmt.Errorf("failed to accept new conn to echo service: %v", err)
	}
	io.Copy(conn, conn)
	conn.Close()
	return nil
}

func TestProxy(t *testing.T) {
	go func() {
		if err := echoServer("127.0.0.1:2222"); err != nil {
			t.Fatal(err)
		}
	}()

	lb := NewLoadBalancerRR()
	lb.OnUpdate([]api.Endpoints{{"echo", []string{"127.0.0.1:2222"}}})

	p := NewProxier(lb)
	if err := p.AddService("echo", 2223); err != nil {
		t.Fatalf("error adding new service: %v", err)
	}
	conn, err := net.Dial("tcp", "127.0.0.1:2223")
	if err != nil {
		t.Fatalf("error connecting to proxy: %v", err)
	}
	magic := "aaaaa"
	if _, err := conn.Write([]byte(magic)); err != nil {
		t.Fatalf("error writing to proxy: %v", err)
	}
	buf := make([]byte, 5)
	if _, err := conn.Read(buf); err != nil {
		t.Fatalf("error reading from proxy: %v", err)
	}
	if string(buf) != magic {
		t.Fatalf("bad echo from proxy: got: %q, expected %q", string(buf), magic)
	}
}
