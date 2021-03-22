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

	// painting status is reset if any packets come in that aren't painting mode (5)
	if in.Mode != 5 && player.PaintingSlots != nil {
		player.PaintingSlots = nil
	}

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
				one := takeAmt(&player.HeldSlot, 1)
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
		// no check between clickedItem and window inventory b/c clicked item is always empty
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
		hotBar := player.Inventory[36:45]
		window.Inventory[slot], hotBar[hotBarSlot] = hotBar[hotBarSlot], window.Inventory[slot]
	case 3: // middle click (only defined for creative in other windows)
		rejected = true
	case 4: // drop key or no-op
		rejected = true
	case 5: // painting (drag) mode
		switch in.Button {
		default:
			rejected = true
		case 0: // start left mouse drag
			player.PaintingLock.Lock()
			player.PaintingSlots = make([]uint8, 0, len(window.Inventory))
			player.PaintingLock.Unlock()
		case 1: // add slot for left mouse drag
			player.PaintingLock.Lock()
			defer player.PaintingLock.Unlock()
			if player.PaintingSlots == nil {
				rejected = true
				break
			}
			player.PaintingSlots = append(player.PaintingSlots, uint8(slot))
		case 2: // end left mouse drag
			player.PaintingLock.Lock()
			defer player.PaintingLock.Unlock()
			if player.PaintingSlots == nil || len(player.PaintingSlots) > int(player.HeldSlot.ItemCount) {
				rejected = true
				break
			}

			// check to make sure all slots are empty or match the id
			for _, slot := range player.PaintingSlots {
				if window.Inventory[slot].Present && window.Inventory[slot].ItemID != player.HeldSlot.ItemID {
					rejected = true
					break
				}
			}

			if rejected {
				break
			}

			amt := int(player.HeldSlot.ItemCount) / len(player.PaintingSlots)
			for _, slot := range player.PaintingSlots {
				stack := takeAmt(&player.HeldSlot, amt)
				if window.Inventory[slot].Present {
					placeStack(&stack, &window.Inventory[slot])
				} else {
					window.Inventory[slot] = stack
				}

			}
			player.PaintingSlots = nil
		}
	case 6: // double click
		// vanilla behaviour: only double clicks which come from picking up a new item do anything
		if window.Inventory[slot].Present || !player.HeldSlot.Present {
			rejected = true
			break
		}

		// look through all inventory slots to find matching items
		for i := range window.Inventory {
			if player.HeldSlot.ItemCount == 64 {
				break
			}
			if window.Inventory[i].ItemID == player.HeldSlot.ItemID {
				placeStack(&window.Inventory[i], &player.HeldSlot)
			}
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

// placeStack places as many items as it can from one stack onto another
// NOTE: placeStack does not check if stack item id's are correct...
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
// NOTE: does not care if x is greater than stack amount ...
func takeAmt(stack *pk.Slot, amt int) pk.Slot {
	res := pk.Slot{
		Present:   true,
		ItemID:    stack.ItemID,
		ItemCount: pk.Byte(amt),
		NBT:       stack.NBT,
	}

	stack.ItemCount -= pk.Byte(amt)
	if stack.ItemCount <= 0 {
		stack.Present = false
		stack.ItemID = 0
	}

	return res
}
