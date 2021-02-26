package server

import (
	"bytes"
	"github.com/panjf2000/gnet"
	"gogs/api/data"
	dataGen "gogs/impl/data"
	"gogs/impl/logger"
	pk "gogs/impl/net/packet"
	"gogs/impl/net/packet/clientbound"
	"gogs/impl/net/packet/serverbound"
	"math"
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
	// if chunk border was crossed, update view pos and send new chunks
	chunkX := int(in.X) / 16
	chunkZ := int(in.Z) / 16
	if int(pos.X)/16 != chunkX || int(pos.Z)/16 != chunkZ {
		buf := bytes.Buffer{}
		buf.Write(clientbound.UpdateViewPosition{
			ChunkX: pk.VarInt(chunkX),
			ChunkZ: pk.VarInt(chunkZ),
		}.CreatePacket().Encode())

		// TODO: Add unload chunks for chunks out of range? not sure if needed

		biomes := make([]pk.VarInt, 1024, 1024)
		for i := range biomes {
			biomes[i] = 1
		}

		// TODO: change chunks sent to be based on client side render distance
		// TODO: optimize: just load the new chunks in the distance instead of sending all chunks nearby
		for x := -6; x < 6; x++ {
			for z := -6; z < 6; z++ {
				column, _ := s.world.GetChunk(x+chunkX, z+chunkZ)

				if column == nil {
					chunkDataArray := clientbound.ChunkDataArray{
						clientbound.ChunkSection{
							BlockCount:   0,
							BitsPerBlock: 4,
							Palette: clientbound.ChunkPalette{
								Length:  1,
								Palette: []pk.VarInt{0},
							},
							DataArrayLength: pk.VarInt(256),
							DataArray:       make([]pk.Long, 256),
						},
					}
					chunk := clientbound.ChunkData{
						ChunkX:         pk.Int(chunkX + x),
						ChunkZ:         pk.Int(chunkZ + z),
						FullChunk:      true,
						PrimaryBitMask: 1,
						Heightmaps: pk.NBT{
							V: clientbound.Heightmap{
								MotionBlocking: make([]int64, 37),
								WorldSurface:   make([]int64, 37),
							},
						},
						BiomesLength:     1024,
						Biomes:           biomes,
						Size:             pk.VarInt(len(chunkDataArray.Encode())),
						Data:             chunkDataArray,
						NumBlockEntities: 0,
						BlockEntities:    nil,
					}.CreatePacket().Encode()
					buf.Write(chunk)
				} else {
					chunkDataArray := make(clientbound.ChunkDataArray, len(column.Level.Sections)-1)
					bitMask := 0

					for i, section := range column.Level.Sections[1:] {
						bitsPerBlock := int64(math.Ceil(math.Log2(float64(len(section.Palette)))))
						if bitsPerBlock < 4 {
							bitsPerBlock = 4
						}
						bitMask |= 1 << section.Y
						blockData := make([]pk.Long, len(section.BlockStates))
						palette := make([]pk.VarInt, len(section.Palette))

						for i, block := range section.Palette {
							palette[i] = pk.VarInt(dataGen.ParseBlockId(block.Name))
						}

						for i, blockState := range section.BlockStates {
							blockData[i] = pk.Long(blockState)
						}
						chunkDataArray[i] = clientbound.ChunkSection{
							BlockCount:   4096,
							BitsPerBlock: pk.UByte(bitsPerBlock),
							Palette: clientbound.ChunkPalette{
								Length:  pk.VarInt(len(palette)),
								Palette: palette,
							},
							DataArrayLength: pk.VarInt(len(blockData)),
							DataArray:       blockData,
						}
					}

					chunk := clientbound.ChunkData{
						ChunkX:         pk.Int(x + chunkX),
						ChunkZ:         pk.Int(z + chunkZ),
						FullChunk:      true,
						PrimaryBitMask: pk.VarInt(bitMask),
						Heightmaps: pk.NBT{
							V: clientbound.Heightmap{
								MotionBlocking: make([]int64, 37),
								WorldSurface:   make([]int64, 37),
							},
						},
						BiomesLength:     1024,
						Biomes:           biomes,
						Size:             pk.VarInt(len(chunkDataArray.Encode())),
						Data:             chunkDataArray,
						NumBlockEntities: 0,
						BlockEntities:    nil,
					}.CreatePacket().Encode()
					buf.Write(chunk)
				}
			}
		}

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
