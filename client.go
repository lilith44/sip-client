package main

import (
	"bytes"
	"net"
	"sync"

	"github.com/jart/gosip/sip"
	"go.uber.org/zap"
)

type client struct {
	conn      net.Conn
	connMutex sync.RWMutex

	options ClientOptions

	logger *zap.SugaredLogger

	pool sync.Map

	messageHandler MessageHandler
}

type channel chan *sip.Msg

type MessageHandler func(c *client, msg *sip.Msg)

func NewClient(options ClientOptions) *client {
	logger, _ := zap.NewDevelopment()

	return &client{
		options: options,

		logger:         logger.Sugar(),
		messageHandler: options.MessageHandler,
	}
}

func (c *client) Connect() error {
	c.connMutex.Lock()
	defer c.connMutex.Unlock()

	var conn net.Conn
	if c.options.Protocol == "udp" {
		local, err := net.ResolveUDPAddr("udp", c.options.Local.FullHost())
		if err != nil {
			return err
		}
		server, err := net.ResolveUDPAddr("udp", c.options.Server.FullHost())
		if err != nil {
			return err
		}

		conn, err = net.DialUDP(c.options.Protocol, local, server)
		if err != nil {
			return err
		}
	}

	c.conn = conn
	return nil
}

func (c *client) Disconnect() {
	c.connMutex.Lock()
	defer c.connMutex.Unlock()

	_ = c.conn.Close()
}

func (c *client) Listen() {
	var (
		n      int
		err    error
		buffer []byte
	)

	for {
		buffer = make([]byte, 2048)
		n, err = c.conn.Read(buffer)
		if err != nil {
			c.logger.Errorf("read data failed: %s", err)
			continue
		}

		var msg *sip.Msg
		msg, err = sip.ParseMsg(buffer[:n])
		if err != nil {
			c.logger.Errorf("parse sip message failed: %s", err)
			continue
		}

		c.logger.Infof("receive message from %s\n%s", c.options.Server.FullHost(), msg)

		ch, loaded := c.pool.LoadAndDelete(msg.CSeq)
		if !loaded {
			c.messageHandler(c, msg)
			continue
		}

		ch.(channel) <- msg
	}
}

func (c *client) Send(msg *sip.Msg) (*sip.Msg, error) {
	var b bytes.Buffer
	msg.Append(&b)
	c.logger.Infof("send message to %s\n%s", c.options.Server.FullHost(), msg)

	n, err := c.conn.Write(b.Bytes())
	if err != nil {
		c.logger.Errorf("send message error: %s", err)
		return nil, err
	}
	if n != b.Len() {
		c.logger.Errorf("send message error: %s", err)
		return nil, err
	}

	ch := make(channel)
	c.pool.Store(msg.CSeq, ch)

	return <-ch, nil
}
