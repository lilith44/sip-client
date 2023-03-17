package main

import "encoding/xml"

type Message struct {
	payload any
}

func NewMessage(payload any) *Message {
	return &Message{payload: payload}
}

func (m *Message) Data() []byte {
	switch p := m.payload.(type) {
	case []byte:
		return p
	default:
		data, _ := xml.Marshal(m.payload)
		return data
	}
}

func (m *Message) ContentType() string {
	return "Application/MANSCDP+xml"
}
