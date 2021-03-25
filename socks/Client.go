package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"sync"
	"time"
)

var(
	lPool = sync.Pool{
		New: func() interface{} {
			return make([]byte, 10 * 1024)
		},
	}
)

func main() {
	server, err := net.Listen("tcp", ":9077")
	if err != nil {
		fmt.Printf("Listen failed: %v\n", err)
		return
	}

	for {
		client, err := server.Accept()
		if err != nil {
			fmt.Printf("Accept failed: %v", err)
			continue
		}
		go process(client)
	}
}



func parseSocks5Request(b []byte) ([]byte, bool) {
	n := len(b)
	resp := b[:0]
	ver := b[0]
	cmd := b[1]
	rsv := b[2]
	atyp := b[3]
	// only support tcp
	resp = append(resp, ver)
	// success
	resp = append(resp, 0x00)
	/*X'00' succeeded
	X'01' general SOCKS server failure
	X'02' connection not allowed by ruleset
	X'03' Network unreachable
	X'04' Host unreachable
	X'05' Connection refused
	X'06' TTL expired
	X'07' Command not supported
	X'08' Address type not supported
	X'09' to X'FF' unassigned*/
	resp = append(resp, rsv)
	resp = append(resp, atyp)
	if cmd == 1 {
		var host, port string
		switch b[3] {
		case 0x01: //IP V4
			host = net.IPv4(b[4], b[5], b[6], b[7]).String()
		case 0x03: //Domain
			host = string(b[5 : n-2]) //b[4] domain length
		case 0x04: //IP V6
			host = net.IP{b[4], b[5], b[6], b[7], b[8], b[9], b[10], b[11], b[12], b[13], b[14], b[15], b[16], b[17], b[18], b[19]}.String()
		}
		port = strconv.Itoa(int(b[n-2])<<8 | int(b[n-1]))
		//LOG.Printf("type %d， target host %s port %s\n", atyp, string(host), port)
		// socks to http, send to remote
		to5 := to5Connect(host, port)
		return to5, true
	} else {
		// failed
		resp[1] = 0x01
		resp = append(resp, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00}...)
	}
	return resp, false
}

func to5Connect(host, port string) []byte {
	var bf bytes.Buffer
	bf.WriteString("SOCKS5 /socks5 HTTP/1.1\r\n")
	bf.WriteString(fmt.Sprintf("Shost: %s\r\n", host))
	bf.WriteString(fmt.Sprintf("Sport: %s\r\n", port))
	bf.WriteString("\r\n")
	return bf.Bytes()
}

func handshakeSocks(c net.Conn, buf []byte) (bool, []byte, error) {
	handshake := func(pkg []byte, conn net.Conn) bool {
		ver := pkg[0]
		if ver != 0x05 {
			//log.LOG.Printf("unsupport socks version %d \n", ver)
			return false
		}
		resp := pkg[:0]
		resp = append(resp, 0x05)
		resp = append(resp, 0x00)
		n, err := conn.Write(resp)
		if n != 2 || err != nil {
			return false
		}
		// handshake over
		return true
	}
	n, er := io.ReadAtLeast(c, buf, 3)
	if er != nil {
		return false, nil, er
	}
	// socks5
	if buf[0] == 0x05 {
		ok := handshake(buf[:n], c)
		return ok, nil, nil
	}
	// http, buf is left byte
	return false, buf[:n], nil
}

func parseSocks(conn net.Conn, buf []byte) (bool, []byte, error) {
	conn.SetReadDeadline(time.Now().Add(time.Second * 2))
	n, er := io.ReadAtLeast(conn, buf, 6)
	conn.SetReadDeadline(time.Time{})
	if er != nil {
		return false, nil, er
	}
	// socks5
	if buf[0] == 0x05 {
		to5, ok := parseSocks5Request(buf[:n])
		if !ok {
			conn.Write(to5)
			return false, nil, nil
		}
		return true, to5, nil
	}
	// http, buf is left byte
	return false, buf[:n], nil
}

