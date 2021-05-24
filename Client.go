package main

import (
	"fmt"
	"github.com/hashicorp/yamux"
	"io"
	"log"
	"net"
	"sync"
	"time"
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
func encryptCopy1(dst *net.TCPConn, src *net.TCPConn) {
	defer dst.Close()
	defer src.Close()
	buf := make([]byte, 4096)
	var err error
	n := 0
	for n, err = src.Read(buf); err == nil && n > 0; n, err = src.Read(buf) {
		//5秒无数据传输就断掉连接
		dst.SetDeadline(time.Now().Add(time.Second * 5))
		src.SetDeadline(time.Now().Add(time.Second * 5))
		dst.Write(buf[:n])
	}

}

func encryptCopy2(dst *net.TCPConn, src *net.TCPConn) {
	defer dst.Close()
	defer src.Close()
	buf := make([]byte, 4096)
	var err error
	n := 0
	for n, err = src.Read(buf); err == nil && n > 0; n, err = src.Read(buf) {
		//5秒无数据传输就断掉连接
		dst.SetDeadline(time.Now().Add(time.Second * 5))
		src.SetDeadline(time.Now().Add(time.Second * 5))
		dst.Write(buf[:n])
	}

}

func encryptCopy3(dst io.ReadWriter, src io.ReadWriter) {
	buf := make([]byte, 4096)
	var err error
	n := 0
	for n, err = src.Read(buf); err == nil && n > 0; n, err = src.Read(buf) {
		//5秒无数据传输就断掉连接
		dst.Write(buf[:n])
	}

}

func handleClientRequest3(client *net.TCPConn) {
	tcpaddr, err := net.ResolveTCPAddr("tcp4", "localhost:19077")
	if err != nil {
		log.Println("tcp地址错误", "address", err)
		return
	}
	server, err := net.DialTCP("tcp", nil, tcpaddr)
	session, err := yamux.Client(server, nil)

	stream, err := session.Open()
	if err != nil {
		log.Println(err)
		return
	}
	//b2 := saltByte(b)
	//server.Write([]byte(flagN))
	//go encryptCopy1(client, server) //代理服务端发过来的是密文，编码后就成了明文，并传给浏览器
	//go encryptCopy2(server, client) //客户端收到的是明文，编码后就成了密文并传给代理的服务端
	go encryptCopy3(client, stream)
	go encryptCopy3(stream, client)
	/*if client == nil {
		return
	}
	defer client.Close()

	b := getPoolSmall()
	n, err := client.Read(b[:])

	var method, version, address string

	num := bytes.IndexByte(b[:], '\r')
	if num == -1 {
		return
	}
	s := string(b[:num])
	fmt.Sscanf(s, "%s%s%s", &method,&address,&version)
	if(!strings.HasPrefix(address,"http://")){
		address = "http://" + address
	}
	hostPortURL, err := url.Parse(address)
	if err != nil {
		log.Println(err)
		return
	}

	if hostPortURL.Opaque == sslPort {
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
		tcpaddr, err := net.ResolveTCPAddr("tcp4", "localhost:19077")
		if err != nil {
			log.Println("tcp地址错误", address, err)
			return
		}
		server, err := net.DialTCP("tcp", nil,tcpaddr)
		if err != nil {
			log.Println(err)
			return
		}
		//b2 := saltByte(b)
		//server.Write([]byte(flagN))
		go encryptCopy(client, server) //代理服务端发过来的是密文，编码后就成了明文，并传给浏览器
		go encryptCopy(server, client) //客户端收到的是明文，编码后就成了密文并传给代理的服务端
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
	tcpaddr, err := net.ResolveTCPAddr("tcp4", ":9077")
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
		go handleClientRequest3(client)
	}
}
