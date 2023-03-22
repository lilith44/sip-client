package gb28181

import "encoding/xml"

// Message 收到的xml消息
type Message struct {
	XMLName xml.Name

	CmdType  string `xml:"CmdType"`
	SN       int    `xml:"SN"`
	DeviceID string `xml:"DeviceID"`
	Status   string `xml:"Status"`
}
