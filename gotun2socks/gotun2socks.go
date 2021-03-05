package gotun2socks

import (
	"game-proxy/gotun2socks/tun"
	"game-proxy/gotun2socks/tun2socks"
	"log"
	"os"
	"runtime"
	"runtime/debug"
	"runtime/pprof"
)

type JavaUidCallback interface {
	FindUid(sourceIp string, sourcePort int, destIp string, destPort int) int
}

type Callbacks struct {
	uidCallback JavaUidCallback
}

func (c Callbacks) GetUid(sourceIp string, sourcePort uint16, destIp string, destPort uint16) int {
	if c.uidCallback == nil {
		log.Printf("uid callback is nil")
	}

	return c.uidCallback.FindUid(sourceIp, int(sourcePort), destIp, int(destPort))
}

var tun2SocksInstance *tun2socks.Tun2Socks
var defaultProxy = &tun2socks.ProxyServer{
	ProxyType:  tun2socks.PROXY_TYPE_NONE,
	IpAddress:  ":",
	AuthHeader: "",
	Login:      "",
	Password:   "",
}

var callback *Callbacks = nil

var proxyServerMap map[int]*tun2socks.ProxyServer

func SayHi() string {
	return "hi from tun2http!"
}

func AddProxyServer(uid int, ipPort string, proxyType int, httpAuthHeader string, login string, password string) {
	if proxyServerMap == nil {
		proxyServerMap = make(map[int]*tun2socks.ProxyServer)
	}

	if len(ipPort) < 8 {
		proxyType = tun2socks.PROXY_TYPE_NONE
	}

	proxy := &tun2socks.ProxyServer{
		ProxyType:  proxyType,
		IpAddress:  ipPort,
		AuthHeader: httpAuthHeader,
		Login:      login,
		Password:   password,
	}

	proxyServerMap[uid] = proxy
	log.Printf("Set proxy for uid %d", uid)
}

func SetDefaultProxy(ipPort string, proxyType int, httpAuthHeader string, login string, password string) {
	if len(ipPort) < 8 {
		proxyType = tun2socks.PROXY_TYPE_NONE
	}

	defaultProxy = &tun2socks.ProxyServer{
		ProxyType:  proxyType,
		IpAddress:  ipPort,
		AuthHeader: httpAuthHeader,
		Login:      login,
		Password:   password,
	}
	log.Printf("Set default proxy")
}

func SetUidCallback(javaCallback JavaUidCallback) {
	callback = &Callbacks{
		uidCallback: javaCallback,
	}

	if tun2SocksInstance != nil {
		tun2SocksInstance.SetUidCallback(callback)
	}

	log.Printf("Uid callback set")
}

func Run(descriptor int, maxCpus int) {
	runtime.GOMAXPROCS(maxCpus)

	var tunAddr string = "10.253.253.253"
	var tunGW string = "10.0.0.1"
	var enableDnsCache bool = true

	f := tun.NewTunDev(uintptr(descriptor), "tun0", tunAddr, tunGW)
	tun2SocksInstance = tun2socks.New(f, enableDnsCache)

	tun2SocksInstance.SetDefaultProxy(defaultProxy)
	tun2SocksInstance.SetProxyServers(proxyServerMap)
	if callback != nil && callback.uidCallback != nil {
		tun2SocksInstance.SetUidCallback(callback)
	} else {
		tun2SocksInstance.SetUidCallback(nil)
	}

	go func() {
		tun2SocksInstance.Run()
	}()

	log.Printf("Tun2Htpp started")
	debug.SetTraceback("all")
	debug.SetPanicOnFault(true)
}

func Stop() {
	tun2SocksInstance.Stop()
}

func Prof() {
	pprof.Lookup("goroutine").WriteTo(os.Stdout, 1)
}
