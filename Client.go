package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
)

const (
	flag    = "443"
	flagN   = "80$"
	sslPort = "443"
)

const (
	twitter = "twitter"
	mbga    = "mbga"
)

var (
	smallBufferSize  = 4 * 1024  // 2KB small buffer
	mediumBufferSize = 8 * 1024  // 8KB medium buffer
	largeBufferSize  = 64 * 1024 // 32KB large buffer
	queue1           = make(chan net.Conn, 1)
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

func getPoolSmall() []byte {
	b := sPool.Get().([]byte)
	defer sPool.Put(b)
	return b
}

func getPoolBig() []byte {
	b := lPool.Get().([]byte)
	defer lPool.Put(b)
	return b
}

func handleClientRequest3(client net.Conn) {
	if client == nil {
		return
	}
	defer client.Close()

	b := make([]byte, 4096)
	r := bufio.NewReader(client)

	n, errors := r.Read(b)

	if errors != nil {
		fmt.Println(errors)
		fmt.Println(n)
	}
	bfr := bufio.NewReader(strings.NewReader(string(b)))
	req, err := http.ReadRequest(bfr)

	if err != nil {
		fmt.Println(err)
	}
	var address string
	if req.Method == "CONNECT" {
		client.Write([]byte("HTTP/1.1 200 Connection established\r\n\r\n"))
	}
	hostPort := strings.Split(req.Host, ":")
	if len(hostPort) < 2 {
		address = hostPort[0] + ":80"
	} else {
		address = req.Host
	}

	tcpAddr, err := net.ResolveTCPAddr("tcp4", address)
	if err != nil {
		log.Println("tcp地址错误", address, err)
		return
	}
	server, err := net.DialTCP("tcp", nil, tcpAddr)

	go server.Write(b)

	transport(server, client)

	/*if hostPortURL.Opaque == sslPort {
		address = hostPortURL.Scheme + sslPort
		if strings.Contains(address,"twitter") || strings.Contains(address,"twimg") {
			server, err := net.Dial("tcp", "162.14.8.228:19077")
			if err != nil {
				log.Println(err)
				return
			}

			b2 := saltByte(b)
			server.Write([]byte(flag))
			server.Write(b2[:n])
			//transport(server, client)
			transport(client, server)
		}else{
			server, err := net.Dial("tcp", hostPortURL.Scheme + ":" + hostPortURL.Opaque)
			if err != nil {
				log.Println(err)
				return
			}
			if method == "CONNECT" {
				fmt.Fprint(client, "HTTP/1.1 200 Connection established\r\n\r\n")
			}
			transport(client, server)
		}
	} else {
		if strings.Index(hostPortURL.Host, ":") == -1 {
			address = hostPortURL.Host + ":80"
		} else {
			address = hostPortURL.Host
		}
		server, err := net.Dial("tcp", "localhost:19077")
		if err != nil {
			log.Println(err)
			return
		}
		//b2 := saltByte(b)
		//server.Write([]byte(flagN))
		server.Write(b2[:n])
		//进行转发
		transport(server, client)
	}*/
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
	l, err := net.Listen("tcp", ":9077")
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
