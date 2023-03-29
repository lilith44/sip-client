package sip_client

import (
	"bytes"
	"context"
	"net"
	"runtime/debug"
	"sync"
	"time"

	"github.com/jart/gosip/sip"
	"go.uber.org/zap"
)

type Client struct {
	conn  net.Conn
	mutex sync.RWMutex

	server ServerOptions
	client ClientOptions

	messageHandler MessageHandler
	registerFunc   RegisterFunc
	logger         *zap.SugaredLogger
	pool           sync.Map
}

type MessageHandler func(c *Client, msg *sip.Msg)

type RegisterFunc func(c *Client) error

type msgResponse struct {
	msgs []*sip.Msg
	err  error
}

func NewClient(options Options) (*Client, error) {
	c := &Client{
		server:         options.Server,
		client:         options.Client,
		messageHandler: options.MessageHandler,
		registerFunc:   options.RegisterFunc,
		logger:         options.Logger,
	}

	if err := c.Connect(); err != nil {
		return nil, err
	}

	go c.Listen()

	go func() {
		for {
			time.Sleep(time.Duration(c.client.Register.KeepaliveInterval) * time.Second)

			c.logger.Infof("开始保活")
			err := c.registerFunc(c)
			if err != nil {
				c.logger.Errorf("保活失败：%s", err)
			}
		}
	}()

	if err := c.registerFunc(c); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Client) ClientOptions() ClientOptions {
	return c.client
}

func (c *Client) ServerOptions() ServerOptions {
	return c.server
}

func (c *Client) Logger() *zap.SugaredLogger {
	return c.logger
}

func (c *Client) Connect() error {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return c.connect()
}

func (c *Client) connect() error {
	var conn net.Conn
	if c.server.Protocol == "udp" {
		local, err := net.ResolveUDPAddr("udp", c.client.FullHost())
		if err != nil {
			return err
		}
		server, err := net.ResolveUDPAddr("udp", c.server.FullHost())
		if err != nil {
			return err
		}

		conn, err = net.DialUDP(c.server.Protocol, local, server)
		if err != nil {
			return err
		}
	}

	c.conn = conn
	return nil
}

func (c *Client) reconnect() error {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	c.logger.Warnf("连接断开，重新进行连接")

	_ = c.conn.Close()
	return c.connect()
}

func (c *Client) register() error {
	return c.registerFunc(c)
}

func (c *Client) SetMessageHandler(handler MessageHandler) {
	c.messageHandler = handler
}

func (c *Client) Listen() {
	for {
		ch := make(chan struct{})
		go c.listen(ch)
		<-ch
	}
}

func (c *Client) Send(msg *sip.Msg) (*sip.Msg, error) {
	ch := make(chan *msgResponse)
	defer close(ch)
	go c.send(msg, ch, 1)
	rsp := <-ch
	if len(rsp.msgs) == 0 {
		return nil, rsp.err
	}

	return rsp.msgs[0], rsp.err
}

func (c *Client) SendForNoResponse(msg *sip.Msg) error {
	ch := make(chan *msgResponse)
	defer close(ch)
	go c.send(msg, ch, 0)
	rsp := <-ch
	return rsp.err
}

func (c *Client) SendForMultiResponse(msg *sip.Msg, responseCount int) ([]*sip.Msg, error) {
	ch := make(chan *msgResponse)
	defer close(ch)
	go c.send(msg, ch, responseCount)
	rsp := <-ch
	return rsp.msgs, rsp.err
}

func (c *Client) listen(ch chan struct{}) {
	defer func() {
		if r := recover(); r != nil {
			c.logger.Errorf("[Panic]%s\n%s", r, debug.Stack())
		}

		ch <- struct{}{}
	}()

	var (
		n      int
		err    error
		buffer []byte
	)

	for {
		buffer = make([]byte, 2048)
		n, err = c.conn.Read(buffer)
		if err != nil {
			c.logger.Errorf("读取数据失败：%s", err)
			err = c.reconnect()
			if err != nil {
				c.logger.Errorf("重新连接失败：%s", err)
			}
			continue
		}

		var msg *sip.Msg
		msg, err = sip.ParseMsg(buffer[:n])
		if err != nil {
			c.logger.Errorf("解析sip消息失败：%s", err)
			continue
		}

		c.logger.Infof("收到来自%s的消息\n%s", c.server.FullHost(), msg)

		channel, loaded := c.pool.Load(msg.CallID)
		if !loaded {
			go c.messageHandler(c, msg)
			continue
		}

		if channel.(*Channel).Receive(msg) {
			c.pool.Delete(msg.CallID)
		}
	}
}

func (c *Client) send(msg *sip.Msg, ch chan *msgResponse, responseCount ...int) {
	rsp := new(msgResponse)
	defer func() {
		if r := recover(); r != nil {
			c.logger.Errorf("[Panic]%s\n%s", r, debug.Stack())
		}

		ch <- rsp
	}()

	var b bytes.Buffer
	msg.Append(&b)
	c.logger.Infof("发送消息至%s\n%s", c.server.FullHost(), msg)

	n, err := c.conn.Write(b.Bytes())
	if err != nil || n != b.Len() {
		rsp.err = err
		c.logger.Errorf("发送消息失败：%s", err)
		rsp.err = c.reconnect()
		return
	}

	count := 0
	if len(responseCount) != 0 {
		count = responseCount[0]
	}

	if count == 0 {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.server.Timeout*time.Second)

	channel := NewChannel(ctx, cancel, count)
	c.pool.Store(msg.CallID, channel)

	rsp.msgs, rsp.err = channel.wait()
	if err != nil {
		c.logger.Errorf("等待消息返回发送错误：%s", err)
	}

	return
}
