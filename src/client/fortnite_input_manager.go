package main

import (
	pb "2dFortnite/proto"
	"github.com/veandco/go-sdl2/sdl"
	"time"
	"2dFortnite/pkg/shared"
)

type InputManager struct {
	commandChannel chan *pb.DoActionRequest	
	playerId uint64
	keyState map[int]bool
	currentSlot int32
	slotChanged bool
}

func NewInputManager(commands chan *pb.DoActionRequest, id uint64) *InputManager {
	return &InputManager{
		commandChannel: commands,
		keyState: make(map[int]bool),
		playerId: id,
	}
}

func (manager *InputManager) KeyEvent(key int, state uint32){
	manager.keyState[key] = (state & 1) == 0
}

func (manager *InputManager) MouseWheelEvent(scrollAmount int32, currentSlot int32){
	scrollDistance := scrollAmount % 5

	if scrollDistance != 0 {
		manager.currentSlot = (currentSlot + scrollDistance) % 5
		manager.slotChanged = true
	}
}


func (manager *InputManager) Run(){
	ticker := time.NewTicker(time.Second / fortnite.SERVER_TICKRATE)

	for {
		select {
			case <- ticker.C:
			x := 0.0
			y := 0.0
			if pressed, present := manager.keyState[sdl.K_w]; present && pressed {
				y -= 1
			}
			if pressed, present := manager.keyState[sdl.K_a]; present && pressed {
				x -= 1
			}
			if pressed,present := manager.keyState[sdl.K_s]; present && pressed {
				y += 1
			}
			if pressed, present := manager.keyState[sdl.K_d]; present && pressed {
				x += 1
			}
			
			moveAction := pb.DoActionRequest{
				PlayerId: &pb.PlayerId{Id:manager.playerId},
				ActionType: pb.ActionType_MOVE_PLAYER,
				ActionData: &pb.DoActionRequest_MovePlayer{
					MovePlayer: &pb.MovePlayerRequest{
						Vx: x,
						Vy: y,
						Facing: 0.0, // TODO fill in value
					},
				},
			}

			// opportunity to not send message if values have not changed
			manager.commandChannel <- &moveAction

			if manager.slotChanged{
				slotAction := pb.DoActionRequest{
					PlayerId: &pb.PlayerId{Id:manager.playerId},
					ActionType: pb.ActionType_SELECT_ITEM,
					ActionData: &pb.DoActionRequest_SelectItem{
						SelectItem: &pb.SelectItemRequest{
							SlotNumber: manager.currentSlot,
						},
					},
				}
				manager.commandChannel <- &slotAction
				manager.slotChanged = false
			}
		}
	}
}