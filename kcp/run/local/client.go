package main

import (
	"flag"
	"game-proxy/kcp/config"
	"game-proxy/kcp/service"
	"net/http"
	_ "net/http/pprof"
)

var (
	addr     = flag.String("addr", ":9077", "Local hoz listen address")
	kcp      = flag.Bool("kcp", false, "use kcp protocol")
	remote   = flag.String("remote", "162.14.8.228:19077", "Remote hoz server address")
	password = flag.String("password", "little://!@adDxS$&(dl/*?QKc$mJ?PdTkajGzSNMILH{t4_hvFR>", "Cipher password string")
)

func main() {
	flag.Parse()
	s := service.NewServer(config.Config{
		Addr:       *addr,
		RemoteAddr: *remote,
		Cipher:     *password,
		KCP:        *kcp,
	})
	go func() {
		http.ListenAndServe(":7777", nil)
	}()
	s.Start()
}
