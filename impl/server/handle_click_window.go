package server

import (
	"bytes"
	"github.com/panjf2000/gnet"
	"gogs/impl/logger"
	pk "gogs/impl/net/packet"
	"gogs/impl/net/packet/clientbound"
	"gogs/impl/net/packet/packetids"
	"gogs/impl/net/packet/serverbound"
	"log"
)

func (s *Server) handleClickWindow(conn gnet.Conn, pkt pk.Packet) (out []byte, err error) {
	in := serverbound.ClickWindow{}
	if err = in.FromPacket(pkt); err != nil {
		return
	}

	player := s.playerFromConn(conn)
	slot := int(in.Slot)

	logger.Printf("%v", in)

	var rejected, windowSlotChanged, heldSlotChanged bool

	// TODO: support other windows
	window := &player.InventoryComponent
	heldSlot := player.HeldSlot

	window.InventoryLock.RLock()
	// TODO: this check doesn't work for all modes ... move or change this
	if slot >= 0 && slot < len(window.Inventory) && window.Inventory[slot] != in.ClickedItem {
		log.Printf("rejected: %v, %v", window.Inventory[slot], in.ClickedItem)
		rejected = true
	}
	window.InventoryLock.RUnlock()

	if !rejected {
		switch in.Mode {
		default: // not implemented/supported
			rejected = true
		case 0: // Normal left & right click
			switch in.Button {
			case 0: // Normal left click
				if slot == -999 {
					// TODO: drop held item
				} else if heldSlot.Present && in.ClickedItem.ItemID == heldSlot.ItemID {
					// combine item stacks
					sum := in.ClickedItem.ItemCount + heldSlot.ItemCount
					if sum > 64 {
						in.ClickedItem.ItemCount = 64
						heldSlot.ItemCount = sum - 64
					} else {
						in.ClickedItem.ItemCount = sum
						heldSlot.ItemCount = 0
						heldSlot.ItemID = 0
						heldSlot.Present = false
					}
					windowSlotChanged = true
					heldSlotChanged = true
				} else {
					// swap items
					in.ClickedItem, heldSlot = heldSlot, in.ClickedItem
					windowSlotChanged = true
					heldSlotChanged = true
				}
			case 1: // Normal right click
				if slot == -999 {
					// TODO: drop held item
				} else if !heldSlot.Present {
					// pick up (greater) half of the stack
					heldSlot.Present = true
					heldSlot.ItemID = in.ClickedItem.ItemID
					heldSlot.ItemCount = in.ClickedItem.ItemCount/2 + (in.ClickedItem.ItemCount & 1)
					in.ClickedItem.ItemCount /= 2
					if in.ClickedItem.ItemCount == 0 {
						in.ClickedItem.ItemID = 0
						in.ClickedItem.Present = false
					}
					windowSlotChanged = true
					heldSlotChanged = true
				} else if in.ClickedItem.ItemID == heldSlot.ItemID || in.ClickedItem.ItemID == 0 {
					// put a single item into the selected slot (air or same item type)
					heldSlot.ItemCount -= 1
					if heldSlot.ItemCount == 0 {
						heldSlot.Present = false
						heldSlot.ItemID = 0
					}
					in.ClickedItem.Present = true
					in.ClickedItem.ItemID = heldSlot.ItemID
					in.ClickedItem.ItemCount += 1
					windowSlotChanged = true
					heldSlotChanged = true
				}
			}
		}
	}

	buf := bytes.Buffer{}
	if rejected {
		window.InventoryLock.RLock()
		buf.Write(clientbound.WindowItems{
			WindowID: 0,
			Count:    pk.Short(len(window.Inventory)),
			SlotData: window.Inventory,
		}.CreatePacket().Encode())
		if slot != -999 {
			buf.Write(clientbound.SetSlot{
				WindowID: 0,
				Slot:     in.Slot,
				SlotData: window.Inventory[slot],
			}.CreatePacket().Encode())
		}
		window.InventoryLock.RUnlock()
		buf.Write(clientbound.SetSlot{
			WindowID: -1,
			Slot:     -1,
			SlotData: player.HeldSlot,
		}.CreatePacket().Encode())
	} else {
		if windowSlotChanged {
			window.InventoryLock.Lock()
			window.Inventory[slot] = in.ClickedItem
			window.InventoryLock.Unlock()
		}
		if heldSlotChanged {
			player.HeldSlot = heldSlot
		}
	}

	// TODO: create WindowConfirmation packet struct
	buf.Write(pk.Marshal(
		packetids.WindowConfirmationClientbound,
		pk.Byte(0), // window id
		in.ActionNumber,
		pk.Boolean(!rejected),
	).Encode())

	out = buf.Bytes()
	return
}
