package server

import (
	"bytes"
	"github.com/panjf2000/gnet"
	"gogs/impl/logger"
	pk "gogs/impl/net/packet"
	"gogs/impl/net/packet/clientbound"
	"gogs/impl/net/packet/packetids"
	"gogs/impl/net/packet/serverbound"
)

func (s *Server) handleClickWindow(conn gnet.Conn, pkt pk.Packet) (out []byte, err error) {
	in := serverbound.ClickWindow{}
	if err = in.FromPacket(pkt); err != nil {
		return
	}

	player := s.playerFromConn(conn)
	slot := int(in.Slot)

	logger.Printf("%v", in)

	if slot == -1 {
		// clicking on border of window, ignore
		return
	}

	var rejected bool

	// TODO: support other windows
	window := &player.InventoryComponent

	window.InventoryLock.Lock()
	defer window.InventoryLock.Unlock()
	player.HeldSlotLock.Lock()
	defer player.HeldSlotLock.Unlock()

	if !rejected {
		switch in.Mode {
		default: // not implemented/supported
			rejected = true
		case 0: // Normal left & right click
			switch in.Button {
			case 0: // Normal left click
				if slot == -999 {
					// TODO: drop held item
					rejected = true
					break
				}

				if window.Inventory[slot] != in.ClickedItem {
					rejected = true
					break
				}

				if player.HeldSlot.Present && window.Inventory[slot].ItemID == player.HeldSlot.ItemID {
					placeStack(&player.HeldSlot, &window.Inventory[slot])
				} else {
					// swap items
					window.Inventory[slot], player.HeldSlot = player.HeldSlot, window.Inventory[slot]
				}
			case 1: // Normal right click
				if slot == -999 {
					// TODO: drop held item
					rejected = true
					break
				}

				if window.Inventory[slot] != in.ClickedItem {
					rejected = true
					break
				}

				if !player.HeldSlot.Present {
					player.HeldSlot = halveStack(&window.Inventory[slot])
				} else if window.Inventory[slot].ItemID == player.HeldSlot.ItemID || window.Inventory[slot].ItemID == 0 {
					one := takeOne(&player.HeldSlot)
					if window.Inventory[slot].Present {
						window.Inventory[slot].ItemCount += 1
					} else {
						window.Inventory[slot] = one
					}
				} else {
					// swap items
					window.Inventory[slot], player.HeldSlot = player.HeldSlot, window.Inventory[slot]
				}
			}
		case 1: // shift left/right click
			rejected = true
		case 2: // number keys
			// clicked item is always empty for number keys, don't need to compare to window

			if in.Button < 0 || in.Button > 8 {
				rejected = true // shouldn't ever occur, but check in case ...
				break
			}

			hotBarSlot := uint8(in.Button)

			if in.WindowID != 0 {
				player.InventoryLock.Lock()
				defer player.InventoryLock.Unlock()
			}

			// swap selected item with hot bar slot
			window.Inventory[slot], player.Inventory[hotBarSlot+36] = player.Inventory[hotBarSlot+36], window.Inventory[slot]
		case 3: // middle click (only defined for creative in other windows)
			rejected = true
		case 4: // drop key or no-op
			rejected = true
		case 5: // drag mode
			rejected = true
		case 6: // double click
			rejected = true
		}
	}

	buf := bytes.Buffer{}
	buf.Write(pk.Marshal(
		packetids.WindowConfirmationClientbound,
		in.WindowID,
		in.ActionNumber,
		pk.Boolean(!rejected), // accepted
	).Encode())

	if rejected {
		buf.Write(clientbound.WindowItems{
			WindowID: in.WindowID,
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
		buf.Write(clientbound.SetSlot{
			WindowID: -1,
			Slot:     -1,
			SlotData: player.HeldSlot,
		}.CreatePacket().Encode())
	}

	out = buf.Bytes()
	return
}

func placeStack(from *pk.Slot, onto *pk.Slot) {
	sum := from.ItemCount + onto.ItemCount
	if sum > 64 {
		onto.ItemCount = 64
		from.ItemCount = sum - 64
	} else {
		onto.ItemCount = sum
		from.ItemCount = 0
		from.ItemID = 0
		from.Present = false
	}
}

// halveStack returns the greater half of the stack
func halveStack(stack *pk.Slot) pk.Slot {
	half := pk.Slot{
		Present:   true,
		ItemID:    stack.ItemID,
		ItemCount: stack.ItemCount/2 + (stack.ItemCount & 1),
		NBT:       stack.NBT,
	}

	stack.ItemCount /= 2
	if stack.ItemCount == 0 {
		stack.ItemID = 0
		stack.Present = false
	}

	return half
}

// takeOne takes a single item from the stack and returns it
func takeOne(stack *pk.Slot) pk.Slot {
	one := pk.Slot{
		Present:   true,
		ItemID:    stack.ItemID,
		ItemCount: 1,
		NBT:       stack.NBT,
	}

	stack.ItemCount -= 1
	if stack.ItemCount == 0 {
		stack.Present = false
		stack.ItemID = 0
	}

	return one
}
