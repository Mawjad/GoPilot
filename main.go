package main

import rl "github.com/gen2brain/raylib-go/raylib"

type Entity int

type VelocityComponent struct {
	Velocity rl.Vector2
	Gravity  float32
}

func (v *VelocityComponent) ApplyGravity() {
	v.Velocity.Y += v.Gravity
}

func (v *VelocityComponent) CheckCollision(bounds *rl.Rectangle, obstacles []rl.Rectangle) {
	for _, obstacle := range obstacles {
		if rl.CheckCollisionRecs(*bounds, obstacle) {
			v.Velocity.Y = 0
		}
	}
}

func UpdatePosition(bounds *rl.Rectangle, velocityComponent *VelocityComponent) {
	bounds.X += velocityComponent.Velocity.X
	bounds.Y += velocityComponent.Velocity.Y
}

func main() {
	// Initialization
	rl.InitWindow(800, 450, "GoPilot")
	defer rl.CloseWindow()

	// Define the rectangle
	square := rl.Rectangle{X: 350, Y: 200, Width: 100, Height: 100}

	// Variables for dragging logic
	dragging := false
	mouseOffset := rl.Vector2{}

	// Setup the frame rate
	rl.SetTargetFPS(60)

	// Main game loop
	for !rl.WindowShouldClose() {
		// Update the dragging logic
		if rl.IsMouseButtonPressed(rl.MouseLeftButton) && rl.CheckCollisionPointRec(rl.GetMousePosition(), square) {
			dragging = true
			mousePosition := rl.GetMousePosition()
			mouseOffset.X = mousePosition.X - square.X
			mouseOffset.Y = mousePosition.Y - square.Y
		} else if rl.IsMouseButtonReleased(rl.MouseLeftButton) {
			dragging = false
		}

		if dragging {
			mousePosition := rl.GetMousePosition()
			square.X = mousePosition.X - mouseOffset.X
			square.Y = mousePosition.Y - mouseOffset.Y
		}

		// Begin drawing
		rl.BeginDrawing()

		rl.ClearBackground(rl.RayWhite)
		rl.DrawRectangleRec(square, rl.Gray)

		// End drawing
		rl.EndDrawing()
	}
}
