package sip_client

import (
	"context"
	"errors"

	"github.com/jart/gosip/sip"
)

type Channel struct {
	ctx    context.Context
	cancel context.CancelFunc

	count int
	ch    []*sip.Msg
}

func NewChannel(ctx context.Context, cancel context.CancelFunc, count int) *Channel {
	return &Channel{
		ctx:    ctx,
		cancel: cancel,
		count:  count,
	}
}

func (c *Channel) Receive(msg *sip.Msg) (completed bool) {
	if c.count == 0 {
		return true
	}

	c.ch = append(c.ch, msg)
	if len(c.ch) == c.count {
		completed = true
		c.cancel()
	}

	return completed
}

func (c *Channel) wait() ([]*sip.Msg, error) {
	<-c.ctx.Done()

	err := c.ctx.Err()
	if errors.Is(err, context.Canceled) {
		return c.ch, nil
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return nil, errors.New("timed out")
	}
	return nil, err
}
