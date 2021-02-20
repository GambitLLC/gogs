package clientbound

import (
	"bytes"
	pk "gogs/impl/net/packet"
)

type PlayerInfo struct {
	Action     pk.VarInt
	NumPlayers pk.VarInt
	Players     []pk.Encodable
}

func (s PlayerInfo) Encode() []byte {
	buf := bytes.Buffer{}
	buf.Write(s.Action.Encode())
	buf.Write(s.NumPlayers.Encode())
	for _, v := range s.Players {
		buf.Write(v.Encode())
	}
	return buf.Bytes()
}

type PlayerInfoAddPlayer struct {
	UUID pk.UUID
	Name pk.String	// 16
	NumProperties pk.VarInt
	Properties Properties
	Gamemode pk.VarInt
	Ping pk.VarInt
	HasDisplayName pk.Boolean
	DisplayName pk.Chat	// Optional
}

type Property struct {
}

type Properties []Property

func (a Properties) Encode() []byte {
	return nil
}

func (s PlayerInfoAddPlayer) Encode() []byte {
	buf := bytes.Buffer{}
	buf.Write(s.UUID.Encode())
	buf.Write(s.Name.Encode())
	buf.Write(s.NumProperties.Encode())
	if s.NumProperties > 0 {
		buf.Write(s.Properties.Encode())
	}
	buf.Write(s.Gamemode.Encode())
	buf.Write(s.Ping.Encode())
	buf.Write(s.HasDisplayName.Encode())
	if s.HasDisplayName {
		buf.Write(s.DisplayName.Encode())
	}
	return buf.Bytes()
}