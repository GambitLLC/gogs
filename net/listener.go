package net

import (
	"bufio"
	pk "gogs/net/packet"
	"net"
	"strconv"
)

type MCListener struct{ net.Listener }

func NewListener(host string, port int) (*MCListener, error) {
	addr, err := net.ResolveTCPAddr("tcp", host+":"+strconv.Itoa(port))
	if err != nil {
		return nil, err
	}

	tcp, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return nil, err
	}

	return &MCListener{tcp}, nil
}

func (l *MCListener) Accept() (Conn, error) {
	c, err := l.Listener.Accept()
	return &conn{
		Socket: c,
		Reader: bufio.NewReader(c),
	}, err
}

type Conn interface {
	ReadPacket() (pk.Packet, error)
	WritePacket(p pk.Packet) error
	Close() error
}

type conn struct {
	Socket net.Conn
	Reader pk.Reader
}

func (c *conn) ReadPacket() (pk.Packet, error) {
	return pk.Decode(c.Reader)
}

func (c *conn) WritePacket(p pk.Packet) error {
	_, err := c.Socket.Write(p.Encode())
	return err
}

func (c *conn) Close() error {
	return c.Socket.Close()
}
