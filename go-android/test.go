package go_android

import (
	"golang.org/x/sys/unix"
	"log"
	"net"
	"os"
	"sync"
)
var (
	sPool = sync.Pool{
		New: func() interface{} {
			return make([]byte, 1024)
		},
	}
	mPool = sync.Pool{
		New: func() interface{} {
			return make([]byte, 2048)
		},
	}
)
type VpnService interface {
	Protect(fd int) bool
}

func ProtectFd(v VpnService,fd int)  bool{
	return v.Protect(fd)
}


func IpDataGet(b []byte,v VpnService)  {
	log.Println("begin get byte")
	log.Println("======================")
	log.Println("open unix socket")
	fd,err := unix.Socket(unix.AF_INET6,unix.SOCK_STREAM, unix.IPPROTO_TCP)
	log.Println(fd)
	if err != nil{
		log.Fatal(err)
	}
	log.Println("protect fd")
	v.Protect(fd)

	address := "121.127.253.117"

	ipStore := [16]byte{}
	for x,y := range address {
		ipStore[x] = uint8(y)
	}
	sa := &unix.SockaddrInet6{
		Addr: ipStore,
		Port: 11431,
	}
	log.Println("connect fd")
	unixErr := unix.Connect(fd,sa)
	if unixErr != nil{
		log.Fatal(unixErr)
	}
	log.Println("new file")
	file := os.NewFile(uintptr(fd), "Socket")
	if file == nil {
		log.Fatal(nil)
	}
	log.Println("connect file")
	conn, err := net.FileConn(file)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("begin write to pc server")
	n,_ := conn.Write(b)
	log.Println(n)
}
