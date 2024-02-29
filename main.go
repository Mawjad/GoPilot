package main

import (
	"fmt"
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	MaxVelocity        float32 = 2000.0
	Gravity            float32 = 1280.0
	PlayerAcceleration float32 = 60.0
	PlayerElasticity   float32 = 0.02
	PlayerFriction     float32 = 0.02
	PlayerDrag         float64 = 0.02
	screenWidth        int32   = 1200
	screenHeight       int32   = 850
)

type Entity struct {
	Body      rl.Rectangle
	Physics   PhysicsComponent
	Draggable PlayerDragComponent
}

type PhysicsComponent struct {
	Velocity     rl.Vector2
	Acceleration rl.Vector2
	MaxVelocity  float32
	Gravity      float32
	Mass         float32
	Friction     float32
	Elasticity   float32
	Drag         float64
}

type PlayerDragComponent struct {
	Dragging    bool
	MouseOffset rl.Vector2
}

func NewPhysicsComponent(mass float32, friction float32, elasticity float32, drag float64) PhysicsComponent {
	return PhysicsComponent{
		Velocity:     rl.Vector2{X: 0, Y: 0},
		Acceleration: rl.Vector2{X: 0, Y: 0},
		Gravity:      Gravity,
		Mass:         mass,
		Friction:     friction,
		Elasticity:   elasticity,
		Drag:         drag,
	}
}

func (pc *PhysicsComponent) ApplyGravity() {
	gravitationalForce := pc.Gravity * pc.Mass
	gravitationalAcceleration := gravitationalForce / pc.Mass
	pc.Acceleration.Y += gravitationalAcceleration
}

func (pc *PhysicsComponent) ApplyForce(force rl.Vector2) {
	accelerationDueToForce := rl.Vector2{X: force.X / pc.Mass, Y: force.Y / pc.Mass}
	pc.Acceleration.X += accelerationDueToForce.X
	pc.Acceleration.Y += accelerationDueToForce.Y
}

func (pc *PhysicsComponent) ApplyDrag() {
	dragForceX := -pc.Drag * float64(pc.Velocity.X) * math.Abs(float64(pc.Velocity.X))
	dragForceY := -pc.Drag * float64(pc.Velocity.Y) * math.Abs(float64(pc.Velocity.Y))

	pc.Acceleration.X += float32(dragForceX) / pc.Mass
	pc.Acceleration.Y += float32(dragForceY) / pc.Mass
}

func CheckAABBCollision(rect1, rect2 rl.Rectangle) bool {
	if rect1.X+rect1.Width < rect2.X || rect2.X+rect2.Width < rect1.X {
		return false
	}
	if rect1.Y+rect1.Height < rect2.Y || rect2.Y+rect2.Height < rect1.Y {
		return false
	}
	return true
}

func GetCollisionSide(entityBody rl.Rectangle, obstacle rl.Rectangle) (collides bool, side string) {
	if !CheckAABBCollision(entityBody, obstacle) {
		return false, ""
	}

	// Calculate overlap on both axes
	overlapX := math.Min(float64(entityBody.X+entityBody.Width), float64(obstacle.X+obstacle.Width)) -
		math.Max(float64(entityBody.X), float64(obstacle.X))
	overlapY := math.Min(float64(entityBody.Y+entityBody.Height), float64(obstacle.Y+obstacle.Height)) -
		math.Max(float64(entityBody.Y), float64(obstacle.Y))

	// Determine primary collision side based on smallest overlap
	if overlapX > overlapY {
		// Vertical collision is more significant
		if entityBody.Y+entityBody.Height/2 < obstacle.Y+obstacle.Height/2 {
			return true, "top"
		} else {
			return true, "bottom"
		}
	} else {
		// Horizontal collision is more significant
		if entityBody.X+entityBody.Width/2 < obstacle.X+obstacle.Width/2 {
			return true, "left"
		} else {
			return true, "right"
		}
	}
}

func capVelocity(pc *PhysicsComponent) {
	velocityLength := float32(math.Sqrt(float64(pc.Velocity.X*pc.Velocity.X + pc.Velocity.Y*pc.Velocity.Y)))
	if velocityLength > MaxVelocity {
		scaleFactor := MaxVelocity / velocityLength
		pc.Velocity.X *= scaleFactor
		pc.Velocity.Y *= scaleFactor
	}
}

func adjustPositionForCollision(body *rl.Rectangle, side string, obstacle rl.Rectangle) {
	switch side {
	case "left":
		body.X = obstacle.X - body.Width
	case "right":
		body.X = obstacle.X + obstacle.Width
	case "top":
		body.Y = obstacle.Y - body.Height
	case "bottom":
		body.Y = obstacle.Y + obstacle.Height
	}
}

