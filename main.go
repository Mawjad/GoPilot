package main

import (
	"fmt"
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const maxSpeed float32 = 15

type Entity struct {
	Rectangle rl.Rectangle
	Velocity  VelocityComponent
	Draggable DragComponent
}

type VelocityComponent struct {
	Velocity rl.Vector2
	Gravity  float32
}

type DragComponent struct {
	Dragging    bool
	MouseOffset rl.Vector2
}

func Sign(x float64) float32 {
	if math.Signbit(x) {
		return -1
	}
	if x == 0 {
		return 0
	}
	return 1
}

func (v *VelocityComponent) ApplyGravity() {
	v.Velocity.Y += v.Gravity
}

func (v *VelocityComponent) CheckCollision(bounds *rl.Rectangle, obstacles []rl.Rectangle) {
	for _, obstacle := range obstacles {
		if rl.CheckCollisionRecs(*bounds, obstacle) {
			// Simplified collision response
			if v.Velocity.Y > 0 { // Moving down
				bounds.Y = obstacle.Y - bounds.Height
				v.Velocity.Y = 0
			} else if v.Velocity.Y < 0 { // Moving up
				bounds.Y = obstacle.Y + obstacle.Height
				v.Velocity.Y = 0
			}

			if v.Velocity.X > 0 { // Moving right
				bounds.X = obstacle.X - bounds.Width
				v.Velocity.X = 0
			} else if v.Velocity.X < 0 { // Moving left
				bounds.X = obstacle.X + obstacle.Width
				v.Velocity.X = 0
			}
		}
	}
}

func UpdatePosition(e *Entity, screenWidth, screenHeight int32) {
	e.Rectangle.X += e.Velocity.Velocity.X
	e.Rectangle.Y += e.Velocity.Velocity.Y

	if e.Rectangle.Y > float32(screenHeight) {
		e.Rectangle.Y = -e.Rectangle.Height
	}
}

func DragSystem(e *Entity, obstacles []rl.Rectangle, deltaTime float64) {
	mousePosition := rl.GetMousePosition()
	if rl.IsMouseButtonPressed(rl.MouseLeftButton) && rl.CheckCollisionPointRec(mousePosition, e.Rectangle) {
		e.Draggable.Dragging = true
		e.Draggable.MouseOffset.X = mousePosition.X - e.Rectangle.X
		e.Draggable.MouseOffset.Y = mousePosition.Y - e.Rectangle.Y
	} else if rl.IsMouseButtonReleased(rl.MouseLeftButton) && e.Draggable.Dragging {
		e.Draggable.Dragging = false
		clamping_value := 0.1
		// Convert Raylib's float32 to float64 for calculation, then back to float32
		e.Velocity.Velocity.X = float32((float64(mousePosition.X) - float64(e.Rectangle.X+e.Draggable.MouseOffset.X)) / deltaTime * clamping_value)
		e.Velocity.Velocity.Y = float32((float64(mousePosition.Y) - float64(e.Rectangle.Y+e.Draggable.MouseOffset.Y)) / deltaTime * clamping_value)
		e.Velocity.Velocity.X = float32(math.Min(float64(maxSpeed), math.Abs(float64(e.Velocity.Velocity.X)))) * Sign(float64(e.Velocity.Velocity.X))
		e.Velocity.Velocity.Y = float32(math.Min(float64(maxSpeed), math.Abs(float64(e.Velocity.Velocity.Y)))) * Sign(float64(e.Velocity.Velocity.Y))
		//debug
		fmt.Printf("Velocity X: %v, Y: %v\n", e.Velocity.Velocity.X, e.Velocity.Velocity.Y)
	}

	if e.Draggable.Dragging {
		newX := mousePosition.X - e.Draggable.MouseOffset.X
		newY := mousePosition.Y - e.Draggable.MouseOffset.Y
		tempRect := rl.Rectangle{X: newX, Y: newY, Width: e.Rectangle.Width, Height: e.Rectangle.Height}

		collisionDetected := false
		for _, obstacle := range obstacles {
			if rl.CheckCollisionRecs(tempRect, obstacle) {
				collisionDetected = true
				break
			}
		}

		if !collisionDetected {
			e.Rectangle.X = newX
			e.Rectangle.Y = newY
		}
	}
}

func main() {
	screenWidth := int32(1200)
	screenHeight := int32(850)
	rl.InitWindow(screenWidth, screenHeight, "GoPilot ECS")
	defer rl.CloseWindow()

	player := Entity{
		Rectangle: rl.Rectangle{X: 350, Y: 200, Width: 100, Height: 100},
		Velocity:  VelocityComponent{Gravity: 0.5},
		Draggable: DragComponent{},
	}

	obstacles := []rl.Rectangle{}

	obstacles = append(obstacles, rl.Rectangle{X: 200, Y: 400, Width: 800, Height: 50})
	obstacles = append(obstacles, rl.Rectangle{X: 200, Y: 200, Width: 200, Height: 100})

	rl.SetTargetFPS(60)

	var lastFrameTime float64 = rl.GetTime()

	for !rl.WindowShouldClose() {

		currentTime := rl.GetTime()
		deltaTime := currentTime - lastFrameTime
		lastFrameTime = currentTime

		if !player.Draggable.Dragging {
			player.Velocity.ApplyGravity()
		}

		DragSystem(&player, obstacles, deltaTime)

		player.Velocity.CheckCollision(&player.Rectangle, obstacles)

		if !player.Draggable.Dragging {
			UpdatePosition(&player, screenWidth, screenHeight)
		}

		rl.BeginDrawing()
		rl.ClearBackground(rl.RayWhite)
		rl.DrawRectangleRec(player.Rectangle, rl.Red)
		for _, obstacle := range obstacles {
			rl.DrawRectangleRec(obstacle, rl.Blue)
		}
		rl.EndDrawing()
	}
}
