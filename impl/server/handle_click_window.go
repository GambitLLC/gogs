package server

import (
	"bytes"
	"github.com/panjf2000/gnet"
	"gogs/impl/logger"
	pk "gogs/impl/net/packet"
	"gogs/impl/net/packet/clientbound"
	"gogs/impl/net/packet/serverbound"
)

func (s *Server) handleClickWindow(conn gnet.Conn, pkt pk.Packet) (out []byte, err error) {
	in := serverbound.ClickWindow{}
	if err = in.FromPacket(pkt); err != nil {
		return
	}

	player := s.playerFromConn(conn)
	slot := int16(in.Slot)

	buf := bytes.Buffer{}

	logger.Printf("%v", in)

	switch in.Mode {
	case 0: // Normal left & right click
		switch in.Button {
		case 0: // Normal left click
			if slot == -999 {
				// TODO: drop held item
			} else {
				player.Inventory[slot], player.HeldSlot = player.HeldSlot, player.Inventory[slot]
				buf.Write(clientbound.SetSlot{
					WindowID: 0,
					Slot:     in.Slot,
					SlotData: player.Inventory[slot],
				}.CreatePacket().Encode())
				buf.Write(clientbound.SetSlot{
					WindowID: -1,
					Slot:     -1,
					SlotData: player.HeldSlot,
				}.CreatePacket().Encode())
			}
		case 1: // Normal right click
		}
	}

	out = buf.Bytes()

	return
}
