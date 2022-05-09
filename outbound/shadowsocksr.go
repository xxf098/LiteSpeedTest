package outbound

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"strconv"

	"github.com/Dreamacro/go-shadowsocks2/core"
	"github.com/Dreamacro/go-shadowsocks2/shadowaead"
	"github.com/Dreamacro/go-shadowsocks2/shadowstream"
	C "github.com/xxf098/lite-proxy/constant"
	"github.com/xxf098/lite-proxy/transport/dialer"
	"github.com/xxf098/lite-proxy/transport/ssr/obfs"
	"github.com/xxf098/lite-proxy/transport/ssr/protocol"
)

type ShadowSocksR struct {
	*Base
	cipher   core.Cipher
	obfs     obfs.Obfs
	protocol protocol.Protocol
}

type ShadowSocksROption struct {
	Name          string `proxy:"name,omitempty"`
	Server        string `proxy:"server"`
	Port          int    `proxy:"port"`
	Password      string `proxy:"password"`
	Cipher        string `proxy:"cipher"`
	Obfs          string `proxy:"obfs"`
	ObfsParam     string `proxy:"obfs-param,omitempty"`
	Protocol      string `proxy:"protocol"`
	ProtocolParam string `proxy:"protocol-param,omitempty"`
	UDP           bool   `proxy:"udp,omitempty"`
	Remarks       string `proxy:"remarks,omitempty"`
}

func (ssr *ShadowSocksR) StreamConn(c net.Conn, metadata *C.Metadata) (net.Conn, error) {
	c = ssr.obfs.StreamConn(c)
	c = ssr.cipher.StreamConn(c)
	var (
		iv  []byte
		err error
	)
	switch conn := c.(type) {
	case *shadowstream.Conn:
		iv, err = conn.ObtainWriteIV()
		if err != nil {
			return nil, err
		}
	case *shadowaead.Conn:
		return nil, fmt.Errorf("invalid connection type")
	}
	c = ssr.protocol.StreamConn(c, iv)
	_, err = c.Write(serializesSocksAddr(metadata))
	return c, err
}

func (ssr *ShadowSocksR) DialContext(ctx context.Context, metadata *C.Metadata) (net.Conn, error) {
	c, err := dialer.DialContext(ctx, "tcp", ssr.addr)
	if err != nil {
		return nil, fmt.Errorf("%s connect error: %w", ssr.addr, err)
	}
	tcpKeepAlive(c)

	return ssr.StreamConn(c, metadata)
}

func (ssr *ShadowSocksR) DialUDP(metadata *C.Metadata) (net.PacketConn, error) {
	pc, err := dialer.ListenPacket("udp", "")
	if err != nil {
		return nil, err
	}

	addr, err := resolveUDPAddr("udp", ssr.addr)
	if err != nil {
		return nil, err
	}

	pc = ssr.cipher.PacketConn(pc)
	pc = ssr.protocol.PacketConn(pc)
	return &ssPacketConn{PacketConn: pc, rAddr: addr}, nil
}

func (ssr *ShadowSocksR) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]string{
		"type": "shadowsocksr",
	})
}

func NewShadowSocksR(option *ShadowSocksROption) (*ShadowSocksR, error) {
	addr := net.JoinHostPort(option.Server, strconv.Itoa(option.Port))
	if option.Cipher == "none" {
		option.Cipher = "dummy"
	}
	cipher := option.Cipher
	password := option.Password
	coreCiph, err := core.PickCipher(cipher, nil, password)
	if err != nil {
		return nil, fmt.Errorf("ssr %s initialize error: %w", addr, err)
	}
	var (
		ivSize int
		key    []byte
	)
	if option.Cipher == "dummy" {
		ivSize = 0
		key = core.Kdf(option.Password, 16)
	} else {
		ciph, ok := coreCiph.(*core.StreamCipher)
		if !ok {
			return nil, fmt.Errorf("%s is not dummy or a supported stream cipher in ssr", cipher)
		}
		ivSize = ciph.IVSize()
		key = ciph.Key
	}

	obfs, obfsOverhead, err := obfs.PickObfs(option.Obfs, &obfs.Base{
		Host:   option.Server,
		Port:   option.Port,
		Key:    key,
		IVSize: ivSize,
		Param:  option.ObfsParam,
	})
	if err != nil {
		return nil, fmt.Errorf("ssr %s initialize obfs error: %w", addr, err)
	}

	protocol, err := protocol.PickProtocol(option.Protocol, &protocol.Base{
		Key:      key,
		Overhead: obfsOverhead,
		Param:    option.ProtocolParam,
	})
	if err != nil {
		return nil, fmt.Errorf("ssr %s initialize protocol error: %w", addr, err)
	}

	return &ShadowSocksR{
		Base: &Base{
			name: option.Name,
			addr: addr,
			udp:  option.UDP,
		},
		cipher:   coreCiph,
		obfs:     obfs,
		protocol: protocol,
	}, nil
}