func process(client net.Conn) {
	buf := lPool.Get().([]byte)
	defer lPool.Put(buf)
	ok, data, err := handshakeSocks(client, buf)
	if err != nil {
		//log.LOG.Println("serverSide read first time error ", err)
		return
	}
	if ok{
		ok,data, err = parseSocks(client, buf)
		fmt.Println(data)
	}

	/*if err := Socks5Auth(client); err != nil {
		fmt.Println("auth error:", err)
		client.Close()
		return
	}

	target, err := Socks5Connect(client)
	if err != nil {
		fmt.Println("connect error:", err)
		client.Close()
		return
	}*/
	x,err := net.Dial("tcp", "160.116.118.18:19077")
	//x,err := net.Dial("tcp", "162.14.8.228:19077")
	if err != nil{
		log.Fatal(err)
		return
	}
	defer x.Close()
	defer client.Close()
	//n,errs := x.Read(y[:])
	/*if errs != nil {
		log.Fatal(errs)
	}*/


	transport(client, x)
}


func Socks5Auth(client net.Conn) (err error) {
	buf := make([]byte, 256)

	// 读取 VER 和 NMETHODS
	n, err := io.ReadFull(client, buf[:2])
	if n != 2 {
		return errors.New("reading header: " + err.Error())
	}

	ver, nMethods := int(buf[0]), int(buf[1])
	if ver != 5 {
		return errors.New("invalid version")
	}

	// 读取 METHODS 列表
	n, err = io.ReadFull(client, buf[:nMethods])
	if n != nMethods {
		return errors.New("reading methods: " + err.Error())
	}
	//无需认证
	n, err = client.Write([]byte{0x05, 0x00})
	if n != 2 || err != nil {
		return errors.New("write rsp err: " + err.Error())
	}

	return nil
}


func Socks5Connect(client net.Conn) (net.Conn, error) {
	buf := make([]byte, 256)

	n, err := io.ReadFull(client, buf[:4])
	if n != 4 {
		return nil, errors.New("read header: " + err.Error())
	}

	ver, cmd, _, atyp := buf[0], buf[1], buf[2], buf[3]
	if ver != 5 || cmd != 1 {
		return nil, errors.New("invalid ver/cmd")
	}

	addr := ""
	switch atyp {
	case 1:
		n, err = io.ReadFull(client, buf[:4])
		if n != 4 {
			return nil, errors.New("invalid IPv4: " + err.Error())
		}
		addr = fmt.Sprintf("%d.%d.%d.%d", buf[0], buf[1], buf[2], buf[3])

	case 3:
		n, err = io.ReadFull(client, buf[:1])
		if n != 1 {
			return nil, errors.New("invalid hostname: " + err.Error())
		}
		addrLen := int(buf[0])

		n, err = io.ReadFull(client, buf[:addrLen])
		if n != addrLen {
			return nil, errors.New("invalid hostname: " + err.Error())
		}
		addr = string(buf[:addrLen])

	case 4:
		return nil, errors.New("IPv6: no supported yet")

	default:
		return nil, errors.New("invalid atyp")
	}
	n, err = io.ReadFull(client, buf[:2])
	if n != 2 {
		return nil, errors.New("read port: " + err.Error())
	}
	port := binary.BigEndian.Uint16(buf[:2])
	destAddrPort := fmt.Sprintf("%s:%d", addr, port)
	dest, err := net.Dial("tcp", destAddrPort)
	if err != nil {
		return nil, errors.New("dial dst: " + err.Error())
	}
	n, err = client.Write([]byte{0x05, 0x00, 0x00, 0x01, 0, 0, 0, 0, 0, 0})
	if err != nil {
		dest.Close()
		return nil, errors.New("write rsp: " + err.Error())
	}
	return dest, nil
}

func transport(rw1, rw2 io.ReadWriter) error {
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

func copyBuffer(dst io.Writer, src io.Reader) error {
	buf := lPool.Get().([]byte)
	defer lPool.Put(buf)

	_, err := io.CopyBuffer(dst, src, buf)

	return err
}

func Socks5Forward(client, target net.Conn) {
	forward := func(src, dest net.Conn) {
		defer src.Close()
		defer dest.Close()
		io.Copy(src, dest)
	}
	go forward(client, target)
	go forward(target, client)
}