func ProcessMouseDrag(player *Entity, deltaTime float64) {
	if rl.IsMouseButtonDown(rl.MouseLeftButton) {
		if !player.Draggable.Dragging {
			mousePosition := rl.GetMousePosition()
			// Check if the mouse is within the player's body for initial click
			if rl.CheckCollisionPointRec(mousePosition, player.Body) {
				player.Draggable.Dragging = true
				player.Draggable.MouseOffset.X = mousePosition.X - player.Body.X
				player.Draggable.MouseOffset.Y = mousePosition.Y - player.Body.Y
			}
		} else {
			// Calculate the direction and apply acceleration towards the cursor
			mousePosition := rl.GetMousePosition()
			direction := rl.Vector2{X: mousePosition.X - (player.Body.X + player.Draggable.MouseOffset.X), Y: mousePosition.Y - (player.Body.Y + player.Draggable.MouseOffset.Y)}
			direction = rl.Vector2Normalize(direction)
			player.Physics.Acceleration.X += direction.X * PlayerAcceleration
			player.Physics.Acceleration.Y += direction.Y * PlayerAcceleration
		}
	} else {
		player.Draggable.Dragging = false
	}
}

func (pc *PhysicsComponent) Update(entity *Entity, deltaTime float64, obstacles []rl.Rectangle) {
	pc.ApplyGravity()

	pc.ApplyDrag()

	// Temporarily calculate the new position without applying friction globally
	tempVelocity := pc.Velocity
	tempVelocity.X += pc.Acceleration.X * float32(deltaTime)
	tempVelocity.Y += pc.Acceleration.Y * float32(deltaTime)

	tempBody := rl.Rectangle{
		X:      entity.Body.X + tempVelocity.X*float32(deltaTime),
		Y:      entity.Body.Y + tempVelocity.Y*float32(deltaTime),
		Width:  entity.Body.Width,
		Height: entity.Body.Height,
	}

	collisionOccurred := false

	for _, obstacle := range obstacles {
		collides, side := GetCollisionSide(tempBody, obstacle)
		if collides {
			fmt.Printf("Collision with %v at %v side. Vel before: %v\n", obstacle, side, pc.Velocity)
			collisionOccurred = true

			if side == "left" || side == "right" {
				tempVelocity.X = -tempVelocity.X * pc.Elasticity
				tempVelocity.X *= 1 - pc.Friction
			}
			if side == "top" {
				tempVelocity.Y = 0
				entity.Body.Y = obstacle.Y - entity.Body.Height
				tempVelocity.X *= 1 - pc.Friction
			} else if side == "bottom" {
				tempVelocity.Y = -tempVelocity.Y * pc.Elasticity // Bounce effect
			}

			adjustPositionForCollision(&entity.Body, side, obstacle)
			fmt.Printf("Vel after: %v, Pos after: %v\n", pc.Velocity, entity.Body)
			break // Handle one collision at a time for simplicity
		}
	}

	// Ensure the adjusted velocity is applied to the entity
	if collisionOccurred {
		pc.Velocity = tempVelocity
	} else {
		// Update velocity without collisions
		pc.Velocity.X += pc.Acceleration.X * float32(deltaTime)
		pc.Velocity.Y += pc.Acceleration.Y * float32(deltaTime)
	}

	// Update entity's position based on the adjusted or continued velocity
	entity.Body.X += pc.Velocity.X * float32(deltaTime)
	entity.Body.Y += pc.Velocity.Y * float32(deltaTime)

	capVelocity(pc)
}

func main() {
	rl.InitWindow(screenWidth, screenHeight, "Physics and Collision Demo")
	defer rl.CloseWindow()

	resetGame := func() Entity {
		return Entity{
			Body:      rl.Rectangle{X: 350, Y: 200, Width: 100, Height: 100},
			Physics:   NewPhysicsComponent(1.0, PlayerFriction, PlayerElasticity, PlayerDrag),
			Draggable: PlayerDragComponent{},
		}
	}

	player := resetGame()

	obstacle := rl.Rectangle{X: 200, Y: 400, Width: 800, Height: 50}

	rl.SetTargetFPS(60)

	for !rl.WindowShouldClose() {
		deltaTime := float64(rl.GetFrameTime())

		if rl.IsKeyPressed(rl.KeyR) {
			player = resetGame()
		}

		if rl.IsKeyPressed(rl.KeyEscape) {
			break
		}

		ProcessMouseDrag(&player, deltaTime) // Updated to apply acceleration

		player.Physics.Update(&player, deltaTime, []rl.Rectangle{obstacle})

		rl.BeginDrawing()
		rl.ClearBackground(rl.RayWhite)

		// Draw player
		rl.DrawRectangleRec(player.Body, rl.Red)

		// Draw obstacle
		rl.DrawRectangleRec(obstacle, rl.Blue)

		rl.EndDrawing()
	}
}
