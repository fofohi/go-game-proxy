package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/http"
)

func main() {

	ln,err := TCPListener(":19077")
	if err != nil{
		return
	}
	for  {
		conn,e := ln.Accept()
		if e != nil {
			return
		}
		go Handle(conn)
	}

}

func Handle(conn net.Conn) {
	defer conn.Close()
	b := make([]byte,10)
	r := bufio.NewReader(conn)
	for {
		_,errs := r.Read(b)
		if errs == io.EOF{
			break
		}
		fmt.Print(string(b))
		b = b[:0]
	}

	req, err := http.ReadRequest(bufio.NewReader(conn))
	if err != nil {
		return
	}
	defer req.Body.Close()

	fmt.Print(req)

}

type Listener interface {
	net.Listener
}

type tcpKeepAliveListener struct {
	*net.TCPListener
}

type tcpListener struct {
	net.Listener
}


func (ln tcpKeepAliveListener) Accept() (c net.Conn, err error) {
	tc, err := ln.AcceptTCP()
	if err != nil {
		return
	}
	tc.SetKeepAlive(true)
	tc.SetKeepAlivePeriod(1000)
	return tc, nil
}

func TCPListener(addr string) (Listener, error) {
	laddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return nil, err
	}
	ln, err := net.ListenTCP("tcp", laddr)
	if err != nil {
		return nil, err
	}
	return &tcpListener{Listener: tcpKeepAliveListener{ln}}, nil
}