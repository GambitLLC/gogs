package chat

import (
	"encoding/json"
)

type Message struct {
	Text string `json:"text"`

	Bold          *bool `json:"bold,boolean,omitempty"`
	Italic        *bool `json:"italic,boolean,omitempty"`
	Underlined    *bool `json:"underlined,boolean,omitempty"`
	Strikethrough *bool `json:"strikethrough,boolean,omitempty"`
	Obfuscated    *bool `json:"obfuscated,boolean,omitempty"`

	Color int `json:"color,string,omitempty"`

	Extra []*Message `json:"extra,omitempty"`
}

func NewMessage(text string) Message {
	return Message{Text: text}
}

func (m Message) AsJSON() string {
	if text, err := json.Marshal(m); err != nil {
		panic(err)
	} else {
		return string(text)
	}
}

func (m *Message) Append(msgs ...Message) {
	length := len(m.Extra) + len(msgs)
	// expand array to make appending faster
	extra := make([]*Message, length)
	copy(extra, m.Extra)
	m.Extra = extra

	for _, v := range msgs {
		m.Extra = append(m.Extra, &v)
	}
}
