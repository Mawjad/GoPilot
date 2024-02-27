package main

import (
    "math"

    rl "github.com/gen2brain/raylib-go/raylib"
)

const (
    MaxVelocity       float32 = 32.0
    Gravity           float32 = 9.8
    PlayerAcceleration float32 = 3.2
    FrictionCoefficient float32 = 1.
	ElasticityCoefficient float32 = 0.5
	screenWidth int32 = 1200
	screenHeight int32 = 850
)

type Entity struct {
    Body      rl.Rectangle
    Physics   PhysicsComponent
    Draggable DragComponent
}

type PhysicsComponent struct {
    Velocity     rl.Vector2
    Acceleration rl.Vector2
    Gravity      float32
    Mass         float32
    Friction     float32
	Elasticity   float32
}

type DragComponent struct {
    Dragging    bool
    MouseOffset rl.Vector2
}

func NewPhysicsComponent(mass float32) PhysicsComponent {
    return PhysicsComponent{
        Velocity:     rl.Vector2{X: 0, Y: 0},
        Acceleration: rl.Vector2{X: 0, Y: 0},
        Gravity:      Gravity,
        Mass:         mass,
        Friction:     FrictionCoefficient,
		Elasticity:   ElasticityCoefficient,
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

    centerXEntity := entityBody.X + entityBody.Width/2
    centerYEntity := entityBody.Y + entityBody.Height/2
    centerXObstacle := obstacle.X + obstacle.Width/2
    centerYObstacle := obstacle.Y + obstacle.Height/2

    horizontalDist := centerXEntity - centerXObstacle
    verticalDist := centerYEntity - centerYObstacle

    if math.Abs(float64(horizontalDist)) > math.Abs(float64(verticalDist)) {
        if horizontalDist > 0 {
            return true, "right"
        } else {
            return true, "left"
        }
    } else {
        if verticalDist > 0 {
            return true, "bottom"
        } else {
            return true, "top"
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
        body.X = obstacle.X + obstacle.Width
    case "right":
        body.X = obstacle.X - body.Width
    case "top":
        body.Y = obstacle.Y + obstacle.Height
    case "bottom":
        body.Y = obstacle.Y - body.Height
    }
}

func (pc *PhysicsComponent) Update(entity *Entity, deltaTime float64, obstacles []rl.Rectangle) {
    pc.ApplyGravity()

    // Temporarily calculate the new position without applying friction globally
    tempVelocity := pc.Velocity
    tempVelocity.X += pc.Acceleration.X * float32(deltaTime)
    tempVelocity.Y += pc.Acceleration.Y * float32(deltaTime)

    tempBody := rl.Rectangle{
        X: entity.Body.X + tempVelocity.X * float32(deltaTime),
        Y: entity.Body.Y + tempVelocity.Y * float32(deltaTime),
        Width: entity.Body.Width,
        Height: entity.Body.Height,
    }

    collisionOccurred := false

    for _, obstacle := range obstacles {
        collides, side := GetCollisionSide(tempBody, obstacle)
        if collides {
            collisionOccurred = true

            // Apply friction in response to the collision
            if side == "left" || side == "right" {
                tempVelocity.X = -tempVelocity.X * pc.Elasticity // Bounce effect with elasticity
                tempVelocity.X *= 1 - pc.Friction // Apply friction horizontally
            }
            if side == "top" || side == "bottom" {
                tempVelocity.Y = -tempVelocity.Y * pc.Elasticity // Bounce effect with elasticity
                tempVelocity.Y *= 1 - pc.Friction // Apply friction vertically
            }

            // Adjust position based on collision side
            adjustPositionForCollision(&entity.Body, side, obstacle)
            break // Assuming handling one collision at a time
        }
    }

    // If no collision occurred, update the entity's velocity
    if !collisionOccurred {
        pc.Velocity = tempVelocity
    }

    // Finally, update the entity's position based on the adjusted velocity
    entity.Body.X += pc.Velocity.X * float32(deltaTime)
    entity.Body.Y += pc.Velocity.Y * float32(deltaTime)

	capVelocity(pc)

    // Reset acceleration after applying it
    pc.Acceleration = rl.Vector2{X: 0, Y: 0}
}


func main() {
	rl.InitWindow(screenWidth, screenHeight, "GoPilot ECS")
	defer rl.CloseWindow()

	playerMass := float32(1.0)
	player := Entity{
		Body: rl.Rectangle{X: 350, Y: 200, Width: 100, Height: 100},
		Physics: NewPhysicsComponent(playerMass, FrictionCoefficient),
		Draggable: DragComponent{
			Dragging:    false,
			MouseOffset: rl.Vector2{},
		},
	}

	obstacles := []rl.Rectangle{
		{X: 200, Y: 400, Width: 800, Height: 50},
		{X: 200, Y: 200, Width: 200, Height: 100},
	}

	rl.SetTargetFPS(60)

	for !rl.WindowShouldClose() {
		deltaTime := rl.GetFrameTime() // Use GetFrameTime for deltaTime

		ProcessDrag(&player)

		player.Physics.Update(&player, float64(deltaTime), obstacles)

		rl.BeginDrawing()
		rl.ClearBackground(rl.RayWhite)
		rl.DrawRectangleRec(player.Body, rl.Red)
		for _, obstacle := range obstacles {
			rl.DrawRectangleRec(obstacle, rl.Blue)
		}
		rl.EndDrawing()
	}
}