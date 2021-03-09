package service

import (
	"bufio"
	"bytes"
	"fmt"
	"game-proxy/kcp/cipher"
	"game-proxy/kcp/util"
	"github.com/xtaci/kcp-go"
	"net"
	"net/http"
	//"runtime/debug"
	"strings"
	"time"
)

var (
	cache       = make(map[string][]byte, 1)
	bufferCache = make([]byte, 0, 3)
)

type Connection struct {
	conn net.Conn
	s    *Server
}

func (c *Connection) handle() {
	defer func() {
		if r := recover(); r != nil {
			////log.LOG.Printf("Recover from handle, %v, Stack::\n%s\n", r, debug.Stack())
		}
		_ = c.conn.Close()
	}()
	if c.s.RemoteAddr == "" {
		c.serverSide()
	} else {
		c.clientSide()
	}
}

func (server *Connection) serverSide() {
	var remote net.Conn
	var err error
	buf := util.GetSmallPool()
	defer util.SPool.Put(buf)
	n, err := server.conn.Read(buf)
	if err != nil {
		//log.LOG.Println("serverSide read first time error ", err)
		return
	}
	// decode
	data, _ := server.s.cipher.Decrypt(buf[:n])
	// parse host
	br := bufio.NewReader(bytes.NewReader(data))
	req, err := http.ReadRequest(br)
	if err != nil {
		//log.LOG.Printf("Http ReadRequest error %v\n", err)
		return
	}
	host := req.URL.Host
	if len(host) > 0 && strings.Index(host, ":") == -1 {
		host += ":80"
	} else if host == "" {
		host = fmt.Sprint(req.Header.Get("Shost"), ":", req.Header.Get("Sport"))
	}
	//log.LOG.Println("try connect real host::" + host)

	//todo cache with nginx
	// dial remote
	remote, err = net.DialTimeout("tcp", host, time.Second*5)
	if err != nil {
		//log.LOG.Printf("dial imeout real remote error %v\n", err)
		return
	}
	defer func() {
		_ = remote.Close()
		_ = server.conn.Close()
	}()
	var established []byte
	switch req.Method {
	case "SOCKS5":
		// response socks5 established
		established = []byte{0x05, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	case "CONNECT":
		established = []byte("HTTP/1.1 200 Connection established\r\n\r\n")
	default:
		// write http pack to real host
		_, err = remote.Write(data)
		if err != nil {
			//log.LOG.Println("Write HTTP header to remote error")
			return
		}
		//log.LOG.Println("HTTP write request.")
		//println("HTTP write request.")
	}
	if len(established) > 0 {
		established, _ := server.s.cipher.Encrypt(established)
		_, err = server.conn.Write(established)
		if err != nil {
			//log.LOG.Println("write established error ", err)
			return
		}
	}
	pipe(server.conn, remote, server.s.cipher, false, true)
}

func (client *Connection) clientSide() {
	var remote net.Conn
	var err error
	if client.s.Config.KCP {
		remote, err = kcp.DialWithOptions(client.s.RemoteAddr, nil, 10, 3)
	} else {
		remote, err = net.DialTimeout("tcp", client.s.RemoteAddr, time.Second*10)
	}
	if err != nil {
		//log.LOG.Printf("net dial failed err %s >> %s\n", err.Error(), client.s.RemoteAddr)
		return
	}
	defer func() {
		_ = remote.Close()
		_ = client.conn.Close()
	}()
	// try handshake socks5
	buf := util.GetSmallPool()
	defer util.SPool.Put(buf)
	ok, data, _ := handshakeSocks(client.conn, buf)
	if ok {
		// socks5 read
		ok, data, err = parseSocks(client.conn, buf)
		if ok {
			// send socks5 to http
			ok = client.writeExBytes(data, remote)
			if !ok {
				return
			}
		} else {
			return
		}
	} else if data != nil {
		////log.LOG.Println(string(data))
		// http read bytes to remote
		num := bytes.IndexByte(data[:], '\r')
		if num == -1 {
			return
		}
		s := string(data[:num])
		var method, version, address string
		fmt.Sscanf(s, "%s%s%s", &method, &address, &version)
		if method != "CONNECT" {
			if strings.Contains(address, "game-a1") {
				//file cache
				remote2, _ := net.DialTimeout("tcp", "121.127.253.117:11431", time.Second*10)
				defer remote2.Close()
				remote2.Write(data)
				fmt.Println(cache)
				pipe(client.conn, remote2, nil, true, false)
				return
			} else {
				ok = client.writeExBytes(data, remote)
				if !ok {
					return
				}
			}
		} else {
			ok = client.writeExBytes(data, remote)
			if !ok {
				return
			}
		}
	} else {
		// socks5 ver check failed
		return
	}
	pipe(client.conn, remote, client.s.cipher, true, true)
}

func pipe(local, remote net.Conn, cp cipher.Cipher, localSide bool, needEnc bool) {
	defer func() {
		_ = local.Close()
		_ = remote.Close()
	}()
	var errChan = make(chan error)
	go func() {
		buf1 := util.GetMiddlePool()
		defer util.MPool.Put(buf1)
		for {
			// copy remote <=> local <=> client
			n, err := remote.Read(buf1)
			if err != nil {
				//log.LOG.Println("remote read error ", err)
				errChan <- err
				break
			}
			// decode
			var pack []byte
			if needEnc {
				if localSide {
					pack, _ = cp.Decrypt(buf1[:n])
				} else {
					pack, _ = cp.Encrypt(buf1[:n])
				}
			} else {
				pack = buf1[:n]
				y := cache["test"]

				if y != nil {

				} else {
					bufferCache = append(bufferCache, pack...)
					cache["test"] = bufferCache
					bufferCache = bufferCache[:0]
				}
			}
		write:
			{
				wn, err := local.Write(pack)
				if err != nil {
					////log.LOG.Println("copy remote to client error ", err)
					errChan <- err
				}
				if wn < len(pack) {
					pack = pack[n:]
					goto write
				}
			}
		}
	}()
	go func() {
		buf2 := util.GetMiddlePool()
		defer util.MPool.Put(buf2)
		for {
			n, err := local.Read(buf2)
			if err != nil {
				//fmt.Println("local read error ", err)
				//log.LOG.Println("local remote addr  ", local.RemoteAddr())
				errChan <- err
				break
			}
			var pack []byte
			// encode to remote
			if needEnc {
				if localSide {
					pack, _ = cp.Encrypt(buf2[:n])
				} else {
					pack, _ = cp.Decrypt(buf2[:n])
				}
			} else {
				pack = buf2[:n]
			}
		write:
			wn, err := remote.Write(pack)
			if err != nil {
				//log.LOG.Println("copy client to remote error ", err)
				errChan <- err
				break
			}
			if wn < len(pack) {
				pack = pack[n:]
				goto write
			}
		}
	}()
	err := <-errChan
	fmt.Println(err)
}

func (c *Connection) writeExBytes(data []byte, remote net.Conn) bool {
	endata, err := c.s.cipher.Encrypt(data)
	if err != nil {
		//log.LOG.Printf("encrypt http data err %v\n", err)
		return false
	}
	_, err = remote.Write(endata)
	if err != nil {
		return false
	}
	return true
}
