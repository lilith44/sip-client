package sip_client

const ContentTypeXML = "Application/MANSCDP+xml"

type XMLPayload struct {
	payload []byte
}

func NewXMLPayload(payload []byte) *XMLPayload {
	return &XMLPayload{payload: payload}
}

func (x *XMLPayload) Data() []byte {
	return x.payload
}

func (x *XMLPayload) ContentType() string {
	return ContentTypeXML
}
