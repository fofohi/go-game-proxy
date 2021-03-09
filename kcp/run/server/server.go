package main

import (
	"flag"
	"game-proxy/kcp/config"
	"game-proxy/kcp/service"
	"net/http"
	_ "net/http/pprof"
)

var (
	addrServer     = flag.String("addr", ":9077", "Local hoz listen address")
	kcpServer      = flag.Bool("kcp", false, "use kcp protocol")
	passwordServer = flag.String("password", "little://!@adDxS$&(dl/*?QKc$mJ?PdTkajGzSNMILH{t4_hvFR>", "Cipher password string")
)

func main() {
	flag.Parse()
	s := service.NewServer(config.Config{
		Addr:   *addrServer,
		Cipher: *passwordServer,
		KCP:    *kcpServer,
	})
	go func() {
		http.ListenAndServe(":7778", nil)
	}()
	s.Start()
}
