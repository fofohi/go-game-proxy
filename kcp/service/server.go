package service

import (
	"game-proxy/kcp/cipher"
	"game-proxy/kcp/cipher/little"
	"game-proxy/kcp/config"
	"github.com/xtaci/kcp-go"
	"net"
	"strings"
)

type Server struct {
	config.Config
	cipher cipher.Cipher
	ln     net.Listener
}

func NewServer(config config.Config) *Server {
	return &Server{
		Config: config,
	}
}

func (s *Server) Start() {
	var ln net.Listener
	var err error
	if s.Config.RemoteAddr == "" && s.Config.KCP {
		ln, err = kcp.ListenWithOptions(s.Config.Addr, nil, 10, 3)
	} else {
		ln, err = net.Listen("tcp", s.Config.Addr)
	}
	if err != nil {
		//log.LOG.Printf("server startup err %v \n", err)
	}
	pass := strings.Split(s.Config.Cipher, "://")
	if len(pass) != 2 {
		//log.LOG.Fatal(errors.New("Cipher must be like scheme://password "))
		return
	}
	var waper cipher.Cipher
	var cipherName = pass[0]
	//var key = pass[1]
	switch cipherName {
	case "little":
		pwdStr := "BH1rStJwNP1YIvNI4Y+8ZVWyqsX47QCTOJTpGLnL2VQHqV0pPu8ZLk3yBc5sRNWmpYjqL2jY9LiFr9EaUsT1Voy3sBadZDKBPQ3g3yP6wOtvrHNxisbuTrPxEHZ6i6sSPAw6mB0rFEsB1OSjXPzlhkmb4lmee1+1aeOgHPaDmUF0vzskwS2iA4TK7ArJ1+fCvWJmY6i2/pDMh1qh3I3PJtBXyBUhET+7w9s5UfcXCVBTQ9beJ1tHC3d5TwgzgkJqkTGkHt1tp2HaTM0fcmd+lY43IP+tsbosJQb7lpqStA94gIlef/AwKnXTQJc1vkZF6Jz5bscCG2CuNhPmKJ8OfA=="
		pwd, err := little.ParsePassword(pwdStr)
		if err != nil {
			panic(err)
		}
		// 混淆加密
		waper = little.NewCipher(pwd)
	default:
		//log.LOG.Fatalf("Unsuport cipher %s \n", cipherName)
	}
	//log.LOG.Printf("cipher_name=%s, password=%s\n", cipherName, key)
	s.cipher = waper
	//log.LOG.Printf("Server startup, with config %+v", s.Config)
	for {
		conn, err := ln.Accept()
		if err != nil {
			//log.LOG.Printf("Accept connection err %v \n", err)
			panic(err)
		}
		nc := &Connection{s: s, conn: conn}
		go nc.handle()
	}
}
