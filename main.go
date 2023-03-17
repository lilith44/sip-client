package main

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/jart/gosip/sdp"
	"github.com/jart/gosip/sip"
	"github.com/jart/gosip/util"
	"github.com/labstack/echo/v4"
	"github.com/lilith44/echo-validator"
	"github.com/lilith44/sip-client/gb28281"
	"github.com/lilith44/sip-client/utils"
)

const (
	sipServerId   = "44020000002000000008"
	domain        = "4402000000"
	sipServerHost = "192.168.181.75"
	sipServerPort = 5060

	user   = "44020000002000001008"
	media  = "44020000002000001009"
	device = "99072898621320000141"
)

var template = `<? xmlversion="1.0"?>
<Control>
<CmdType>DeviceControl</CmdType>
<SN>11</SN>
<DeviceID>99072898621320000141</DeviceID>
<PTZCmd>%s</PTZCmd>
<Info>
<ControlPriority>5</ControlPriority>
</Info>
</Control>
`

func main() {
	c := NewClient(ClientOptions{
		Protocol: "udp",
		Server: ConnectionConfig{
			Host: sipServerHost,
			Port: sipServerPort,
		},
		Local: ConnectionConfig{
			Host: "192.168.118.62",
			Port: 5060,
		},
		AutoReconnect:  false,
		User:           sipServerId,
		MessageHandler: handler,
	})

	if err := c.Connect(); err != nil {
		return
	}

	go c.Listen()

	register(c)

	e := echo.New()
	e.Validator = echo_validator.New()
	e.HTTPErrorHandler = func(err error, c echo.Context) {
		_ = c.JSON(http.StatusInternalServerError, struct {
			Message string `json:"message"`
		}{Message: err.Error()})
	}

	g := e.Group("")
	g.POST("/controls", func(ctx echo.Context) (err error) {
		req := new(struct {
			Ptz gb28281.PtzConfig `json:"ptz"`
		})

		if err = ctx.Bind(req); err != nil {
			return err
		}
		if err = ctx.Validate(req); err != nil {
			return err
		}

		go func() {
			defer func() {
				if r := recover(); r != nil {
					err = errors.New("panic occurs")
				}
			}()

			sendCmd(c, gb28281.NewPtzCmd(req.Ptz).String())
		}()

		return nil
	})

	g.POST("/videos", func(ctx echo.Context) (err error) {
		video(c)

		return nil
	})

	e.Start(":8080")
}

func sendCmd(c *client, cmd string) {
	localAddr := c.conn.LocalAddr().(*net.UDPAddr)

	msg := &sip.Msg{
		Method: sip.MethodMessage,
		Request: &sip.URI{
			Scheme: "sip",
			User:   device,
			Host:   domain,
		},
		From: &sip.Addr{
			Uri: &sip.URI{
				Scheme: "sip",
				User:   user,
				Host:   domain,
			},
			Param: &sip.Param{Name: "tag", Value: util.GenerateTag()},
		},
		To: &sip.Addr{
			Uri: &sip.URI{
				Scheme: "sip",
				User:   device,
				Host:   domain,
			},
		},
		Via: &sip.Via{
			Host: localAddr.IP.String(),
			Port: uint16(localAddr.Port),
		},
		Contact: &sip.Addr{
			Uri: &sip.URI{
				Scheme: "sip",
				User:   user,
				Host:   localAddr.IP.String(),
				Port:   uint16(localAddr.Port),
			},
		},
		CallID:     fmt.Sprintf("%s@%s", util.GenerateCallID(), localAddr.IP.String()),
		CSeq:       util.GenerateCSeq(),
		CSeqMethod: sip.MethodMessage,
		UserAgent:  "IP Camera",
		Payload:    NewMessage([]byte(fmt.Sprintf(template, cmd))),
	}

	c.Send(msg)
}

