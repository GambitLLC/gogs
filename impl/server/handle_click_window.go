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

	logger.Printf("%v", in)

	var windowSlotChanged bool
	var heldSlotChanged bool

	// TODO: support other windows
	window := player.Inventory

	switch in.Mode {
	case 0: // Normal left & right click
		switch in.Button {
		case 0: // Normal left click
			if slot == -999 {
				// TODO: drop held item
			} else if player.HeldSlot.Present && window[slot].ItemID == player.HeldSlot.ItemID {
				sum := window[slot].ItemCount + player.HeldSlot.ItemCount
				if sum > 64 {
					window[slot].ItemCount = 64
					player.HeldSlot.ItemCount = sum - 64
				} else {
					window[slot].ItemCount = sum
					player.HeldSlot.ItemCount = 0
					player.HeldSlot.Present = false
				}
				windowSlotChanged = true
				heldSlotChanged = true

			} else {
				window[slot], player.HeldSlot = player.HeldSlot, window[slot]
				windowSlotChanged = true
				heldSlotChanged = true
			}
		case 1: // Normal right click
			if slot == -999 {
				// TODO: drop held item
			} else if !player.HeldSlot.Present {
				// pick up (greater) half of the stack
				player.HeldSlot.Present = true
				player.HeldSlot.ItemID = window[slot].ItemID
				player.HeldSlot.ItemCount = window[slot].ItemCount/2 + (window[slot].ItemCount & 1)
				window[slot].ItemCount /= 2
				windowSlotChanged = true
				heldSlotChanged = true
			} else if player.HeldSlot.ItemCount > 1 && (window[slot].ItemID|player.HeldSlot.ItemID == 0) {
				// put a single item into the selected slot (air or same item type)
			}
		}
	}

	buf := bytes.Buffer{}
	if windowSlotChanged {
		buf.Write(clientbound.SetSlot{
			WindowID: 0,
			Slot:     in.Slot,
			SlotData: window[slot],
		}.CreatePacket().Encode())
	}
	if heldSlotChanged {
		buf.Write(clientbound.SetSlot{
			WindowID: -1,
			Slot:     -1,
			SlotData: player.HeldSlot,
		}.CreatePacket().Encode())
	}

	out = buf.Bytes()

	return
}
