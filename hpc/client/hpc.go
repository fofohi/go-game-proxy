/*加密传输的proxy，采用RC4加密，
 */
package main

import (
	"crypto/rc4"
	"flag"
	"fmt"
	"log"
	"net"
	"runtime"
	"time"
)

type Rc4 struct {
	C *rc4.Cipher
}

var pwd = "helloworld"
var ffilename = flag.String("f", "config.json", "配置文件名")

var serverIP string
var localPort string
var serverPort string


func main() {
	localPort = ":9077"
	//serverIP = "43.242.203.152"
	//serverPort = "19077"

	serverIP = "222.186.173.147"
	serverPort = "10107"
	tcpaddr, err := net.ResolveTCPAddr("tcp4", localPort)
	if err != nil {
		fmt.Println("侦听地址错", err)
		return
	}
	tcplisten, err := net.ListenTCP("tcp", tcpaddr)
	if err != nil {
		fmt.Println("开始tcp侦听出错", err)
	}
	fmt.Println("代理客户端已启动，服务端口是:", localPort)
	for {
		client, err := tcplisten.AcceptTCP()
		if err != nil {
			log.Println("当前协程数量：", runtime.NumGoroutine())
			if client != nil {
				client.Close()
			}
			log.Panic(err)
		}

		log.Println("当前协程数量：", runtime.NumGoroutine())
		go handleAClientConn(client)
	}
}

func handleAClientConn(client *net.TCPConn) {

	//defer client.Close()
	c1, _ := rc4.NewCipher([]byte(pwd))
	c2, _ := rc4.NewCipher([]byte(pwd))
	pcTos := &Rc4{c1}
	psToc := &Rc4{c2}

	if client == nil {
		fmt.Println("tcp连接空")
		return
	}

	address := serverIP + ":" + serverPort
	fmt.Println("服务器地址address:", address)
	tcpaddr, err := net.ResolveTCPAddr("tcp4", address)
	if err != nil {
		log.Println("tcp地址错误", address, err)
		return
	}



	server, err := net.DialTCP("tcp", nil, tcpaddr)
	if err != nil {
		log.Println("拨号服务器失败", err)
		return
	}
	//defer server.Close()
	//进行转发,这两句顺序不能倒，否则tcp连接不会自动关掉，会越来越多，只有等系统的tcp,timout到来
	//才能关闭掉。
	go psToc.encryptCopy(client, server) //代理服务端发过来的是密文，编码后就成了明文，并传给浏览器
	go pcTos.encryptCopy(server, client) //客户端收到的是明文，编码后就成了密文并传给代理的服务端
}
func (c *Rc4) encryptCopy(dst *net.TCPConn, src *net.TCPConn) {
	defer dst.Close()
	defer src.Close()
	buf := make([]byte, 4096)
	var err error
	n := 0
	for n, err = src.Read(buf); err == nil && n > 0; n, err = src.Read(buf) {
		//5秒无数据传输就断掉连接
		dst.SetDeadline(time.Now().Add(time.Second * 5))
		src.SetDeadline(time.Now().Add(time.Second * 5))
		c.C.XORKeyStream(buf[:n], buf[:n])

		dst.Write(buf[:n])
	}

}