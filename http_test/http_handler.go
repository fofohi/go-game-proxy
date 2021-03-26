package main

import (
	"bufio"
	"fmt"
	"github.com/go-log/log"
	"io"
	"net"
	"net/http"
	"time"
)
var timeout = time.Second * 500

var Debug = true
type HandlerOption func(opts *HandlerOptions)

type Handler interface {
	Handle(net.Conn)
}

//
type HttpHandler struct {
	options *HandlerOptions
}

func (h *HttpHandler) Handle(conn net.Conn) {
	defer conn.Close()

	req, err := http.ReadRequest(bufio.NewReader(conn))
	if err != nil {
		fmt.Println("[http] %s - %s : %s", conn.RemoteAddr(), conn.LocalAddr(), err)
		return
	}
	defer req.Body.Close()

	h.handleRequest(conn, req)
}
func (h *HttpHandler) handleRequest(conn net.Conn, req *http.Request) {
	if req == nil {
		return
	}

	host := req.Host
	if _, port, _ := net.SplitHostPort(host); port == "" {
		host = net.JoinHostPort(host, "80")
	}


	req.Header.Del("Gost-Target")

	resp := &http.Response{
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     http.Header{},
	}
	resp.Header.Add("Proxy-Agent", "gost/"+"1")

	if req.Method == "PRI" || (req.Method != http.MethodConnect && req.URL.Scheme != "http") {
		resp.StatusCode = http.StatusBadRequest

		resp.Write(conn)
		return
	}

	req.Header.Del("Proxy-Authorization")

	var err error
	var cc net.Conn

	if req.Method != http.MethodConnect {
		err = h.forwardRequest(conn, req)
		if err == nil {
			return
		}
		log.Logf("[http] %s -> %s : %s", conn.RemoteAddr(), conn.LocalAddr(), err)
	}

	cc, err = net.DialTimeout("tcp",host,timeout)

	if err != nil {
		resp.StatusCode = http.StatusServiceUnavailable
		resp.Write(conn)
		return
	}
	defer cc.Close()
	if req.Method == http.MethodConnect {
		b := []byte("HTTP/1.1 200 Connection established\r\n" +
			"Proxy-Agent: gost/" + "1" + "\r\n\r\n")
		conn.Write(b)
	} else {
		req.Header.Del("Proxy-Connection")

		if err = req.Write(cc); err != nil {
			log.Logf("[http] %s -> %s : %s", conn.RemoteAddr(), conn.LocalAddr(), err)
			return
		}
	}

	log.Logf("[http] %s <-> %s", conn.RemoteAddr(), host)
	transport(conn, cc)
	log.Logf("[http] %s >-< %s", conn.RemoteAddr(), host)
}

func (h *HttpHandler) forwardRequest(conn net.Conn, req *http.Request) error {
	host := req.Host

	cc, err := net.DialTimeout("tcp","localhost:19077",timeout)
	if err != nil {
		return err
	}
	defer cc.Close()

	errc := make(chan error, 1)

	go func() {
		errc <- copyBuffer(conn, cc)
	}()
	go func() {
		for {
			cc.SetWriteDeadline(time.Now().Add(WriteTimeout))
			if !req.URL.IsAbs() {
				req.URL.Scheme = "http" // make sure that the URL is absolute
			}
			err := req.WriteProxy(cc)
			if err != nil {
				log.Logf("[http] %s -> %s : %s", conn.RemoteAddr(), conn.LocalAddr(), err)
				errc <- err
				return
			}
			cc.SetWriteDeadline(time.Time{})

			req, err = http.ReadRequest(bufio.NewReader(conn))
			if err != nil {
				errc <- err
				return
			}
		}
	}()
	log.Logf("[http] %s <-> %s", conn.RemoteAddr(), host)
	<-errc
	log.Logf("[http] %s >-< %s", conn.RemoteAddr(), host)

	return nil
}

//

//handler opt
type HandlerOptions struct{

}

func HTTPHandler(opts ...HandlerOption) Handler {
	return &HttpHandler{}
}

///transport

func transport(rw1, rw2 io.ReadWriter) error {
	errc := make(chan error, 1)
	go func() {
		errc <- copyBuffer(rw1, rw2)
	}()

	go func() {
		errc <- copyBuffer(rw2, rw1)
	}()

	err := <-errc
	if err != nil && err == io.EOF {
		err = nil
	}
	return err
}

func copyBuffer(dst io.Writer, src io.Reader) error {
	buf := lPool.Get().([]byte)
	defer lPool.Put(buf)

	_, err := io.CopyBuffer(dst, src, buf)
	return err
}