package engine

import "math"

type ObjectType int

const (
	// Other represents an unknown object type
	// Other is a generic entity that doesn't fit into any other category
	Other ObjectType = iota

	// Creature represents a creature object type
	// Creature is a living entity that can move and interact with the world
	// Can be a player, enemy, NPC, etc.
	//
	// Creature can have:
	// - health, mana, and other stats
	// - equipment, weapons, and armor
	// - skills, spells, and abilities
	// - AI, behavior, and pathfinding
	// - quests, dialogues, and storylines
	// - etc.
	//
	// Creature can't leave the world boundaries,
	// can't move through the floor, terrain or structure
	Creature

	// Projectile represents a projectile object type
	// Projectile is a moving entity that can hit other objects
	// Can be a bullet, arrow, missile, etc.
	//
	// Projectile can have:
	// - damage, speed, and range
	// - area of effect, penetration, and explosion
	// - target, homing, and tracking
	// - physics, gravity, and collisions
	// - etc.
	//
	// Projectile can fly off the screen and be removed by the client
	// but can't move through the floor, terrain or structure
	Projectile

	// Particle represents a particle object type
	// Particle is a visual effect that can move and animate
	// Can be a smoke, fire, explosion, etc.
	//
	// Particle can have:
	// - texture, color, and size
	// - animation, rotation, and scaling
	// - fading, blending, and transparency
	// - physics, gravity, and collisions
	// - etc.
	//
	// Particle can fly off the screen and be removed by the client
	// and can move through the floor, terrain or structure
	Effect

	// Terrain represents a terrain object type
	// Terrain is a static entity that can block other objects
	// Can be a ground, wall, platform, etc.
	//
	// Terrain can have:
	// - texture, color, and size
	// - collision, friction, and bounciness
	// - physics, gravity, and collisions
	// - etc.
	//
	// Terrain can't leave the world boundaries,
	// can't move through the floor, terrain or structure
	// and can block creatures, projectiles, and particles
	// Terrain can be used for collision detection and response
	Terrain

	// Structure represents a structure object type
	// Structure is a static entity that can block other objects
	// Can be a building, house, tree, etc.
	//
	// Structure can have:
	// - texture, color, and size
	// - collision, friction, and bounciness
	// - physics, gravity, and collisions
	// - etc.
	//
	// Structure can't leave the world boundaries,
	// can't move through the floor, terrain or structure
	// and can block creatures, projectiles, and particles
	// Structure can be used for collision detection and response
	Structure

	// Item represents an item object type
	// Item is a static entity that can be picked up by creatures
	// Can be a weapon, armor, potion, etc.
	//
	// Item can have:
	// - texture, color, and size
	// - stats, bonuses, and effects
	// - rarity, value, and weight
	// - etc.
	//
	// Item can't leave the world boundaries,
	// can't move through the floor, terrain or structure
	// and can be picked up by creatures
	// Item can be used for inventory, equipment, and looting
	Item
)

// Vector represents a 2D vector
type Vector struct {
	X, Y float64
}

// Impulse represents a single impulse affecting an object
// Влияние разных значений Damping:
// >1.0	     | Увеличение импульса.	Используется редко, например, для ускорения ракет.
// 1.0	     | Нет затухания, импульс постоянный.	Редко используется, например, для постоянного ускорения.
// 0.95-0.99 | Медленное затухание.	Стрела, плавное движение через сопротивление.
// 0.8-0.9	 | Умеренное затухание.	Прыжок персонажа, разлетающиеся осколки.
// 0.5	     | Быстрое затухание.	Эффекты, исчезающие почти сразу, например, магические частицы.
// 0.1-0.2	 | Очень быстрое затухание.	Используется для взрывов, ударов, отскоков.
// 0.0	     | Немедленное затухание.	Импульс исчезает сразу после применения.
type Impulse struct {
	Direction Vector   // Direction and magnitude of the impulse
	Damping   float64  // Damping factor
	Next      *Impulse // Pointer to the next impulse in the list
}

