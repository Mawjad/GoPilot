package main

import rl "github.com/gen2brain/raylib-go/raylib"

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

func DragSystem(e *Entity, obstacles []rl.Rectangle) {
	if rl.IsMouseButtonPressed(rl.MouseLeftButton) && rl.CheckCollisionPointRec(rl.GetMousePosition(), e.Rectangle) {
		e.Draggable.Dragging = true
		mousePosition := rl.GetMousePosition()
		e.Draggable.MouseOffset.X = mousePosition.X - e.Rectangle.X
		e.Draggable.MouseOffset.Y = mousePosition.Y - e.Rectangle.Y
	} else if rl.IsMouseButtonReleased(rl.MouseLeftButton) {
		e.Draggable.Dragging = false
	}

	if e.Draggable.Dragging {
		mousePosition := rl.GetMousePosition()
		// Calculate new position without immediately updating the entity's rectangle
		newX := mousePosition.X - e.Draggable.MouseOffset.X
		newY := mousePosition.Y - e.Draggable.MouseOffset.Y

		// Create a temporary rectangle to represent the new position
		tempRect := rl.Rectangle{X: newX, Y: newY, Width: e.Rectangle.Width, Height: e.Rectangle.Height}

		// Assume no collision initially
		collisionDetected := false

		// Check for collisions with obstacles
		for _, obstacle := range obstacles {
			if rl.CheckCollisionRecs(tempRect, obstacle) {
				collisionDetected = true
				break // If any collision is detected, break the loop
			}
		}

		// Update the entity's position only if no collision is detected
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

	for !rl.WindowShouldClose() {
		if !player.Draggable.Dragging {
			player.Velocity.ApplyGravity()
		}

		DragSystem(&player, obstacles)

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
