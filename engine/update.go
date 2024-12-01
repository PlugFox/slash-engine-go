package engine

import "math"

// Calculate physics and update object positions
func (engine *Engine) update(elapsed float64) {
	if elapsed <= 0 {
		return // Skip if no time has passed
	}

	world := engine.world
	if world == nil {
		return
	}

	// Update positions of all objects
	for _, obj := range world.Objects {
		switch obj.Type {
		case Projectile:
			obj._updateProjectile(world, elapsed)
		case Effect:
			obj._updateEffect(world, elapsed)
		case Creature:
			obj._updateCreature(world, elapsed)
		case Item:
			obj._updateItem(world, elapsed)
		case Structure:
			obj._updateStructure(world, elapsed)
		case Terrain:
			obj._updateTerrain(world, elapsed)
		case Other:
			obj._updateOther(world, elapsed)
		}
	}

	// TODO: Add collision detection and response here
	// impulse damping during collisions
}

// Apply gravity to an object
func _applyGravity(obj *Object, gravity float64) {
	if obj.GravityFactor != 0 {
		obj.Velocity.Y += -gravity * obj.GravityFactor
	}
}

// Apply impulses to an object with elapsed time
func _applyImpulses(obj *Object, elapsed float64) {
	const negligibleImpulse = negligibleFloat // Threshold for removing negligible impulses

	var prev *Impulse
	current := obj.Impulses

	for current != nil {
		// Apply impulse to velocity, scaled by elapsed time
		obj.Velocity.X += current.Direction.X * elapsed
		obj.Velocity.Y += current.Direction.Y * elapsed

		// Apply damping to the impulse based on elapsed time
		damping := current.Damping // Damping factor
		switch {
		case damping == 1:
			// No damping
		case damping <= negligibleImpulse:
			// Immediate damping to zero
			current.Direction.X = 0
			current.Direction.Y = 0
		default:
			// Apply damping to the impulse direction
			current.Direction.X *= math.Pow(damping, elapsed)
			current.Direction.Y *= math.Pow(damping, elapsed)
		}

		// Check if the impulse has decayed to negligible values
		if math.Abs(current.Direction.X) < negligibleImpulse && math.Abs(current.Direction.Y) < negligibleImpulse {
			// Remove impulse from the list
			if prev == nil {
				obj.Impulses = current.Next
			} else {
				prev.Next = current.Next
			}
			current = current.Next
			continue
		}

		// Move to the next impulse
		prev = current
		current = current.Next
	}
}

// Extrapolate object position based on velocity and elapsed time
func _extrapolatePosition(obj *Object, elapsed float64) {
	obj.Position.X += obj.Velocity.X * elapsed
	obj.Position.Y += obj.Velocity.Y * elapsed
}

// Update projectiles (such as arrow) based on physics, gravity, and collisions
func (obj *Object) _updateProjectile(world *World, elapsed float64) {
	// Apply gravity
	_applyGravity(obj, world.Gravity)

	// Apply impulses with elapsed time
	_applyImpulses(obj, elapsed)

	// Extrapolate object position based on velocity
	_extrapolatePosition(obj, elapsed)

	// Stop object if it hits the ground and moving downward
	if obj.onTheFloor() && obj.movingDownward() {
		obj.Velocity.Y = 0
		obj.Position.Y = 0
	}
}

// Update effects and particles (such as explosion) based on physics, gravity, and collisions
func (obj *Object) _updateEffect(world *World, elapsed float64) {
	// Apply gravity
	_applyGravity(obj, world.Gravity)

	// Apply impulses with elapsed time
	_applyImpulses(obj, elapsed)

	// Extrapolate object position based on velocity
	_extrapolatePosition(obj, elapsed)
}

// Update creatures (such as player) based on physics, gravity, and collisions
func (obj *Object) _updateCreature(world *World, elapsed float64) {
	// Apply gravity
	_applyGravity(obj, world.Gravity)

	// Apply impulses with elapsed time
	_applyImpulses(obj, elapsed)

	// Extrapolate object position based on velocity
	_extrapolatePosition(obj, elapsed)

	// Clamp to world boundaries and stop object if it hits the ground or walls
	obj.Position.X = clamp(obj.Position.X, obj.Size.X/2, world.Boundary.X-obj.Size.X/2)
	obj.Position.Y = clamp(obj.Position.Y, 0, world.Boundary.Y-obj.Size.Y)

	// Stop object if it hits the ground and moving downward
	if obj.onTheFloor() && obj.movingDownward() {
		obj.Velocity.Y = 0
		obj.Position.Y = 0
	}
}

// Update items (such as coins) based on physics, gravity, and collisions
func (obj *Object) _updateItem(world *World, elapsed float64) {
	// Apply gravity
	_applyGravity(obj, world.Gravity)

	// Apply impulses with elapsed time
	_applyImpulses(obj, elapsed)

	// Extrapolate object position based on velocity
	obj.Position.X += obj.Velocity.X * elapsed
	obj.Position.Y += obj.Velocity.Y * elapsed

	// Clamp to world boundaries and stop object if it hits the ground or walls
	obj.Position.X = clamp(obj.Position.X, obj.Size.X/2, world.Boundary.X-obj.Size.X/2)
	obj.Position.Y = clamp(obj.Position.Y, 0, world.Boundary.Y-obj.Size.Y)

	// Stop object if it hits the ground and moving downward
	if obj.onTheFloor() && obj.movingDownward() {
		obj.Velocity.Y = 0
		obj.Position.Y = 0
	}
}

// Update structures (such as walls), no physics or gravity applied
func (obj *Object) _updateStructure(world *World, elapsed float64) {}

// Update terrain (such as ground), no physics or gravity applied
func (obj *Object) _updateTerrain(world *World, elapsed float64) {}

// Update unknown objects, no physics or gravity applied
func (obj *Object) _updateOther(world *World, elapsed float64) {}
