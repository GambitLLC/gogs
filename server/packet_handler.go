package server

import (
	"fmt"

	"github.com/GambitLLC/gogs/entities"
	"github.com/GambitLLC/gogs/net"
	pk "github.com/GambitLLC/gogs/net/packet"
	"github.com/GambitLLC/gogs/net/packet/clientbound"
	"github.com/GambitLLC/gogs/net/packet/packetids"
)

func (s *Server) onLoginStart(conn net.Conn, pkt pk.Packet) error {
	var name pk.String
	if err := pkt.Unmarshal(&name); err != nil {
		return err
	}

	// TODO: handle encryption (send encryption start)

	u := pk.NameToUUID(string(name)) // todo: get uuid from mojang servers
	if err := conn.WritePacket(pk.Marshal(
		packetids.LoginSuccess,
		pk.UUID(u),
		name,
	)); err != nil {
		return err
	}

	player := s.createPlayer(string(name), u, conn)
	if err := conn.WritePacket(s.joinGamePacket(player)); err != nil {
		return err
	}

	s.Broadcast(fmt.Sprintf("%v has joined the game", name))

	// send out new player info to everyone already online
	s.broadcastPacket(clientbound.PlayerInfo{
		Action:     0,
		NumPlayers: 1,
		Players: []pk.Encodable{
			clientbound.PlayerInfoAddPlayer{
				UUID:           pk.UUID(player.UUID),
				Name:           pk.String(player.Name),
				NumProperties:  0,
				Properties:     nil,
				Gamemode:       pk.VarInt(player.GameMode),
				Ping:           1,
				HasDisplayName: false,
				DisplayName:    "",
			},
		},
	}.CreatePacket(), conn)

	// TODO: spawn player should be occurring when players enter range (not join game), do logic elsewhere (tick?)
	s.broadcastPacket(clientbound.SpawnPlayer{
		EntityID:   pk.VarInt(player.ID()),
		PlayerUUID: pk.UUID(player.UUID),
		X:          pk.Double(player.X),
		Y:          pk.Double(player.Y),
		Z:          pk.Double(player.Z),
		Yaw:        pk.Angle(player.Yaw / 360 * 256),
		Pitch:      pk.Angle(player.Pitch / 360 * 256),
	}.CreatePacket(), conn)

	return nil
}

func (s *Server) joinGamePacket(player *entities.Player) pk.Packet {
	return clientbound.JoinGame{
		EntityID:       pk.Int(player.ID()),
		IsHardcore:     false,
		Gamemode:       pk.UByte(player.GameMode),
		PrevGamemode:   -1,
		WorldCount:     1,
		WorldNames:     []pk.Identifier{"world"},
		DimensionCodec: pk.NBT{V: clientbound.MinecraftDimensionCodec},
		Dimension:      pk.NBT{V: clientbound.MinecraftOverworld},
		WorldName:      "world",
		HashedSeed:     0,
		MaxPlayers:     pk.VarInt(s.MaxPlayers),
		ViewDistance:   pk.VarInt(s.ViewDistance),
		RDI:            false,
		ERS:            false,
		IsDebug:        false,
		IsFlat:         true,
	}.CreatePacket()
}
