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
	"os"
	"sync"
)


func getPoolBig() []byte  {
	b := lPool.Get().([]byte)
	defer lPool.Put(b)
	return b
}

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

func main3() {
	/*go func() {
		http.ListenAndServe("0.0.0.0:18080", nil)
	}()*/
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	port := ":19077"
	if len(os.Args) > 1 {
		port = os.Args[1]
	}

	l, err := net.Listen("tcp", port)
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
	s := string(b[:3])
	is443 := s == "443"

	if is443 {
		b2 := sPool.Get().([]byte)
		defer sPool.Put(b2)
		b2 = deSaltByte(b)

		s := string(b2[3:])
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
		//transport(server, client)
		transport(client,server)
	} else {
    	//b2 := deSaltByte(b[3:])
		bnr := bufio.NewReader(bytes.NewReader(/*b2[:]*/b))
		req, err := http.ReadRequest(bnr)
		if err != nil {
			return
		}
		server, err := net.Dial("tcp4", GetAddress(req.URL))
		if err != nil {
			return
		}


		/*if strings.HasPrefix(req.URL.Host, "game-a") {
			_ = GetResponse(server, client, req)
		} else {*/
		defer server.Close()
		GetResponse(server, client, req)
		//fmt.Println(x)
		//req.Write(server)
		//transport(server, client)
		//transport(client,server)
		//}
	}
}

func deSaltByte(b []byte) []byte {
	b2 := sPool.Get().([]byte)
	defer sPool.Put(b2)
	for k, v := range b {
		if v == 0 {
			continue
		}
		b2[k] = v ^ 2
	}
	return b2
}

func GetResponse(server net.Conn, client net.Conn, req *http.Request) *http.Response {
	defer server.Close()
	req.Write(server)
	resp, _ := http.ReadResponse(bufio.NewReader(server), req)
	if resp == nil{
		return nil
	}
	resp.Write(client)
	return resp
}

func GetResponse2(server net.Conn, client net.Conn, req *http.Request) {
	lock := &sync.RWMutex{}
	lock.Lock()
	req.Write(server)
	resp, _ := http.ReadResponse(bufio.NewReader(server), req)
	if resp == nil{
		return
	}
	lock.Unlock()
	resp.Write(client)

}

func transport333(rw1, rw2 io.ReadWriter) error {
	error0 := make(chan error, 1)
	go func() {
		error0 <- copyBuffer(rw1, rw2)
	}()

	go func() {
		error0 <- copyBuffer(rw2, rw1)
	}()

	err := <-error0
	if err != nil && err == io.EOF {
		err = nil
	}
	return err
}

func copyBuffer3333(dst io.Writer, src io.Reader) error {
	buf := lPool.Get().([]byte)
	defer lPool.Put(buf)

	_, err := io.CopyBuffer(dst, src, buf)

	return err
}
