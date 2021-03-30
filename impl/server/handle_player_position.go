package server

import (
	"gogs/impl/ecs"
	"gogs/impl/logger"
	"gogs/impl/net"
	pk "gogs/impl/net/packet"
	"gogs/impl/net/packet/clientbound"
	"gogs/impl/net/packet/serverbound"
	"time"
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

	prevChunks := make(ecs.ChunkSet)
	for x, xMap := range player.KnownChunks {
		for z := range xMap {
			prevChunks.Add(x, z)
		}
	}

	newChunks := make(ecs.ChunkSet)

	viewDistance := int(player.ViewDistance)
	for x := chunkX - viewDistance; x <= chunkX+viewDistance; x++ {
		for z := chunkZ - viewDistance; z <= chunkZ+viewDistance; z++ {
			if player.KnownChunks.Contains(x, z) {
				prevChunks.Remove(x, z)
			} else {
				newChunks.Add(x, z)
				player.KnownChunks.Add(x, z)
			}
		}
	}

	// send data for the new chunks
	tick := time.Tick(10 * time.Millisecond)
	for x, xMap := range newChunks {
		for z := range xMap {
			// slow down chunk packets: sending too many too fast causes client to lag due to rendering?
			// TODO: determine if issue is something else or if there's another way to fix this
			select {
			case <-tick:
				if err = player.Connection.WritePacket(s.chunkDataPacket(x, z)); err != nil {
					return
				}
			}
		}
	}

	// remove old chunks and send unload chunk packet
	for x, xMap := range prevChunks {
		for z := range xMap {
			player.KnownChunks.Remove(x, z)
			unload := clientbound.UnloadChunk{
				ChunkX: pk.Int(x),
				ChunkZ: pk.Int(z),
			}
			if err = player.Connection.WritePacket(unload.CreatePacket()); err != nil {
				return
			}
		}
	}

	return nil
}
