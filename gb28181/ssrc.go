package gb28181

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// SSRC RTP里的SSRC
type SSRC struct {
	// 域ID，例如13010000002000000001
	deviceID string
	// 流id，需要唯一，[1, 9999]以内
	streamId int
	// 是否是实时流
	isRealTime bool
}

func NewSSRC(isRealTime bool, deviceID string, streamId int) (*SSRC, error) {
	if streamId <= 0 || streamId > 9999 {
		return nil, errors.New(fmt.Sprintf("streamId必须在[1, 9999]以内"))
	}

	return &SSRC{
		deviceID:   deviceID,
		streamId:   streamId,
		isRealTime: isRealTime,
	}, nil
}

func (s *SSRC) String() string {
	b := strings.Builder{}
	if s.isRealTime {
		b.WriteByte('0')
	} else {
		b.WriteByte('1')
	}

	b.WriteString(s.deviceID[3:8])
	stream := strconv.FormatInt(int64(s.streamId), 10)
	if len(stream) < 4 {
		b.WriteString(strings.Repeat("0", 4-len(stream)))
	}
	b.WriteString(stream)
	return b.String()
}
