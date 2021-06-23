package server

import (
	"github.com/GambitLLC/gogs/logger"
	"github.com/GambitLLC/gogs/net"
	pk "github.com/GambitLLC/gogs/net/packet"
	"github.com/GambitLLC/gogs/net/packet/clientbound"
	"github.com/GambitLLC/gogs/net/packet/packetids"
	"github.com/GambitLLC/gogs/net/packet/serverbound"
)

func (s *Server) onClickWindow(conn net.Conn, pkt pk.Packet) (err error) {
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
		if (in.Button != 0 && in.Button != 1) || slot < 0 {
			rejected = true
			break
		}

		// ignore shift clicking empty slots
		selectedItem := &window.Inventory[slot]
		if !selectedItem.Present {
			break
		}

		if in.WindowID == 0 {
			// TODO: handle items that shift into armor slots
			var inventory []pk.Slot

			if slot < 9 {
				// shift into main inventory or hot bar from armor/crafting slots
				inventory = player.Inventory[9:45]
			} else if slot < 36 {
				// shift into hot bar from main inventory
				inventory = player.Inventory[36:45]
			} else if slot >= 36 && slot < 45 {
				// shift from hot bar into main inventory
				inventory = player.Inventory[9:36]
			} else {
				rejected = true
				break
			}

			// find slots in the inventory with same item type
			for i := range inventory {
				if inventory[i].ItemID == selectedItem.ItemID {
					placeStack(selectedItem, &inventory[i])
				}
				if !selectedItem.Present {
					break
				}
			}

			if selectedItem.Present {
				// find an empty slot to place remainder into
				for i := range inventory {
					if !inventory[i].Present {
						inventory[i], *selectedItem = *selectedItem, inventory[i]
						break
					}
				}
			}
		}
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
			fallthrough
		case 4: // start right mouse drag
			player.PaintingLock.Lock()
			player.PaintingSlots = make([]uint8, 0, len(window.Inventory))
			player.PaintingLock.Unlock()
		case 1: // add slot for left mouse drag
			fallthrough
		case 5: // add slot for right mouse drag
			player.PaintingLock.Lock()
			defer player.PaintingLock.Unlock()
			if player.PaintingSlots == nil {
				rejected = true
				break
			}
			player.PaintingSlots = append(player.PaintingSlots, uint8(slot))
		case 2: // end left mouse drag
			fallthrough
		case 6: // end right mouse drag
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

			var amt int
			if in.Button == 2 { // left mouse
				amt = int(player.HeldSlot.ItemCount) / len(player.PaintingSlots)
			} else if in.Button == 6 { // right mouse
				amt = 1
			} else {
				rejected = true
				break
			}

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

	if err = conn.WritePacket(pk.Marshal(
		packetids.WindowConfirmationClientbound,
		in.WindowID,
		in.ActionNumber,
		pk.Boolean(!rejected), // accepted
	)); err != nil {
		return
	}

	if rejected {
		if err = conn.WritePacket(clientbound.WindowItems{
			WindowID: in.WindowID,
			Count:    pk.Short(len(window.Inventory)),
			SlotData: window.Inventory,
		}.CreatePacket()); err != nil {
			return
		}

		if slot != -999 {
			if err = conn.WritePacket(clientbound.SetSlot{
				WindowID: 0,
				Slot:     in.Slot,
				SlotData: window.Inventory[slot],
			}.CreatePacket()); err != nil {
				return
			}
		}

		if err = conn.WritePacket(clientbound.SetSlot{
			WindowID: -1,
			Slot:     -1,
			SlotData: player.HeldSlot,
		}.CreatePacket()); err != nil {
			return
		}
	}

	return
}

// placeStack places as many items as it can from one stack onto another
// NOTE: placeStack does not check if the two items are stackable...
func placeStack(from *pk.Slot, onto *pk.Slot) {
	sum := from.ItemCount + onto.ItemCount
	if sum > 64 {
		onto.ItemCount = 64
		from.ItemCount = sum - 64
	} else {
		onto.ItemCount = sum
		*from = pk.Slot{}
	}
}

// halveStack returns the greater half of the stack
func halveStack(stack *pk.Slot) pk.Slot {
	half := pk.Slot{
		Present:   stack.Present,
		ItemID:    stack.ItemID,
		ItemCount: stack.ItemCount/2 + (stack.ItemCount & 1),
		NBT:       stack.NBT,
	}

	stack.ItemCount /= 2
	if stack.ItemCount == 0 {
		*stack = pk.Slot{}
	}

	return half
}

// takeOne takes a single item from the stack and returns it
// NOTE: does not care if amt is greater than stack amount ...
func takeAmt(stack *pk.Slot, amt int) pk.Slot {
	res := pk.Slot{
		Present:   stack.Present,
		ItemID:    stack.ItemID,
		ItemCount: pk.Byte(amt),
		NBT:       stack.NBT,
	}

	stack.ItemCount -= pk.Byte(amt)
	if stack.ItemCount <= 0 {
		*stack = pk.Slot{}
	}

	return res
}
