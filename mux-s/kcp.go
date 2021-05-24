package main

import (
	"crypto/md5"
	"crypto/rand"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net"

	"github.com/xtaci/kcp-go"
	"golang.org/x/crypto/chacha20"
)

type config struct {
	ListenAddr string
	RemoteAddr string
	Key        []byte
}

var GlobalConfig = &config{}

type CipherStream interface {
	Read(p []byte) (int, error)
	Write(p []byte) (int, error)
	Close() error
}

type Chacha20Stream struct {
	key     []byte
	encoder *chacha20.Cipher
	decoder *chacha20.Cipher
	conn    net.Conn
}

func NewChacha20Stream(key []byte, conn net.Conn) (*Chacha20Stream, error) {
	s := &Chacha20Stream{
		key:  key, // should be exactly 32 bytes
		conn: conn,
	}

	var err error
	nonce := make([]byte, chacha20.NonceSizeX)
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}

	s.encoder, err = chacha20.NewUnauthenticatedCipher(s.key, nonce)
	if err != nil {
		return nil, err
	}

	if n, err := s.conn.Write(nonce); err != nil || n != len(nonce) {
		return nil, errors.New("write nonce failed: " + err.Error())
	}
	return s, nil
}

func (s *Chacha20Stream) Read(p []byte) (int, error) {
	if s.decoder == nil {
		nonce := make([]byte, chacha20.NonceSizeX)
		if n, err := io.ReadAtLeast(s.conn, nonce, len(nonce)); err != nil || n != len(nonce) {
			return n, errors.New("can't read nonce from stream: " + err.Error())
		}
		decoder, err := chacha20.NewUnauthenticatedCipher(s.key, nonce)
		if err != nil {
			return 0, errors.New("generate decoder failed: " + err.Error())
		}
		s.decoder = decoder
	}

	n, err := s.conn.Read(p)
	if err != nil || n == 0 {
		return n, err
	}

	dst := make([]byte, n)
	pn := p[:n]
	s.decoder.XORKeyStream(dst, pn)
	copy(pn, dst)
	return n, nil
}

func (s *Chacha20Stream) Write(p []byte) (int, error) {
	dst := make([]byte, len(p))
	s.encoder.XORKeyStream(dst, p)
	return s.conn.Write(dst)
}

func (s *Chacha20Stream) Close() error {
	return s.conn.Close()
}

func Socks5Forward(client, target CipherStream) {
	forward := func(src, dest CipherStream) {
		defer src.Close()
		defer dest.Close()
		io.Copy(src, dest)
	}
	go forward(client, target)
	go forward(target, client)
}
func main() {
	listenAddr := flag.String("listenAddr", "127.0.0.1:19077", "")
	remoteAddr := flag.String("remoteAddr", "", "")
	role := flag.String("role", "B", "A or B")
	secret := flag.String("secret", "test", "")
	flag.Parse()

	if *secret == "" {
		fmt.Println("Please specify a secret.")
		return
	}

	GlobalConfig.ListenAddr = *listenAddr
	GlobalConfig.RemoteAddr = *remoteAddr
	GlobalConfig.Key = []byte(fmt.Sprintf("%x", md5.Sum([]byte(*secret))))
	fmt.Printf("[%s] -> [%s], role = %s, secret = %s\n", *listenAddr, *remoteAddr, *role, *secret)

	if *role == "A" {
		serverA()
	} else {
		serverB()
	}
}

func serverA() {
	server, err := net.Listen("tcp", GlobalConfig.ListenAddr)
	if err != nil {
		fmt.Printf("Listen failed: %v\n", err)
		return
	}

	for {
		client, err := server.Accept()
		if err != nil {
			fmt.Printf("Accept failed: %v\n", err)
			continue
		}
		go RelayTCPToKCP(client)
	}
}

func RelayTCPToKCP(client net.Conn) {
	block, _ := kcp.NewNoneBlockCrypt(nil)
	sess, err := kcp.DialWithOptions(GlobalConfig.RemoteAddr, block, 10, 3)
	if err != nil {
		client.Close()
		return
	}

	remote, err := NewChacha20Stream(GlobalConfig.Key, sess)
	if err != nil {
		client.Close()
		return
	}

	Socks5Forward(client, remote)
}

func serverB() {
	block, _ := kcp.NewNoneBlockCrypt(nil)
	listener, err := kcp.ListenWithOptions(GlobalConfig.ListenAddr, block, 10, 3)
	if err != nil {
		log.Fatal(err)
	}
	for {
		client, err := listener.AcceptKCP()
		if err != nil {
			fmt.Println("err accept kcp client:", err)
			continue
		}
		go RelayKCPToTCP(client)
	}
}

func RelayKCPToTCP(client net.Conn) {
	src, err := NewChacha20Stream(GlobalConfig.Key, client)
	if err != nil {
		client.Close()
		return
	}

	remote, err := net.Dial("tcp", GlobalConfig.RemoteAddr)
	if err != nil {
		client.Close()
		return
	}
	Socks5Forward(src, remote)
}