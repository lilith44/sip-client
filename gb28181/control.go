package gb28181

import (
	"fmt"
	"math"
	"math/rand"
	"time"
)

var controlOrder = `<? xml version="1.0"?>
<Control>
<CmdType>DeviceControl</CmdType>
<SN>%d</SN>
<DeviceID>%s</DeviceID>
<PTZCmd>%s</PTZCmd>
<Info>
<ControlPriority>5</ControlPriority>
</Info>
</Control>
`

// NewControl 创建一个Control命令
func NewControl(deviceID string, cmd Cmd) []byte {
	return []byte(fmt.Sprintf(controlOrder, randInt(), deviceID, cmd.String()))
}

func randInt() int {
	rand.Seed(time.Now().Unix())
	return rand.Intn(math.MaxInt16)
}
