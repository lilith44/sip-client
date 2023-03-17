package main

import (
	"fmt"
	"strings"
)

type ClientOptions struct {
	Protocol string
	Server   ConnectionConfig
	Local    ConnectionConfig

	AutoReconnect bool

	User string

	MessageHandler MessageHandler
}

type ConnectionConfig struct {
	Host string
	Port int
}

func (c ConnectionConfig) FullHost() string {
	if c.Port == 0 || strings.Contains(c.Host, ":") {
		return c.Host
	}

	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}
