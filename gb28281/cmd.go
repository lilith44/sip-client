package gb28281

import (
	"strconv"
	"strings"

	"github.com/lilith44/easy"
)

const (
	ZoomOut = 5
	ZoomIn  = 4
	Up      = 3
	Down    = 2
	Left    = 1
	Right   = 0
)

type Cmd []int

func newCmd() Cmd {
	cmd := make(Cmd, 8)
	cmd[0] = 165 // A5
	cmd[1] = 15  // 0F
	cmd[2] = 01  // 01
	return cmd
}

func (c Cmd) SetCheckBit() {
	c[7] += easy.Sum(c[:7]...) % (1 << 8)
}

func (c Cmd) String() string {
	c.SetCheckBit()

	b := strings.Builder{}
	b.Grow(2 * len(c))
	for i := range c {
		hex := strconv.FormatInt(int64(c[i]), 16)
		if len(hex) == 1 {
			b.WriteByte('0')
		}
		b.WriteString(hex)
	}
	return strings.ToUpper(b.String())
}

type (
	PtzRotateParam struct {
		// 左、下为-1，右、上为1
		Direction int `json:"direction" validate:"oneof=-1 0 1"`
		Speed     int `json:"speed" validate:"min=0,max=255"`
	}

	PtzZoomParam struct {
		// -1为放大，1为缩小
		Direction int `json:"direction" validate:"oneof=-1 0 1"`
		Speed     int `json:"speed" validate:"min=0,max=15"`
	}

	PtzConfig struct {
		Zoom       *PtzZoomParam   `json:"zoom"`
		Horizontal *PtzRotateParam `json:"horizontal"`
		Vertical   *PtzRotateParam `json:"vertical"`
	}
)

func NewPtzCmd(config PtzConfig) Cmd {
	cmd := newCmd()

	if config.Zoom != nil {
		if config.Zoom.Direction == -1 {
			cmd[3] += 1 << ZoomIn
		}
		if config.Zoom.Direction == 1 {
			cmd[3] += 1 << ZoomOut
		}

		cmd[6] = config.Zoom.Speed << 4
	}

	if config.Horizontal != nil {
		if config.Horizontal.Direction == -1 {
			cmd[3] += 1 << Left
		}
		if config.Horizontal.Direction == 1 {
			cmd[3] += 1 << Right
		}

		cmd[4] = config.Horizontal.Speed
	}

	if config.Vertical != nil {
		if config.Vertical.Direction == -1 {
			cmd[3] += 1 << Down
		}
		if config.Vertical.Direction == 1 {
			cmd[3] += 1 << Up
		}

		cmd[5] = config.Vertical.Speed
	}

	return cmd
}