func register(c *client) {
	msg := &sip.Msg{
		Method: sip.MethodRegister,
		Request: &sip.URI{
			Scheme: "sip",
			User:   sipServerId,
			Host:   domain,
		},
		From: &sip.Addr{
			Uri: &sip.URI{
				Scheme: "sip",
				User:   user,
				Host:   domain,
			},
			Param: &sip.Param{Name: "tag", Value: util.GenerateTag()},
		},
		To: &sip.Addr{
			Uri: &sip.URI{
				Scheme: "sip",
				User:   user,
				Host:   domain,
			},
		},
		Via: &sip.Via{
			Host: c.options.Local.Host,
			Port: uint16(c.options.Local.Port),
		},
		Contact: &sip.Addr{
			Uri: &sip.URI{
				Scheme: "sip",
				User:   user,
				Host:   c.options.Local.Host,
				Port:   uint16(c.options.Local.Port),
			},
		},
		CallID:     fmt.Sprintf("%s@%s", util.GenerateCallID(), c.options.Local.Host),
		CSeq:       1,
		CSeqMethod: sip.MethodRegister,
		Expires:    7200,
		UserAgent:  "IP Camera",
	}

	rsp, err := c.Send(msg)
	if err != nil {
		return
	}

	msg.CSeq = 2
	msg.CallID = fmt.Sprintf("%s@%s", util.GenerateCallID(), c.options.Local.Host)

	realmStart := strings.Index(rsp.WWWAuthenticate, `realm="`) + 7
	realmEnd := strings.Index(rsp.WWWAuthenticate[realmStart:], `"`)
	realm := rsp.WWWAuthenticate[realmStart : realmStart+realmEnd]

	nonceStart := strings.Index(rsp.WWWAuthenticate, `nonce="`) + 7
	nonceEnd := strings.Index(rsp.WWWAuthenticate[nonceStart:], `"`)
	nonce := rsp.WWWAuthenticate[nonceStart : nonceStart+nonceEnd]
	uri := fmt.Sprintf("sip:%s@%s", user, domain)

	msg.Authorization = fmt.Sprintf(`Digest username="%s",realm="%s",nonce="%s",uri="%s",response="%s",algorithm=MD5`, user, realm, nonce, uri, utils.CalculateResponse(user, realm, "12345678", sip.MethodRegister, uri, nonce))

	c.Send(msg)
}

func video(c *client) {
	msg := &sip.Msg{
		Method: sip.MethodInvite,
		Request: &sip.URI{
			Scheme: "sip",
			User:   device,
			Host:   domain,
		},
		From: &sip.Addr{
			Uri: &sip.URI{
				Scheme: "sip",
				User:   user,
				Host:   domain,
			},
			Param: &sip.Param{Name: "tag", Value: util.GenerateTag()},
		},
		To: &sip.Addr{
			Uri: &sip.URI{
				Scheme: "sip",
				User:   device,
				Host:   domain,
			},
		},
		Via: &sip.Via{
			Host: c.options.Local.Host,
			Port: uint16(c.options.Local.Port),
		},
		Contact: &sip.Addr{
			Uri: &sip.URI{
				Scheme: "sip",
				User:   user,
				Host:   c.options.Local.Host,
				Port:   uint16(c.options.Local.Port),
			},
		},
		CallID:     fmt.Sprintf("%s@%s", util.GenerateCallID(), c.options.Local.Host),
		CSeq:       util.GenerateCSeq(),
		CSeqMethod: sip.MethodInvite,
		UserAgent:  "IP Camera",
		Subject:    fmt.Sprintf("%s:%d,%s:%d", device, 0, user, 0),
		Payload: &sdp.SDP{
			Origin: sdp.Origin{
				User:    device,
				ID:      "0",
				Version: "0",
				Addr:    "192.168.118.62",
			},
			Addr:  "118.25.168.40",
			Audio: nil,
			Video: &sdp.Media{
				Proto: "UDP/RTP/AVP",
				Port:  10000,
				Codecs: []sdp.Codec{
					{
						PT:   96,
						Name: "PS",
						Rate: 90000,
					},
					{
						PT:   97,
						Name: "MPEG4",
						Rate: 90000,
					},
					{
						PT:   98,
						Name: "H264",
						Rate: 90000,
					},
				},
			},
			Session:  "Play",
			Time:     "0 0",
			RecvOnly: true,
		},
	}

	c.Send(msg)
}

func handler(c *client, _ *sip.Msg) {
	msg := &sip.Msg{
		Method: sip.MethodAck,
		Request: &sip.URI{
			Scheme: "sip",
			User:   device,
			Host:   domain,
		},
		From: &sip.Addr{
			Uri: &sip.URI{
				Scheme: "sip",
				User:   user,
				Host:   domain,
			},
			Param: &sip.Param{Name: "tag", Value: util.GenerateTag()},
		},
		To: &sip.Addr{
			Uri: &sip.URI{
				Scheme: "sip",
				User:   device,
				Host:   domain,
			},
		},
		Via: &sip.Via{
			Host: c.options.Local.Host,
			Port: uint16(c.options.Local.Port),
		},
		Contact: &sip.Addr{
			Uri: &sip.URI{
				Scheme: "sip",
				User:   user,
				Host:   c.options.Local.Host,
				Port:   uint16(c.options.Local.Port),
			},
		},
		CallID:     fmt.Sprintf("%s@%s", util.GenerateCallID(), c.options.Local.Host),
		CSeq:       1,
		CSeqMethod: sip.MethodAck,
		UserAgent:  "IP Camera",
	}

	c.Send(msg)
}
