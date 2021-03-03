package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof"
	"net/url"
	"strings"
	"sync"
	"time"
)

var (
	tinyBufferSize   = 512
	smallBufferSize  = 4 * 1024  // 2KB small buffer
	mediumBufferSize = 8 * 1024  // 8KB medium buffer
	largeBufferSize  = 32 * 1024 // 32KB large buffer
)
var (
	mapPool = make(map[string]bytes.Buffer)

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

func GetAddress(u *url.URL) string {
	host := u.Hostname()
	port := u.Port()
	if port == "" {
		if u.Scheme == "https" {
			port = "443"
		} else {
			port = "80"
		}
	}
	return fmt.Sprintf("%s:%s", host, port)
}

func main() {
	/*go func() {
		http.ListenAndServe("0.0.0.0:18080", nil)
	}()*/

	log.SetFlags(log.LstdFlags | log.Lshortfile)
	l, err := net.Listen("tcp", ":9079")
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

func handleClientRequest3(client net.Conn) {
	fmt.Println("come in")
	if client == nil {
		return
	}
	defer client.Close()
	b := sPool.Get().([]byte)
	defer sPool.Put(b)

	_, err := client.Read(b)

	if err != nil {
		return
	}
	var method, host, address string

	num := bytes.IndexByte(b[:], '\n')
	is443 := true
	if num == -1 && len(b[:]) > 0 {
		is443 = false
	} else {
		is443 = false
	}

	/*hostAndPort := strings.Split(req.URL.Host,":")
	var oqa string
	if len(hostAndPort) > 1{
		oqa = hostAndPort[1]
	}else{
		oqa = ""
	}*/
	if is443 {
		s := string(b[:num])
		fmt.Sscanf(s, "%s %s %s", &method, &address, &host)
		server, err := net.Dial("tcp4", address)
		if err != nil {
			return
		}
		defer server.Close()
		if method == "CONNECT" {
			fmt.Fprint(client, "HTTP/1.1 200 Connection established\r\n\r\n")
		} else {
			fmt.Println("not connect")
		}
		Transport(server, client)
	} else {
		b2 := sPool.Get().([]byte)
		defer sPool.Put(b2)
		for k, v := range b {
			if v == 0 {
				continue
			}
			b2[k] = v ^ 2
		}

		bnr := bufio.NewReader(bytes.NewReader(b2))

		req, err := http.ReadRequest(bnr)

		if err != nil {
			return
		}
		server, err := net.Dial("tcp4", GetAddress(req.URL))
		if err != nil {
			return
		}
		begin := time.Now().Unix()
		if strings.HasPrefix(req.URL.Host, "game-a") {
			_ = GetResponse(server, client, req)
		} else {
			defer server.Close()
			req.Write(server)
			Transport(server, client)
		}
		fmt.Println("COST TIME ===>", time.Now().Unix()-begin)
	}
}

func GetResponse(server net.Conn, client net.Conn, req *http.Request) *http.Response {
	defer server.Close()
	req.Write(server)
	resp, _ := http.ReadResponse(bufio.NewReader(server), req)
	resp.Write(client)
	return resp
}

func Transport(rw1, rw2 io.ReadWriter) error {
	error0 := make(chan error, 1)
	go func() {
		error0 <- CopyBuffer(rw1, rw2)
	}()

	go func() {
		error0 <- CopyBuffer(rw2, rw1)
	}()

	err := <-error0
	if err != nil && err == io.EOF {
		err = nil
	}
	return err
}

func CopyBuffer(dst io.Writer, src io.Reader) error {
	buf := lPool.Get().([]byte)
	defer lPool.Put(buf)

	_, err := io.CopyBuffer(dst, src, buf)

	return err
}
