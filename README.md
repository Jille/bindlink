# Bindlink

[![Build Status](https://travis-ci.org/Jille/bindlink.png)](https://travis-ci.org/Jille/bindlink)

## High level overview

You have to run two instances of bindlink:

- a master, run on a reliable high-speed server
- a slave, run in a place with a crappy internet connection

The master listens on UDP on the `--listen_port` and waits for the slave to send packets.
The slave uses multiple internet connections, for example by installing a SOCKS server on multiple phones with 4G, to connect to the master and will start spreading traffic over all these links. The slave configures `--proxy_target` to be the external host+port of the master, and a list of SOCKS servers with `--proxies`. Each of these connections is called a link.

The slave decides how many links exists, and the master will just learn about them when it receives a packet through them.

## Internally

tundev is the edge of bindlink. It either exposes a tun devices to the kernel, or uses a TCP connection (when built with `--tags notun`). Traffic that flows into the master's tun, will flow out of the slave's tun on the other side and vice versa. Traffic is send to the system with tundev.Send(), and received by passing a callback into tundev.Run. This callback is usually multiplexer.Send.

multiplexer.Send() is responsible for choosing a link to send the packet over and sending it. The multiplexer chooses one (or more) links, and uses the linkmap's Send() to actually send it over that link.

The linkmap keeps track of all links that can be used to communicate over and abstracts how the links work. UDP and SOCKS links both have the same interface to send a packet over.

The linkmap also decodes packets and calls the multiplexer to handle them. Control packets are passed to multiplexer.HandleControl() and data packets to multiplexer.Received(). On receipt of a data packet the multiplexer will simply send it over to the tundev to pass it to the system and then the packet's journey is complete.

## Internal API

```go
// Reach out to a proxy server to start a connection to the remote daemon.
LinkMap.InitiateLink(proxyAddr)
// Bind on port and wait for the other side to call InitiateLink
LinkMap.StartListener(port)

type Link struct {
}
// Tunnel a packet over a link
Link.Send(packet)

// Teach the multiplexer about a new link to be used
Multiplexer.AddLink()
// Notify the multiplexer we received a control packet.
Multiplexer.HandleControl(Link, ControlPacket)
// Ask the multiplexer to call Link.Send on one or more links.
Multiplexer.Send(packet)
Multiplexer.CraftControl()
```
