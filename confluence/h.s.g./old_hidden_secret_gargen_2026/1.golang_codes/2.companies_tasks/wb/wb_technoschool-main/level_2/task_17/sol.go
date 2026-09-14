package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"time"
)

func main() {
	timeout := flag.Duration("timeout", 10*time.Second, "connection timeout")
	flag.Parse()
	if flag.NArg() < 2 {
		fmt.Fprintln(os.Stderr, "usage: go run sol.go [--timeout=10s] <host> <port>")
		os.Exit(2)
	}
	host, port := flag.Arg(0), flag.Arg(1)
	addr := net.JoinHostPort(host, port)

	d := net.Dialer{Timeout: *timeout}
	conn, err := d.Dial("tcp", addr)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer conn.Close()

	done := make(chan struct{})

	// socket -> stdout
	go func() {
		_, _ = io.Copy(os.Stdout, conn)
		close(done) // server closed or read error
	}()

	// stdin -> socket
	go func() {
		_, _ = io.Copy(conn, os.Stdin)
		_ = conn.(*net.TCPConn).CloseWrite()
	}()

	<-done
}
