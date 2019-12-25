package main

import (
	"flag"
	"log"
	"net/http"
	"strings"

	"github.com/Jille/bindlink/linkmap"
	"github.com/Jille/bindlink/multiplexer"
	"github.com/Jille/bindlink/tundev"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	listenPort  = flag.Int("listen_port", 0, "Listen for incoming connections on this port")
	httpAddr    = flag.String("http_listen_port", ":8080", "Listen on this address for stats")
	targets     = flag.String("targets", "", "Host:port pairs of direct endpoints to connect to")
	proxies     = flag.String("proxies", "", "Host:port pairs of proxy servers")
	proxyTarget = flag.String("proxy_target", "", "Host:port pair to have proxy servers connect to")
)

func main() {
	flag.Parse()

	http.Handle("/metrics", promhttp.Handler())
	go func() {
		log.Fatal(http.ListenAndServe(*httpAddr, nil))
	}()

	isMaster := *listenPort > 0
	tun, err := tundev.New(isMaster)
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
	for _, p := range strings.Split(*targets, ",") {
		if p == "" {
			continue
		}
		if err := lm.InitiateLink(p); err != nil {
			log.Fatalf("Failed to connect to peer %q: %v", p, err)
		}
	}
	for _, p := range strings.Split(*proxies, ",") {
		if p == "" {
			continue
		}
		if err := lm.InitiateLinkOverSOCKS(p, *proxyTarget); err != nil {
			log.Fatalf("Failed to connect to peer %q: %v", p, err)
		}
	}
	mp.Start(tun.Send, lm.Send)
	go tun.Run(mp.Send)
	lm.Run()
}
