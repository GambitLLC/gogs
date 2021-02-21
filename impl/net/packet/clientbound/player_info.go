package clientbound

import (
	"bytes"
	pk "gogs/impl/net/packet"
	"gogs/impl/net/packet/packetids"
)

type PlayerInfo struct {
	Action     pk.VarInt
	NumPlayers pk.VarInt
	Players    players
}

func (s PlayerInfo) CreatePacket() pk.Packet {
	return pk.Marshal(packetids.PlayerInfo, s.Action, s.NumPlayers, s.Players)
}

type players []pk.Encodable

func (a players) Encode() []byte {
	buf := bytes.Buffer{}
	for _, v := range a {
		buf.Write(v.Encode())
	}
	return buf.Bytes()
}

type PlayerInfoAddPlayer struct {
	UUID           pk.UUID
	Name           pk.String // 16
	NumProperties  pk.VarInt
	Properties     Properties
	Gamemode       pk.VarInt
	Ping           pk.VarInt
	HasDisplayName pk.Boolean
	DisplayName    pk.Chat // Optional
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

type PlayerInfoRemovePlayer struct {
	UUID pk.UUID
}

func (s PlayerInfoRemovePlayer) Encode() []byte {
	return s.UUID.Encode()
}

type Property struct {
}

type Properties []Property

func (a Properties) Encode() []byte {
	return nil
}
