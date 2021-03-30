package main

import (
	"crypto/tls"
	"errors"
	"fmt"
	"golang.org/x/net/http2"
	"net"
	"net/http"
)
type tcpKeepAliveListener struct {
	*net.TCPListener
}

type http2ServerConn struct {
	r      *http.Request
	w      http.ResponseWriter
	closed chan struct{}
}


type http2Listener struct {
	server   *http.Server
	connChan chan *http2ServerConn
	addr     net.Addr
	errChan  chan error
}

type Listener interface {
	net.Listener
}

func (l *http2Listener) Accept() (conn net.Conn, err error) {
	select {
	case conn = <-l.connChan:
	case err = <-l.errChan:
		if err == nil {
			err = errors.New("accpet on closed listener")
		}
	}
	return
}

func (l *http2Listener) Addr() net.Addr {
	return l.addr
}

func (l *http2Listener) Close() (err error) {
	select {
	case <-l.errChan:
	default:
		err = l.server.Close()
		l.errChan <- err
		close(l.errChan)
	}
	return nil
}

func (l *http2Listener) handleFunc(w http.ResponseWriter, r *http.Request) {
	conn := &http2ServerConn{
		r:      r,
		w:      w,
		closed: make(chan struct{}),
	}
	select {
	case l.connChan <- conn:
	default:
		fmt.Println("[http2] %s - %s: connection queue is full", r.RemoteAddr, l.server.Addr)
		return
	}

	<-conn.closed
}

func HTTP2Listener(addr string, config *tls.Config) (Listener, error) {
	l := &http2Listener{
		connChan: make(chan *http2ServerConn, 1024),
		errChan:  make(chan error, 1),
	}

	server := &http.Server{
		Addr:      addr,
		Handler:   http.HandlerFunc(l.handleFunc),
		TLSConfig: config,
	}
	if err := http2.ConfigureServer(server, nil); err != nil {
		return nil, err
	}
	l.server = server

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	l.addr = ln.Addr()

	ln = tls.NewListener(tcpKeepAliveListener{ln.(*net.TCPListener)}, config)
	go func() {
		err := server.Serve(ln)
		if err != nil {
			fmt.Println("[http2]", err)
		}
	}()

	return l, nil
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

func main() {

}