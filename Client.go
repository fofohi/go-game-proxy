package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"strings"

	/*"net/http"
	"strings"*/
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
func parseRequestLine(line string) (method, requestURI, proto string, ok bool) {
	s1 := strings.Index(line, " ")
	s2 := strings.Index(line[s1+1:], " ")
	if s1 < 0 || s2 < 0 {
		return
	}
	s2 += s1 + 1
	return line[:s1], line[s1+1 : s2], line[s2+1:], true
}

func handleClientRequest3(client net.Conn) {
	if client == nil {
		return
	}
	defer client.Close()
	r := bufio.NewReader(client)
	_, is, e := r.ReadLine()

	if e != nil {
		fmt.Print(is)
		return
	}

	l2, _, _ := r.ReadLine()

	l2s := string(l2)
	l2sa := strings.Split(l2s, " ")

	if !strings.Contains(l2sa[1], "granbluefantasy") {
		return
	}

	tcpAddr, err := net.ResolveTCPAddr("tcp4", l2sa[1]+":80")
	if err != nil {
		log.Println("tcp地址错误", l2sa[1], err)
		return
	}
	server, err := net.DialTCP("tcp", nil, tcpAddr)

	if err != nil {
		fmt.Println(err)
		return
	}

	c := getPoolSmall()
	go func() {
		for n, erra := r.Read(c); erra == nil && n > 0; n, erra = r.Read(c) {
			fmt.Println(string(c[:n]))
			server.Write(c[:n])
		}
	}()

	transport(server, client)
	/*bfr := bufio.NewReader(strings.NewReader(string(b)))
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




	//transport(server, client)


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
