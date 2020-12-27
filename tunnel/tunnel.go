package tunnel

import (
	"context"
	"io"
	"net"
)

type Conn interface {
	net.Conn
	Metadata() *Metadata
}

type PacketConn interface {
	net.PacketConn
	WriteWithMetadata([]byte, *Metadata) (int, error)
	ReadWithMetadata([]byte) (int, *Metadata, error)
}

type ConnListener interface {
	AcceptConn(Tunnel) (Conn, error)
}

type PacketListener interface {
	AcceptPacket(Tunnel) (PacketConn, error)
}

type ConnDialer interface {
	DialConn(*Address, Tunnel) (net.Conn, error)
}

type PacketDialer interface {
	DialPacket(Tunnel) (PacketConn, error)
}

type Dialer interface {
	ConnDialer
}

type Client interface {
	Dialer
	io.Closer
}

type Server interface {
	ConnListener
	PacketListener
	io.Closer
}

type Tunnel interface {
	Name() string
	// NewClient(context.Context, Client) (Client, error)
	NewServer(context.Context, Server) (Server, error)
}
