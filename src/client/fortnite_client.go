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
	"image/color"
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
	var selectedMaterial int

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
		window, renderer, err = sdl.CreateWindowAndRenderer(WindowWidth, WindowHeight, sdl.WINDOW_OPENGL)
		window.SetTitle(WindowTitle)
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
	pickedUpItem := false
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

		if mB == 0{
			pickedUpItem = false
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

		for i := range currentWorldState.Walls {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				wall := currentWorldState.Walls[i]
				wallWorldX := float64(wall.X * fortnite.WALL_GRID_SIZE)
				wallWorldY := float64(wall.Y * fortnite.WALL_GRID_SIZE)
				if wall.Orientation == pb.WallOrientation_VERTICAL {
					wallScreenX := int32(wallWorldX - currentWorldState.Player.Position.X + fortnite.WALL_GRID_START_X + WindowWidth/2)
					wallScreenY := int32(wallWorldY - currentWorldState.Player.Position.Y + fortnite.WALL_GRID_START_Y + WindowHeight/2)

					wallScreenXEnd := wallScreenX
					wallScreenYEnd := wallScreenY + fortnite.WALL_GRID_SIZE

					wallColor := fortnite.MaterialColours[wall.Material]
					sdl.Do(func(){
						gfx.ThickLineRGBA(renderer, wallScreenX, wallScreenY, wallScreenXEnd, wallScreenYEnd, RectWidth, wallColor.R, wallColor.G, wallColor.B, wallColor.A)
					})
				}else {
					wallScreenX := int32(wallWorldX - currentWorldState.Player.Position.X + fortnite.WALL_GRID_START_X + WindowWidth/2)
					wallScreenY := int32(wallWorldY - currentWorldState.Player.Position.Y + fortnite.WALL_GRID_START_Y + WindowHeight/2)

					wallScreenXEnd := wallScreenX + fortnite.WALL_GRID_SIZE
					wallScreenYEnd := wallScreenY

					wallColor := fortnite.MaterialColours[wall.Material]
					sdl.Do(func(){
						gfx.ThickLineRGBA(renderer, wallScreenX, wallScreenY, wallScreenXEnd, wallScreenYEnd, RectWidth, wallColor.R, wallColor.G, wallColor.B, wallColor.A)
					})
				}
			}(i)
		}
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
					var drawColor color.RGBA
					var itemRect *sdl.Rect
					switch currentWorldState.Items[i].ItemType {
					case pb.ItemType_WEAPON:
						drawColor = fortnite.RarityColours[currentWorldState.Items[i].ItemRarity]
						
						
						itemRect = &sdl.Rect{
							X: int32(WindowWidth/2 + currentWorldState.Items[i].Pos.X - currentWorldState.Player.Position.X) - 12,
							Y: int32(WindowHeight/2 + currentWorldState.Items[i].Pos.Y - currentWorldState.Player.Position.Y) - 5,
							W: 25,
							H: 10,
						}
					case pb.ItemType_AMMO:
						drawColor = color.RGBA{0xFF, 0xFF, 0xFF, 0xFF}

						itemRect = &sdl.Rect{
							X: int32(WindowWidth/2 + currentWorldState.Items[i].Pos.X - currentWorldState.Player.Position.X) - 5,
							Y: int32(WindowHeight/2 + currentWorldState.Items[i].Pos.Y - currentWorldState.Player.Position.Y) - 5,
							W: 10,
							H: 10,
						}
					case pb.ItemType_MATERIAL:
						drawColor = fortnite.MaterialColours[currentWorldState.Items[i].GetMaterial()]

						itemRect = &sdl.Rect{
							X: int32(WindowWidth/2 + currentWorldState.Items[i].Pos.X - currentWorldState.Player.Position.X) - 5,
							Y: int32(WindowHeight/2 + currentWorldState.Items[i].Pos.Y - currentWorldState.Player.Position.Y) - 5,
							W: 10,
							H: 10,
						}
					}

					renderer.SetDrawColor(drawColor.R, drawColor.G, drawColor.B, drawColor.A)

					renderer.DrawRect(itemRect)

					if itemRect.IntersectLine(&mX, &mY, &mX, &mY) {
						renderer.SetDrawColor(255, 0, 0, 0xff)
						itemRect.X -= 3
						itemRect.Y -= 3
						itemRect.W += 6
						itemRect.H += 6
						renderer.DrawRect(itemRect)
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
							pickedUpItem = true
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
		if !pickedUpItem && mB == 1 && !inputManager.BuildWalls{
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

		if inputManager.BuildWalls {
			if mB == 4 && oldMb == 0 {
				selectedMaterial = (selectedMaterial + 1) % 3
			}
		}

		// draw UI

		if !inputManager.BuildWalls{
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

							if inventoryInfo.Item == pb.ItemType_WEAPON {
								if inventoryInfo.Reload != 0 {
									// draw remaining reload time
									reloadSurface, _ := uiFont.RenderUTF8Solid(fmt.Sprintf("%.1f", float32(inventoryInfo.Reload) / fortnite.SERVER_TICKRATE), sdl.Color{R: 255, G: 255, B: 255, A: 255})
									reloadTexture, _ := renderer.CreateTextureFromSurface(reloadSurface)

									renderer.Copy(reloadTexture, nil, &sdl.Rect{
										X: int32(WindowWidth - (55 * 5) - 5 + 55 * i + 3 + 22 - 16),
										Y: int32(WindowHeight- 60 + 3 + 22 - 16),
										W: 32,
										H: 32,
									})
								}
							} else {
								if inventoryInfo.Cooldown != 0 {
									// draw remaining cooldown
									cooldownSurface, _ := uiFont.RenderUTF8Solid(fmt.Sprintf("%.1f", float32(inventoryInfo.Cooldown) / fortnite.SERVER_TICKRATE), sdl.Color{R: 255, G: 255, B: 255, A: 255})
									cooldownTexture, _ := renderer.CreateTextureFromSurface(cooldownSurface)

									renderer.Copy(cooldownTexture, nil, &sdl.Rect{
										X: int32(WindowWidth - (55 * 5) - 5 + 55 * i + 3 + 22 - 16),
										Y: int32(WindowHeight- 60 + 3 + 22 - 16),
										W: 32,
										H: 32,
									})
								}
							}
						}
					})
					wg.Done()
				}(i)
			}
		}else{
			for i := range currentWorldState.Player.Resources {
				wg.Add(1)
				go func(slot int){
					sdl.Do(
						func() {
							resourceInfo := currentWorldState.Player.Resources[slot]
							if resourceInfo.Item == pb.ItemType_MATERIAL {
								renderer.SetDrawColor(200, 200, 200, 0xff)
								drawPosition := MaterialDrawPositions[resourceInfo.GetMaterial()]
								switch resourceInfo.GetMaterial() {
								case pb.Material_WOOD:
									if selectedMaterial == 0 {
										renderer.SetDrawColor(255, 100, 100, 0xff)
									}
								case pb.Material_BRICK:
									if selectedMaterial == 1{
										renderer.SetDrawColor(255, 100, 100, 0xff)
									}
								case pb.Material_METAL:
									if selectedMaterial == 2{
										renderer.SetDrawColor(255, 100, 100, 0xff)
									}
								}
								renderer.DrawRect(&drawPosition)
								drawColor := fortnite.MaterialColours[resourceInfo.GetMaterial()]
								renderer.SetDrawColor(drawColor.R, drawColor.G, drawColor.B, drawColor.A)
								drawPosition.X += 3
								drawPosition.Y += 3
								drawPosition.W -= 6
								drawPosition.H -= 6
								renderer.FillRect(&drawPosition)

								drawPosition.X += 6
								drawPosition.Y += 6
								drawPosition.W -= 12
								drawPosition.H -= 12

								// render amount of material
								amountSurface, _ := uiFont.RenderUTF8Solid(fmt.Sprintf("%d", resourceInfo.StackSize), sdl.Color{R: 255, G: 255, B: 255, A: 255})
								amountTexture, _ := renderer.CreateTextureFromSurface(amountSurface)
								renderer.Copy(amountTexture, nil, &drawPosition)

							}
						})
					wg.Done()
				}(i)
			}

			wg.Add(1)
			go func(){
				sdl.Do(func(){
					for i := int32(0); i < 2; i++ {
						if i == inputManager.BuildSlot {
							renderer.SetDrawColor(255, 100, 100, 0xff)
						} else {
							renderer.SetDrawColor(200, 200, 200, 0xff)
						}

						renderer.DrawRect(&sdl.Rect{
							X: int32(WindowWidth - (55 * 2) - 5 + 55 * i),
							Y: int32(WindowHeight- 60),
							W: 50,
							H: 50,
						})

						drawColor := fortnite.MaterialColours[pb.Material(selectedMaterial)]

						switch i {
						case 0:
							// vertical
							gfx.ThickLineRGBA(renderer, WindowWidth - (55 * 2) - 5 + 55 * i + 25, WindowHeight- 60 + 3, WindowWidth - (55 * 2) - 5 + 55 * i + 25, WindowHeight- 60 + 3 + 44, 5, drawColor.R, drawColor.G, drawColor.B, drawColor.A)
						case 1:
							// horizontal
							gfx.ThickLineRGBA(renderer, WindowWidth - (55 * 2) - 5 + 55 * i + 3, WindowHeight- 60 + 25, WindowWidth - (55 * 2) - 5 + 55 * i + 3 + 44, WindowHeight- 60 + 25, 5, drawColor.R, drawColor.G, drawColor.B, drawColor.A)
						}
					}
				})
				wg.Done()
			}()

			wg.Add(1)
			go func(){
				// draw blue line over where wall is being placed
				wallGridX := math.Round((float64(mX) - fortnite.WALL_GRID_SIZE/2 + currentWorldState.Player.Position.X - WindowWidth/2) / fortnite.WALL_GRID_SIZE)
				wallGridY := math.Round((float64(mY) - fortnite.WALL_GRID_SIZE/2 + currentWorldState.Player.Position.Y - WindowHeight/2) / fortnite.WALL_GRID_SIZE)
				if inputManager.BuildSlot == 0 {
					// vertical
					wallStartX :=  fortnite.WALL_GRID_SIZE * wallGridX + fortnite.WALL_GRID_START_X - currentWorldState.Player.Position.X + WindowWidth/2
					wallStartY :=  fortnite.WALL_GRID_SIZE * wallGridY + fortnite.WALL_GRID_START_Y - currentWorldState.Player.Position.Y + WindowHeight/2
					wallEndX := wallStartX
					wallEndY := wallStartY + fortnite.WALL_GRID_SIZE
					sdl.Do(func(){
						gfx.ThickLineRGBA(renderer, int32(wallStartX), int32(wallStartY), int32(wallEndX), int32(wallEndY), 5, 119, 184, 217, 0xff)
					})

				}else{
					// horizontal
					wallStartX := fortnite.WALL_GRID_SIZE *wallGridX + fortnite.WALL_GRID_START_X - currentWorldState.Player.Position.X + WindowWidth/2
					wallStartY := fortnite.WALL_GRID_SIZE *wallGridY + fortnite.WALL_GRID_START_Y - currentWorldState.Player.Position.Y + WindowHeight/2
					wallEndX := wallStartX + fortnite.WALL_GRID_SIZE
					wallEndY := wallStartY
					sdl.Do(func(){
						gfx.ThickLineRGBA(renderer, int32(wallStartX), int32(wallStartY), int32(wallEndX), int32(wallEndY), 5, 119, 184, 217, 0xff)
					})
				}

				if mB == 1 && oldMb == 0 && !pickedUpItem{
					log.Printf("Placing wall at %f, %f", wallGridX, wallGridY)
					(*client).DoAction(
						context.Background(),
						&pb.DoActionRequest{
							ActionType: pb.ActionType_BUILD_WALL,
							PlayerId: &pb.PlayerId{Id: id},
							ActionData: &pb.DoActionRequest_BuildWall {
								BuildWall: &pb.BuildWallRequest{
									X: int64(wallGridX),
									Y: int64(wallGridY),
									Facing: pb.WallOrientation(uint32(inputManager.BuildSlot)),
									Material: pb.Material(selectedMaterial),
								},
							},
						})		
				}
				wg.Done()
			}()
		}

		if selectedSlot := currentWorldState.Player.Inventory[currentWorldState.Player.EquippedSlot]; selectedSlot.Item == pb.ItemType_WEAPON && !inputManager.BuildWalls {
			// draw weapon name and ammo if equipped
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

		// draw health and shield bars
		wg.Add(1)
		go func(){sdl.Do(func() {

			healthBarBacking := sdl.Rect{
				X: 10,
				Y: WindowHeight - 60,
				W: 200,
				H: 40,
			}

			shieldBarBacking := sdl.Rect{
				X: 10,
				Y: WindowHeight - 60 - 50,
				W: 200,
				H: 40,
			}

			renderer.SetDrawColor(150, 150, 150, 0xff)

			renderer.FillRect(&healthBarBacking)
			renderer.FillRect(&shieldBarBacking)

			healthBar := sdl.Rect{
				X: healthBarBacking.X,
				Y: healthBarBacking.Y,
				W: int32(200 * currentWorldState.Player.Health / 100),
				H: healthBarBacking.H,
			}

			shieldBar := sdl.Rect{
				X: shieldBarBacking.X,
				Y: shieldBarBacking.Y,
				W: int32(200 * currentWorldState.Player.Shields / 100),
				H: shieldBarBacking.H,
			}

			renderer.SetDrawColor(58, 245, 29, 0xff)
			renderer.FillRect(&healthBar)
			renderer.SetDrawColor(29, 101, 245, 0xff)
			renderer.FillRect(&shieldBar)

			// draw health and shields text
			healthTextSurface, _ := uiFont.RenderUTF8Solid(fmt.Sprintf("%d", currentWorldState.Player.Health), sdl.Color{R: 0xff, G: 0xff, B: 0xff, A: 0xff})
			shieldTextSurface, _ := uiFont.RenderUTF8Solid(fmt.Sprintf("%d", currentWorldState.Player.Shields), sdl.Color{R: 0xff, G: 0xff, B: 0xff, A: 0xff})

			healthTextTexture, _ := renderer.CreateTextureFromSurface(healthTextSurface)
			shieldTextTexture, _ := renderer.CreateTextureFromSurface(shieldTextSurface)

			healthHeight := int32(healthTextSurface.Bounds().Max.Y)
			shieldHeight := int32(shieldTextSurface.Bounds().Max.Y)

			renderer.Copy(healthTextTexture, &sdl.Rect{X: 0, Y: 0, W: healthTextSurface.W, H: healthTextSurface.H}, &sdl.Rect{X: 15, Y: WindowHeight - 60 + 25 - healthHeight/2, W: healthTextSurface.W, H: healthTextSurface.H})
			renderer.Copy(shieldTextTexture, &sdl.Rect{X: 0, Y: 0, W: shieldTextSurface.W, H: shieldTextSurface.H}, &sdl.Rect{X: 15, Y: WindowHeight - 60 - 50 + 25 - shieldHeight/2, W: shieldTextSurface.W, H: shieldTextSurface.H})
			wg.Done()
		})}()
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