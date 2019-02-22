package main

import (
	"flag"
	"log"
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
	tun, err := tundev.New()
	if err != nil {
		log.Fatalf("Failed to create TUN device: %v", err)
	}
	mp := multiplexer.New()
	lm := linkmap.New(mp)
	if *listenPort > 0 {
		if err := lm.StartListener(*listenPort); err != nil {
			log.Fatalf("Failed to start listening socket: %v", err)
		}
	}
	for _, p := range strings.Split(*proxies, ",") {
		if p == "" {
			continue
		}
		if err := lm.InitiateLink(p); err != nil {
			log.Fatalf("Failed to connect to peer %q: %v", p, err)
		}
	}
	mp.Start(tun.Send, lm.Send)
	go tun.Run(mp.Send)
	lm.Run()
}
