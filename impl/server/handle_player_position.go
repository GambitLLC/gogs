package server

import (
	"bytes"
	"github.com/panjf2000/gnet"
	"gogs/api/data"
	"gogs/impl/logger"
	pk "gogs/impl/net/packet"
	"gogs/impl/net/packet/clientbound"
	"gogs/impl/net/packet/serverbound"
)

func (s *Server) handlePlayerPosition(conn gnet.Conn, pkt pk.Packet) (out []byte, err error) {
	player := s.playerFromConn(conn)
	logger.Printf("Received player position for %v", player.Name())
	in := serverbound.PlayerPosition{}
	if err = in.FromPacket(pkt); err != nil {
		return
	}

	outPacket := clientbound.EntityPosition{
		EntityID: pk.VarInt(player.EntityID()),
		DeltaX:   pk.Short((float64(in.X*32) - player.Position().X*32) * 128),
		DeltaY:   pk.Short((float64(in.Y*32) - player.Position().Y*32) * 128),
		DeltaZ:   pk.Short((float64(in.Z*32) - player.Position().Z*32) * 128),
		OnGround: in.OnGround,
	}.CreatePacket()

	s.broadcastPacket(outPacket, conn)

	// update player position
	pos := player.Position()

	// TODO: according to wikivg, update view position is sent on change in Y coord as well
	// TODO: update in player position rotation as well (change where this logic occurs ...)
	// if chunk border was crossed, update view pos and send new chunks
	chunkX := int(in.X) >> 4
	chunkZ := int(in.Z) >> 4
	if int(pos.X)>>4 != chunkX || int(pos.Z)>>4 != chunkZ {
		buf := bytes.Buffer{}
		buf.Write(clientbound.UpdateViewPosition{
			ChunkX: pk.VarInt(chunkX),
			ChunkZ: pk.VarInt(chunkZ),
		}.CreatePacket().Encode())

		// TODO: Add unload chunks for chunks out of range? not sure if needed
		// TODO: optimize: just load the new chunks in the distance instead of sending all chunks nearby
		buf.Write(s.chunkDataPackets(player))

		out = buf.Bytes()
	}

	// update player position
	*pos = data.Position{
		X:        float64(in.X),
		Y:        float64(in.Y),
		Z:        float64(in.Z),
		OnGround: bool(in.OnGround),
	}

	return
}
