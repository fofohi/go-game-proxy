package main

import (
	"crypto/rc4"
	"fmt"
	"github.com/xtaci/smux"
	"io"
	"log"
	"net"
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
var pwd = "helloworld"

type Rc4 struct {
	C *rc4.Cipher
}
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


func encryptCopy3(dst io.ReadWriter, src io.ReadWriter) {
	buf := getPoolSmall()
	var err error
	n := 0
	for n, err = src.Read(buf); err == nil && n > 0; n, err = src.Read(buf) {
		buf = saltByte(buf)
		dst.Write(buf[:n])
	}

}

func handleClientRequest3(client *net.TCPConn){

	tcpaddr, err := net.ResolveTCPAddr("tcp4", "localhost:19078")
	if err != nil {
		log.Println("tcp地址错误", "address", err)
		return
	}
	server, err := net.DialTCP("tcp", nil, tcpaddr)
	session, err := smux.Client(server, nil)

	stream, err := session.Open()
	if err != nil {
		log.Println(err)
		return
	}
	//go encryptCopy1(client, server) //代理服务端发过来的是密文，编码后就成了明文，并传给浏览器
	//go encryptCopy2(server, client) //客户端收到的是明文，编码后就成了密文并传给代理的服务端
	go encryptCopy3(client, stream)
	go encryptCopy3(stream, client)


}

func saltByte(b []byte) []byte {
	for k, v := range b {
		if v == 0 {
			b[k] = 0
		}
		b[k] = v ^ 2
	}
	return b
}


func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	tcpaddr, err := net.ResolveTCPAddr("tcp4", ":9078")
	if err != nil {
		fmt.Println("侦听地址错", err)
		return
	}
	l, err := net.ListenTCP("tcp", tcpaddr)
	if err != nil {
		log.Panic(err)
	}
	for {
		client, err := l.AcceptTCP()
		if err != nil {
			log.Panic(err)
		}
		handleClientRequest3(client)
	}
}
