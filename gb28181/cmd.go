package gb28181

import (
	"strconv"
	"strings"

	"github.com/lilith44/easy"
)

const (
	PtzZoomOut = 5
	PtzZoomIn  = 4
	PtzUp      = 3
	PtzDown    = 2
	PtzLeft    = 1
	PtzRight   = 0

	FiIrisReduce = 3
	FiIrisExpand = 2
	FiFocusNear  = 1
	FiFocusAway  = 0
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

type (
	FiIrisParam struct {
		// 放大为1 缩小为-1
		Direction int `json:"direction" validate:"oneof=-1 0 1"`
		Speed     int `json:"speed" validate:"min=0,max=255"`
	}

	FiFocusParam struct {
		// 聚焦近为1 聚焦远为-1
		Direction int `json:"direction" validate:"oneof=-1 0 1"`
		Speed     int `json:"speed" validate:"min=0,max=255"`
	}

	FiConfig struct {
		Iris  *FiIrisParam  `json:"iris"`
		Focus *FiFocusParam `json:"focus"`
	}
)

func NewPtzCmd(config PtzConfig) Cmd {
	cmd := newCmd()

	if config.Zoom != nil {
		if config.Zoom.Direction == -1 {
			cmd[3] += 1 << PtzZoomIn
		}
		if config.Zoom.Direction == 1 {
			cmd[3] += 1 << PtzZoomOut
		}

		cmd[6] = config.Zoom.Speed << 4
	}

	if config.Horizontal != nil {
		if config.Horizontal.Direction == -1 {
			cmd[3] += 1 << PtzLeft
		}
		if config.Horizontal.Direction == 1 {
			cmd[3] += 1 << PtzRight
		}

		cmd[4] = config.Horizontal.Speed
	}

	if config.Vertical != nil {
		if config.Vertical.Direction == -1 {
			cmd[3] += 1 << PtzDown
		}
		if config.Vertical.Direction == 1 {
			cmd[3] += 1 << PtzUp
		}

		cmd[5] = config.Vertical.Speed
	}

	return cmd
}

func NewFiCmd(config FiConfig) Cmd {
	cmd := newCmd()
	// 表示是fi指令
	cmd[3] += 1 << 6

	if config.Iris != nil {
		if config.Iris.Direction == -1 {
			cmd[3] += 1 << FiIrisReduce
		}
		if config.Iris.Direction == 1 {
			cmd[3] += 1 << FiIrisExpand
		}

		cmd[5] = config.Iris.Speed
	}

	if config.Focus != nil {
		if config.Focus.Direction == -1 {
			cmd[3] += 1 << FiFocusNear
		}
		if config.Focus.Direction == 1 {
			cmd[3] += 1 << FiFocusAway
		}

		cmd[4] = config.Focus.Speed
	}

	return cmd
}
