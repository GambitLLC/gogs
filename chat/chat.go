package chat

import (
	"encoding/json"
)

const (
	Chat     = 0
	System   = 1
	GameInfo = 2
)

type Message struct {
	// String component
	Text string `json:"text,omitempty"`

	// Translation component
	Translate string    `json:"translate,omitempty"`
	With      []Message `json:"with,omitempty"`

	// Common
	Bold          bool      `json:"bold,boolean,omitempty"`
	Italic        bool      `json:"italic,boolean,omitempty"`
	Underlined    bool      `json:"underlined,boolean,omitempty"`
	Strikethrough bool      `json:"strikethrough,boolean,omitempty"`
	Obfuscated    bool      `json:"obfuscated,boolean,omitempty"`
	Color         string    `json:"color,omitempty"`
	Extra         []Message `json:"extra,omitempty"`
}

func NewStringComponent(text string) Message {
	return Message{Text: text}
}

func NewTranslationComponent(key string, with ...Message) Message {
	m := Message{Translate: key}
	m.With = make([]Message, len(with))
	copy(m.With, with)
	return m
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
	extra := make([]Message, length)
	copy(extra, m.Extra)
	copy(extra[len(m.Extra):], msgs)
	m.Extra = extra
}
