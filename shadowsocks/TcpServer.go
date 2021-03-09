package main

import (
	"fmt"
	"net"
)

func main() {
	l, err := net.Listen("tcp", ":9078")
	if err != nil {
		return
	}

	for {
		c, err := l.Accept()
		if err != nil {
			return
		}
		go func() {
			defer c.Close()
			addr, err := Handshake(c)
			if err != nil {
				return
			}
			fmt.Println(addr)
		}()
	}

}
