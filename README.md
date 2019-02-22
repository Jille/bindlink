# Bindlink

[![Build Status](https://travis-ci.org/Jille/bindlink.png)](https://travis-ci.org/Jille/bindlink)

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
