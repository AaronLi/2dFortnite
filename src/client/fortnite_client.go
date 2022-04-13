package main

import (
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/gfx"
	"github.com/veandco/go-sdl2/ttf"
	pb "2dFortnite/proto"
	"2dFortnite/pkg/shared"
	"context"
	"os"
	"log"
	"sync"
	"fmt"
	"math"
)

const (
	WindowTitle = "2d Fortnite"
	WindowWidth = 1280
	WindowHeight = 720
	FrameRate = 60

	RectWidth = 20
	RectHeight = 40
	NumRects = WindowHeight / RectHeight
)
var runningMutex sync.Mutex

func run(userInfo *pb.RegisterPlayerRequest, id uint64, client *pb.FortniteServiceClient) int {
	var window *sdl.Window
	var renderer *sdl.Renderer
	var fpsManager gfx.FPSmanager
	var err error
	var currentWorldState *pb.WorldStateResponse
	var uiFont *ttf.Font

	sdl.Do(func(){
		ttf.Init()
		uiFont, err = ttf.OpenFont("fortnite.otf", 30)
	})
	inputManagerCommands := make(chan *pb.DoActionRequest)

	inputManager := NewInputManager(inputManagerCommands, id)

	worldUpdateChan := make(chan *pb.WorldStateResponse)

	go readWorldUpdates(id, client, worldUpdateChan)
	go inputManager.Run()

	if err != nil {
		panic(err)
	}

	sdl.Do(func() {
		window, err = sdl.CreateWindow(WindowTitle, sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, WindowWidth, WindowHeight, sdl.WINDOW_OPENGL)
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create window: %s\n", err)
		return 1
	}
	defer func() {
		sdl.Do(func() {
			window.Destroy()
		})
	}()

	sdl.Do(func() {
		gfx.InitFramerate(&fpsManager)
		gfx.SetFramerate(&fpsManager, FrameRate)
	})

	sdl.Do(func() {
		renderer, err = sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	})
	if err != nil {
		fmt.Fprint(os.Stderr, "Failed to create renderer: %s\n", err)
		return 2
	}
	defer func() {
		sdl.Do(func() {
			renderer.Destroy()
		})
	}()

	sdl.Do(func() {
		renderer.Clear()
	})
	var oldMb uint32= 0 // old mouse button state
	running := true
	for running {
		sdl.Do(func() {
			for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
				switch event.(type) {
				case *sdl.QuitEvent:
					runningMutex.Lock()
					running = false
					runningMutex.Unlock()
				case *sdl.KeyboardEvent:
					inputManager.KeyEvent(int(event.(*sdl.KeyboardEvent).Keysym.Sym), event.(*sdl.KeyboardEvent).Type)
				case *sdl.MouseWheelEvent:
					inputManager.MouseWheelEvent(event.(*sdl.MouseWheelEvent).Y, currentWorldState.Player.EquippedSlot)
				}
			}
			renderer.Clear()
			renderer.SetDrawColor(66, 66, 66, 0xFF)
			renderer.FillRect(&sdl.Rect{0, 0, WindowWidth, WindowHeight})
		})
		mX, mY, mB := sdl.GetMouseState()

		if mB == 1{
			// try to use current item
			switch currentWorldState.Player.Inventory[currentWorldState.Player.EquippedSlot].Item {
				case pb.ItemType_WEAPON:
					// try to fire weapon
					(*client).DoAction(context.Background(), &pb.DoActionRequest{
						ActionType: pb.ActionType_SHOOT_PROJECTILE,
						PlayerId: &pb.PlayerId{
							Id: id,
						},
						ActionData: &pb.DoActionRequest_ShootProjectile{
							ShootProjectile: &pb.ShootProjectileRequest{
								Facing: math.Atan2(float64(mY - WindowHeight/2), float64(mX - WindowWidth/2)) / math.Pi * 180,
							},
						},
					})
						
				case pb.ItemType_CONSUMABLE:
					if oldMb == 0 {
						// do thing
					}
			}
		}

		select {
		case worldUpdate := <-worldUpdateChan:
			currentWorldState = worldUpdate
		default:
		}

		select {
		case input := <-inputManagerCommands:
			_, err := (*client).DoAction(context.Background(), input)
			if err != nil {
				panic(err)
			}
		default:
		}
		// Draw game world
		wg := sync.WaitGroup{}
		for i := range currentWorldState.Players {
			wg.Add(1)
			go func(i int) {
				sdl.Do(func() {
					renderer.SetDrawColor(0xff, 0xff, 0xff, 0xff)
					drawX := int32(WindowWidth/2 + currentWorldState.Players[i].Position.X - currentWorldState.Player.Position.X)
					drawY := int32(WindowHeight/2 + currentWorldState.Players[i].Position.Y - currentWorldState.Player.Position.Y)
					gfx.FilledCircleRGBA(renderer, drawX, drawY, fortnite.PLAYER_RADIUS, 0xff, 0xff, 0xff, 0xff)
					currentWorldState.Players[i].Position.X += currentWorldState.Players[i].Position.VX * (1.0 / fortnite.SERVER_TICKRATE) * (fortnite.SERVER_TICKRATE / float64(FrameRate))
					currentWorldState.Players[i].Position.Y += currentWorldState.Players[i].Position.VY * (1.0 / fortnite.SERVER_TICKRATE) * (fortnite.SERVER_TICKRATE / float64(FrameRate))
				})
				wg.Done()
			}(i)
		}

		for i := range currentWorldState.Items {
			wg.Add(1)
			go func(i int) {
				sdl.Do(func() {
					color := fortnite.RarityColours[currentWorldState.Items[i].ItemRarity]
						renderer.SetDrawColor(color.R, color.G, color.B, color.A)
					
					itemRect := sdl.Rect{
						X: int32(WindowWidth/2 + currentWorldState.Items[i].Pos.X - currentWorldState.Player.Position.X),
						Y: int32(WindowHeight/2 + currentWorldState.Items[i].Pos.Y - currentWorldState.Player.Position.Y),
						W: 25,
						H: 10,
					}

					renderer.DrawRect(&itemRect)

					if itemRect.IntersectLine(&mX, &mY, &mX, &mY) {
						renderer.SetDrawColor(255, 0, 0, 0xff)
						itemRect.X -= 3
						itemRect.Y -= 3
						itemRect.W += 6
						itemRect.H += 6
						renderer.DrawRect(&itemRect)
						if mB == 1 && oldMb == 0 {
							(*client).DoAction(context.Background(), &pb.DoActionRequest{
								ActionType: pb.ActionType_PICKUP_ITEM,
								PlayerId: &pb.PlayerId{Id: id},
								ActionData: &pb.DoActionRequest_PickupItem{
									PickupItem: &pb.PickupItemRequest{
										ItemId: currentWorldState.Items[i].Id,
									},
								},
							})
						}
					}
				})
				wg.Done()
			}(i)
		}

		for i := range currentWorldState.Projectiles {
			wg.Add(1)
			go func(i int) {
				sdl.Do(func() {
					renderer.SetDrawColor(0xff, 0xff, 0xff, 0xff)
					gfx.FilledCircleRGBA(renderer, 
						int32(WindowWidth/2 + currentWorldState.Projectiles[i].Position.X - currentWorldState.Player.Position.X),
						int32(WindowHeight/2 + currentWorldState.Projectiles[i].Position.Y - currentWorldState.Player.Position.Y),
						3,
						0xff, 0xff, 0xff, 0xff)
					
					currentWorldState.Projectiles[i].Position.X += currentWorldState.Projectiles[i].Position.VX * (1.0 / fortnite.SERVER_TICKRATE) * (fortnite.SERVER_TICKRATE / float64(FrameRate))
					currentWorldState.Projectiles[i].Position.Y += currentWorldState.Projectiles[i].Position.VY * (1.0 / fortnite.SERVER_TICKRATE) * (fortnite.SERVER_TICKRATE / float64(FrameRate))
				})
				wg.Done()
			}(i)
		}
		wg.Wait()

		// draw UI

		for i := range currentWorldState.Player.Inventory {
			wg.Add(1)
			go func(i int) {
				sdl.Do(func() {
					inventoryInfo := currentWorldState.Player.Inventory[i]
					if currentWorldState.Player.EquippedSlot == int32(i) {
						renderer.SetDrawColor(255, 100, 100, 0xff)
					} else{
						renderer.SetDrawColor(200, 200, 200, 0xff)
					}
					

					itemRect := sdl.Rect{
						X: int32(WindowWidth - (55 * 5) - 5 + 55 * i),
						Y: int32(WindowHeight- 60),
						W: 50,
						H: 50,
					}

					renderer.DrawRect(&itemRect)
					if inventoryInfo.Item != pb.ItemType_NONE {
						color := fortnite.RarityColours[inventoryInfo.Rarity]
						renderer.SetDrawColor(color.R, color.G, color.B, color.A)
						renderer.FillRect(&sdl.Rect{
							X: int32(WindowWidth - (55 * 5) - 5 + 55 * i + 3),
							Y: int32(WindowHeight- 60 + 3),
							W: 44,
							H: 44,
						})
					}
				})
				wg.Done()
			}(i)
		}

		if selectedSlot := currentWorldState.Player.Inventory[currentWorldState.Player.EquippedSlot]; selectedSlot.Item == pb.ItemType_WEAPON {
			wg.Add(1)
			go func(){sdl.Do(func() {

				var ammoRemaining uint32

				for _, stack := range currentWorldState.Player.Resources {
					if stack.Item == pb.ItemType_AMMO {
						if stack.GetAmmo() == fortnite.WeaponAmmoUsage[selectedSlot.GetWeapon()] {
							ammoRemaining = stack.StackSize
							break
						}
					}
				}

				drawSurface, _ := uiFont.RenderUTF8Solid(fmt.Sprintf("%s %d / %d",fortnite.WeaponDisplayNames[selectedSlot.GetWeapon()], selectedSlot.StackSize, ammoRemaining), sdl.Color{R: 0xff, G: 0xff, B: 0xff, A: 0xff})
				
				drawTexture, _ := renderer.CreateTextureFromSurface(drawSurface)
				bounds := drawSurface.Bounds()

				width := int32(bounds.Max.X)
				height := int32(bounds.Max.Y)

				renderer.Copy(drawTexture, &sdl.Rect{X: 0, Y: 0, W: width, H: height}, &sdl.Rect{X: WindowWidth - (55 * 5), Y: WindowHeight - 60 - height, W: width, H: height})
			})
			wg.Done()
		}()
		}
		wg.Wait()
		

		sdl.Do(func() {
			renderer.Present()
			gfx.FramerateDelay(&fpsManager)
		})
		oldMb = mB
	}

	return 0
}

func readWorldUpdates(id uint64, client *pb.FortniteServiceClient, worldUpdates chan *pb.WorldStateResponse) {
	worldInfo, err := (*client).WorldState(context.Background(), &pb.PlayerId{
		Id: id,
	})

	if err != nil {
		log.Println("Failed to get world state:", err)
		return
	}

	for {
		response, err := worldInfo.Recv()
		if err != nil {
			log.Println("Error receiving world state:", err)
			return
		}
		worldUpdates <- response
	}
}