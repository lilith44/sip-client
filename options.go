package sip_client

import (
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
)

type Options struct {
	Server ServerOptions
	Client ClientOptions

	MessageHandler MessageHandler
	Logger         *zap.SugaredLogger
}

type ServerOptions struct {
	Protocol      string
	Host          string
	Port          int
	Timeout       time.Duration
	AutoReconnect bool
}

type ClientOptions struct {
	Host     string
	Port     int
	User     UserOptions
	Register RegisterOptions
}

type UserOptions struct {
	Name     string
	Domain   string
	Password string
	Agent    string
}

type RegisterOptions struct {
	Expire            int
	KeepaliveInterval int
}

func (s ServerOptions) FullHost() string {
	if s.Port == 0 || strings.Contains(s.Host, ":") {
		return s.Host
	}

	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}

func (c ClientOptions) FullHost() string {
	if c.Port == 0 || strings.Contains(c.Host, ":") {
		return c.Host
	}

	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}
