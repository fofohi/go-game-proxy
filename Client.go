package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"net/url"
	"strings"
	"sync"
)

const (
	flag = "443\r\n"
	flagN = "80\r\n"
	sslPort = "443"
)

const (
	twitter = "twitter"
	mbga = "mbga"
)


var (
	smallBufferSize  = 4 * 1024  // 2KB small buffer
	mediumBufferSize = 8 * 1024  // 8KB medium buffer
	largeBufferSize  = 64 * 1024 // 32KB large buffer
)
var (
	sPool = sync.Pool{
		New: func() interface{} {
			return make([]byte, smallBufferSize)
		},
	}
	mPool = sync.Pool{
		New: func() interface{} {
			return make([]byte, mediumBufferSize)
		},
	}
	lPool = sync.Pool{
		New: func() interface{} {
			return make([]byte, largeBufferSize)
		},
	}
)

func handleClientRequest3(client net.Conn) {
	if client == nil {
		return
	}
	defer client.Close()

	b := sPool.Get().([]byte)
	defer sPool.Put(b)
	n, err := client.Read(b[:])

	var method, version, address string

	num := bytes.IndexByte(b[:], '\r')
	if num == -1 {
		return
	}
	s := string(b[:num])
	fmt.Sscanf(s, "%s%s%s", &method,&address,&version)
	hostPortURL, err := url.Parse(address)
	if err != nil {
		log.Println(err)
		return
	}

	if hostPortURL.Opaque == sslPort {
		address = hostPortURL.Scheme + sslPort
		if strings.Contains(address,"google") || strings.Contains(address,"twitter") || strings.Contains(address,"dmm") {
			server, err := net.Dial("tcp", "121.127.253.117:9079")
			if err != nil {
				log.Println(err)
				return
			}
			b2 := saltByte(b)
			server.Write([]byte(flag))
			server.Write(b2[:n])
			transport(server, client)
		}else{
			server, err := net.Dial("tcp", hostPortURL.Scheme + ":" + hostPortURL.Opaque)
			if err != nil {
				log.Println(err)
				return
			}
			if method == "CONNECT" {
				fmt.Fprint(client, "HTTP/1.1 200 Connection established\r\n\r\n")
			}
			transport(server, client)
		}
	} else {
		if strings.Index(hostPortURL.Host, ":") == -1 {
			address = hostPortURL.Host + ":80"
		} else {
			address = hostPortURL.Host
		}
		server, err := net.Dial("tcp", "localhost:9079")
		if err != nil {
			log.Println(err)
			return
		}
		b2 := saltByte(b)
		server.Write([]byte(flagN))
		server.Write(b2[:n])
		//进行转发
		transport(server, client)
	}
}

func saltByte(b []byte) []byte {
	b2 := sPool.Get().([]byte)
	defer sPool.Put(b2)
	for k, v := range b {
		if v == 0 {
			b2[k] = 0
		}
		b2[k] = v ^ 2
	}
	return b2
}

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

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	l, err := net.Listen("tcp", ":9078")
	if err != nil {
		log.Panic(err)
	}
	for {
		client, err := l.Accept()
		if err != nil {
			log.Panic(err)
		}

		go handleClientRequest3(client)
	}
}
