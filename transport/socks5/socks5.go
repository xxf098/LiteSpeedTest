package socks5

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"net"
	"strconv"
)

// Error represents a SOCKS error
type Error byte

func (err Error) Error() string {
	return "SOCKS error: " + strconv.Itoa(int(err))
}

const (
	AtypIPv4       = 1
	AtypDomainName = 3
	AtypIPv6       = 4
)

// SOCKS errors as defined in RFC 1928 section 6.
const (
	ErrGeneralFailure       = Error(1)
	ErrConnectionNotAllowed = Error(2)
	ErrNetworkUnreachable   = Error(3)
	ErrHostUnreachable      = Error(4)
	ErrConnectionRefused    = Error(5)
	ErrTTLExpired           = Error(6)
	ErrCommandNotSupported  = Error(7)
	ErrAddressNotSupported  = Error(8)
)

// MaxAddrLen is the maximum size of SOCKS address in bytes.
const MaxAddrLen = 1 + 1 + 255 + 2

type Addr []byte

func (a Addr) String() string {
	var host, port string

	switch a[0] {
	case AtypDomainName:
		hostLen := uint16(a[1])
		host = string(a[2 : 2+hostLen])
		port = strconv.Itoa((int(a[2+hostLen]) << 8) | int(a[2+hostLen+1]))
	case AtypIPv4:
		host = net.IP(a[1 : 1+net.IPv4len]).String()
		port = strconv.Itoa((int(a[1+net.IPv4len]) << 8) | int(a[1+net.IPv4len+1]))
	case AtypIPv6:
		host = net.IP(a[1 : 1+net.IPv6len]).String()
		port = strconv.Itoa((int(a[1+net.IPv6len]) << 8) | int(a[1+net.IPv6len+1]))
	}

	return net.JoinHostPort(host, port)
}

func (a Addr) UDPAddr() *net.UDPAddr {
	if len(a) == 0 {
		return nil
	}
	switch a[0] {
	case AtypIPv4:
		var ip [net.IPv4len]byte
		copy(ip[0:], a[1:1+net.IPv4len])
		return &net.UDPAddr{IP: net.IP(ip[:]), Port: int(binary.BigEndian.Uint16(a[1+net.IPv4len : 1+net.IPv4len+2]))}
	case AtypIPv6:
		var ip [net.IPv6len]byte
		copy(ip[0:], a[1:1+net.IPv6len])
		return &net.UDPAddr{IP: net.IP(ip[:]), Port: int(binary.BigEndian.Uint16(a[1+net.IPv6len : 1+net.IPv6len+2]))}
	}
	// Other Atyp
	return nil
}

func SplitAddr(b []byte) Addr {
	addrLen := 1
	if len(b) < addrLen {
		return nil
	}

	switch b[0] {
	case AtypDomainName:
		if len(b) < 2 {
			return nil
		}
		addrLen = 1 + 1 + int(b[1]) + 2
	case AtypIPv4:
		addrLen = 1 + net.IPv4len + 2
	case AtypIPv6:
		addrLen = 1 + net.IPv6len + 2
	default:
		return nil

	}

	if len(b) < addrLen {
		return nil
	}

	return b[:addrLen]
}

func ReadAddr(r io.Reader, b []byte) (Addr, error) {
	if len(b) < MaxAddrLen {
		return nil, io.ErrShortBuffer
	}
	_, err := io.ReadFull(r, b[:1]) // read 1st byte for address type
	if err != nil {
		return nil, err
	}

	switch b[0] {
	case AtypDomainName:
		_, err = io.ReadFull(r, b[1:2]) // read 2nd byte for domain length
		if err != nil {
			return nil, err
		}
		domainLength := uint16(b[1])
		_, err = io.ReadFull(r, b[2:2+domainLength+2])
		return b[:1+1+domainLength+2], err
	case AtypIPv4:
		_, err = io.ReadFull(r, b[1:1+net.IPv4len+2])
		return b[:1+net.IPv4len+2], err
	case AtypIPv6:
		_, err = io.ReadFull(r, b[1:1+net.IPv6len+2])
		return b[:1+net.IPv6len+2], err
	}

	return nil, ErrAddressNotSupported
}

func ParseAddrToSocksAddr(addr net.Addr) Addr {
	var hostip net.IP
	var port int
	if udpaddr, ok := addr.(*net.UDPAddr); ok {
		hostip = udpaddr.IP
		port = udpaddr.Port
	} else if tcpaddr, ok := addr.(*net.TCPAddr); ok {
		hostip = tcpaddr.IP
		port = tcpaddr.Port
	}

	// fallback parse
	if hostip == nil {
		return ParseAddr(addr.String())
	}

	var parsed Addr
	if ip4 := hostip.To4(); ip4.DefaultMask() != nil {
		parsed = make([]byte, 1+net.IPv4len+2)
		parsed[0] = AtypIPv4
		copy(parsed[1:], ip4)
		binary.BigEndian.PutUint16(parsed[1+net.IPv4len:], uint16(port))

	} else {
		parsed = make([]byte, 1+net.IPv6len+2)
		parsed[0] = AtypIPv6
		copy(parsed[1:], hostip)
		binary.BigEndian.PutUint16(parsed[1+net.IPv6len:], uint16(port))
	}
	return parsed
}

// ParseAddr parses the address in string s. Returns nil if failed.
func ParseAddr(s string) Addr {
	var addr Addr
	host, port, err := net.SplitHostPort(s)
	if err != nil {
		return nil
	}
	if ip := net.ParseIP(host); ip != nil {
		if ip4 := ip.To4(); ip4 != nil {
			addr = make([]byte, 1+net.IPv4len+2)
			addr[0] = AtypIPv4
			copy(addr[1:], ip4)
		} else {
			addr = make([]byte, 1+net.IPv6len+2)
			addr[0] = AtypIPv6
			copy(addr[1:], ip)
		}
	} else {
		if len(host) > 255 {
			return nil
		}
		addr = make([]byte, 1+1+len(host)+2)
		addr[0] = AtypDomainName
		addr[1] = byte(len(host))
		copy(addr[2:], host)
	}

	portnum, err := strconv.ParseUint(port, 10, 16)
	if err != nil {
		return nil
	}

	addr[len(addr)-2], addr[len(addr)-1] = byte(portnum>>8), byte(portnum)

	return addr
}

func EncodeUDPPacket(addr Addr, payload []byte) (packet []byte, err error) {
	if addr == nil {
		err = errors.New("address is invalid")
		return
	}
	packet = bytes.Join([][]byte{{0, 0, 0}, addr, payload}, []byte{})
	return
}
