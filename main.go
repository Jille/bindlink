package main

import (
	"flag"
	"strings"

	"github.com/Jille/bindlink/linkmap"
	"github.com/Jille/bindlink/multiplexer"
	"github.com/Jille/bindlink/tundev"
)

var (
	listenPort = flag.Int("listen_port", 0, "Listen for incoming connections on this port")
	proxies    = flag.String("proxies", "", "Host:port pairs of proxy servers")
)

func main() {
	flag.Parse()
	tun, _ := tundev.New()
	mp := multiplexer.New()
	lm := linkmap.New(mp)
	if *listenPort > 0 {
		_ = lm.StartListener(*listenPort)
	}
	for _, p := range strings.Split(*proxies, ",") {
		if p == "" {
			continue
		}
		_ = lm.InitiateLink(p)
	}
	mp.Start(tun.Send, lm.Send)
	lm.Run()
}
