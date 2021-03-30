package server

import (
	"gogs/impl/ecs"
	"gogs/impl/logger"
	"gogs/impl/net"
	pk "gogs/impl/net/packet"
	"gogs/impl/net/packet/clientbound"
	"gogs/impl/net/packet/serverbound"
)

func (s *Server) handlePlayerPosition(conn net.Conn, pkt pk.Packet) (err error) {
	player := s.playerFromConn(conn)
	logger.Printf("Received player position for %v", player.Name)
	in := serverbound.PlayerPosition{}
	if err = in.FromPacket(pkt); err != nil {
		return
	}

	outPacket := clientbound.EntityPosition{
		EntityID: pk.VarInt(player.ID()),
		DeltaX:   pk.Short((float64(in.X*32) - player.X*32) * 128),
		DeltaY:   pk.Short((float64(in.Y*32) - player.Y*32) * 128),
		DeltaZ:   pk.Short((float64(in.Z*32) - player.Z*32) * 128),
		OnGround: in.OnGround,
	}.CreatePacket()

	s.broadcastPacket(outPacket, conn)

	// TODO: according to wikivg, update view position is sent on change in Y coord as well
	// if chunk border was crossed, update view pos and send new chunks
	if int(player.X)>>4 != int(in.X)>>4 || int(player.Z)>>4 != int(in.Z)>>4 {
		err = s.updateViewPosition(player)
	}

	// update player position
	player.X = float64(in.X)
	player.Y = float64(in.Y)
	player.Z = float64(in.Z)

	return
}

func (s *Server) updateViewPosition(player *ecs.Player) (err error) {
	if err = player.Connection.WritePacket(clientbound.UpdateViewPosition{
		ChunkX: pk.VarInt(int(player.X) >> 4),
		ChunkZ: pk.VarInt(int(player.Z) >> 4),
	}.CreatePacket()); err != nil {
		return
	}

	chunkX := int(player.X) >> 4
	chunkZ := int(player.Z) >> 4

	viewDistance := int(player.ViewDistance)
	for x := -viewDistance; x <= viewDistance; x++ {
		for z := -viewDistance; z <= viewDistance; z++ {
			if err = player.Connection.WritePacket(s.chunkDataPacket(chunkX+x, chunkZ+z)); err != nil {
				return
			}
		}
	}

	// TODO: only send new chunks and unload old chunks

	return nil
}