// Object represents a game object
//
// Difference between Object and Particle:
// - Object is always in the world, and can be removed only by the server
// - Particle can fly off the screen and be removed by the client
// - Anchor position for objects usually is the bottom center of the object
// - Anchor position for particles usually is the center of the particle
// - Usually ids of objects are unique, positive integers assigned by the server
// - Usually ids of particles are negative integers assigned by the client or server
type Object struct {
	ID            int        // ID represents the object ID
	Type          ObjectType // Type of the object
	Client        bool       // Client represents if the object is created by the client
	Size          Vector     // Size represents the current object size vector (width, height)
	Velocity      Vector     // Velocity represents the current object velocity vector (x, y)
	Position      Vector     // Position represents the current object's center position vector (x, y)
	Anchor        Vector     // Anchor represents the anchor position for the object from the center of the object
	GravityFactor float64    // Gravity factor (0 = no grav, 1 = full, 2 = double, -1 = reverse, etc.)
	Impulses      *Impulse   // Linked list of active impulses
}

// World represents the game world
type World struct {
	// Gravity of the world (m/s^2)
	// Positive value means gravity is pulling objects down
	// negative value means gravity is pulling objects up
	// By default, gravity is set to 9.81 m/s^2
	Gravity float64

	// Boundary represents the world boundaries (width and height)
	Boundary Vector

	// Objects is a map of major game objects (e.g. players, enemies, bullets)
	// Objects are always in the world, and can be removed only by the server
	// Anchor position for objects is the bottom center of the object
	// Usually ids of objects are unique, positive integers assigned by the server
	Objects map[int]*Object
}

// TODO: Add codec to FlatBuffers encoding and decoding for the engine's world

// -- Internal methods -- //

// Object center position X coordinate (center of the object)
func (obj *Object) positionCenterX() float64 {
	return obj.Position.X
}

// Object center position Y coordinate (center of the object)
func (obj *Object) positionCenterY() float64 {
	return obj.Position.Y
}

// Position of the object's anchor X coordinate
func (obj *Object) positionAnchorX() float64 {
	return obj.positionCenterX() + obj.Anchor.X
}

// Position of the object's anchor Y coordinate
func (obj *Object) positionAnchorY() float64 {
	return obj.positionCenterY() + obj.Anchor.Y
}

// Object bottom position Y coordinate
func (obj *Object) positionBottomY() float64 {
	return obj.Position.Y - obj.Size.Y/2
}

// Object left position X coordinate
func (obj *Object) positionLeftX() float64 {
	return obj.Position.X - obj.Size.X/2
}

// Object right position X coordinate
func (obj *Object) positionRightX() float64 {
	return obj.Position.X + obj.Size.X/2
}

// Object is on the floor (bottom of the object is touching the floor)
func (obj *Object) onTheFloor() bool {
	return math.Abs(obj.positionBottomY()) < negligibleFloat
}

// Object moving upward (velocity is positive in the upward direction)
func (obj *Object) movingUpward() bool {
	return obj.Velocity.Y > 0
}

// Object moving downward (velocity is negative in the downward direction)
func (obj *Object) movingDownward() bool {
	return obj.Velocity.Y < 0
}

// Object moving leftward (velocity is negative in the left direction)
func (obj *Object) movingLeftward() bool {
	return obj.Velocity.X < 0
}

// Object moving rightward (velocity is positive in the right direction)
func (obj *Object) movingRightward() bool {
	return obj.Velocity.X > 0
}

// Is the impulse push in the upward direction, pushing the object up
func (imp *Impulse) isUpward() bool {
	return imp.Direction.Y > 0
}

// Is the impulse push in the downward direction, pushing the object down
func (imp *Impulse) isDownward() bool {
	return imp.Direction.Y < 0
}

// Is the impulse push in the left direction, pushing the object left
func (imp *Impulse) isLeftward() bool {
	return imp.Direction.X < 0
}

// Is the impulse push in the right direction, pushing the object right
func (imp *Impulse) isRightward() bool {
	return imp.Direction.X > 0
}
