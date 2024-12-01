package engine

import (
	"fmt"
	"math"
	"sync"
	"time"
)

// Negligible float value for comparisons
// To check if a float is close to zero and can be considered zero
// For example to remove an impulse if it has decayed to negligible values
const negligibleFloat = 0.01

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
// - Anchor position for objects is the bottom center of the object
// - Anchor position for particles is the center of the object
// - Usually ids of objects are unique, positive integers assigned by the server
// - Usually ids of particles are negative integers assigned by the client or server
// - When an object hits the ground, it stops moving down
type Object struct {
	ID            int      // ID represents the object ID
	Size          Vector   // Size represents the current object size vector (width, height)
	Velocity      Vector   // Velocity represents the current object velocity vector (x, y)
	Position      Vector   // Position represents the current object position
	GravityFactor float64  // Gravity factor (0 = no grav, 1 = full, 2 = double, -1 = reverse, etc.)
	Impulses      *Impulse // Linked list of active impulses
	Particle      bool     // Particle flag (true = particle, false = object)
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

// Engine represents the game physics controller
type Engine struct {
	world        *World        // Game world instance
	mutex        sync.RWMutex  // Game engine mutex
	running      bool          // Running flag
	stopChannel  chan struct{} // Stop channel
	updateSignal chan struct{} // Update signal
	updateTicker *time.Ticker  // Update ticker
	lastUpdate   time.Time     // Last update time
}

// Get the world instance, can be nil
func (engine *Engine) GetWorld() *World {
	engine.mutex.RLock()
	defer engine.mutex.RUnlock()
	return engine.world
}

// Get an object or particle by id
func (engine *Engine) GetObject(id int) *Object {
	engine.mutex.RLock()
	defer engine.mutex.RUnlock()
	world := engine.world
	if world == nil {
		return nil
	}
	return world.Objects[id]
}

// Stop and clear the world
func (engine *Engine) Stop() {
	engine.mutex.Lock()
	defer engine.mutex.Unlock()
	if !engine.running {
		return
	}
	engine.running = false
	close(engine.stopChannel)  // Сигнал для завершения всех горутин
	engine.updateTicker.Stop() // Остановка таймера
}

// Run the world update loop
func (engine *Engine) Run(tickMS float64) {
	if tickMS <= 0 {
		panic(fmt.Errorf("invalid tickMS: %v", tickMS))
	}

	engine.mutex.Lock()
	if engine.running || engine.world == nil {
		engine.mutex.Unlock()
		return
	}
	engine.running = true
	engine.stopChannel = make(chan struct{})     // Создаём новый канал при запуске
	engine.updateSignal = make(chan struct{}, 1) // Обеспечиваем буфер
	engine.mutex.Unlock()

	engine.lastUpdate = time.Now()
	engine.updateTicker = time.NewTicker(time.Duration(tickMS) * time.Millisecond)

	// Горутина для обработки сигналов таймера
	go func() {
		defer engine.updateTicker.Stop()
		for {
			select {
			case <-engine.updateTicker.C:
				select {
				case engine.updateSignal <- struct{}{}:
				default: // Skip if already updating
				}
			case <-engine.stopChannel:
				return // Завершаем выполнение
			}
		}
	}()

	// Горутина для обработки обновлений мира
	go func() {
		for {
			select {
			case <-engine.updateSignal:
				func() {
					engine.mutex.Lock()
					defer engine.mutex.Unlock()
					now := time.Now()                               // Current time
					elapsed := now.Sub(engine.lastUpdate).Seconds() // Elapsed time since last update
					engine.update(elapsed)                          // Update the world
					engine.lastUpdate = now                         // Set last update time
				}()
			case <-engine.stopChannel:
				return // Завершаем выполнение
			}
		}
	}()
}

// Create a new world instance
func (engine *Engine) CreateWorld(gravity float64, boundary Vector) *World {
	engine.mutex.Lock()
	defer engine.mutex.Unlock()
	world := &World{
		Gravity:  gravity,
		Boundary: boundary,
		Objects:  make(map[int]*Object),
	}
	engine.world = world
	return world
}

// Set the world instance
// RTT (round-trip time) is the ping-pong time between client and server
// RTT is used for extrapolation to predict object positions
func (engine *Engine) SetWorld(world *World, rtt float64) {
	engine.mutex.Lock()
	defer engine.mutex.Unlock()
	engine.world = world
	if rtt > 0 {
		engine.update(rtt / 2) // Extrapolate object positions based on half RTT
	}
	engine.lastUpdate = time.Now() // Set last update time
}

// Add an impulse to an object
func (engine *Engine) AddImpulse(id int, direction Vector, damping float64) {
	obj := engine.GetObject(id)
	if obj == nil {
		return
	}

	engine.mutex.Lock()
	defer engine.mutex.Unlock()

	if damping <= negligibleFloat {
		// Immediate damping if damping is negligible or zero
		obj.Velocity.X += direction.X
		obj.Velocity.Y += direction.Y
		return
	}

	newImpulse := &Impulse{
		Direction: direction,
		Damping:   damping,
		Next:      obj.Impulses,
	}
	obj.Impulses = newImpulse
}

// Set the velocity of an object
func (engine *Engine) SetVelocity(id int, velocity Vector) {
	obj := engine.GetObject(id)
	if obj == nil {
		return
	}
	obj.Velocity = velocity
}

// Set the position of an object
func (engine *Engine) SetPosition(id int, position Vector) {
	obj := engine.GetObject(id)
	if obj == nil {
		return
	}
	obj.Position = position
}

// Remove objects by IDs
func (engine *Engine) RemoveObjects(ids []int) {
	engine.mutex.Lock()
	defer engine.mutex.Unlock()
	world := engine.world
	if world == nil {
		return
	}
	for _, id := range ids {
		delete(world.Objects, id)
	}
}

// TODO: Add codec to FlatBuffers encoding and decoding for the engine's world

// Object is on the floor
func (obj *Object) onTheFloor() bool {
	return math.Abs(obj.Position.Y) < negligibleFloat
}

// Apply an impulses to an object
func (obj *Object) applyImpulses(elapsed float64) {
	if elapsed <= 0 {
		return // Skip if no time has passed
	}

	const negligibleImpulse = negligibleFloat // Threshold for removing negligible impulses

	var prev *Impulse
	current := obj.Impulses

	for current != nil {
		// If object bump at floor, stop this impulse
		if obj.onTheFloor() && math.Abs(current.Direction.Y) < negligibleImpulse {
			current.Direction.Y = 0
		}

		// Apply impulse to velocity, scaled by elapsed time
		obj.Velocity.X += current.Direction.X * elapsed
		obj.Velocity.Y += current.Direction.Y * elapsed

		// Apply damping to the impulse based on elapsed time
		damping := current.Damping
		if damping <= negligibleImpulse {
			// Immediate damping
			current.Direction.X = 0
			current.Direction.Y = 0
		} else if damping == 1 {
			// No damping
		} else {
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

// Calculate physics and update object positions
func (engine *Engine) update(elapsed float64) {
	if elapsed <= 0 {
		return // Skip if no time has passed
	}

	world := engine.world
	if world == nil {
		return
	}

	// Update object positions
	// Objects are always in the world, and can be removed only by the server
	for _, obj := range world.Objects {
		// Apply gravity
		if obj.GravityFactor != 0 {
			obj.Velocity.Y += -world.Gravity * obj.GravityFactor
		}

		// Apply impulses with elapsed time
		obj.applyImpulses(elapsed)

		// Extrapolate object position based on velocity
		obj.Position.X += obj.Velocity.X * elapsed
		obj.Position.Y += obj.Velocity.Y * elapsed

		// TODO: Add collision detection and response here

		// Particles can fly off the screen and be removed by the client
		if !obj.Particle {
			// Clamp to world boundaries and stop object if it hits the ground or walls
			obj.Position.X = clamp(obj.Position.X, obj.Size.X/2, world.Boundary.X-obj.Size.X/2)
			obj.Position.Y = clamp(obj.Position.Y, 0, world.Boundary.Y-obj.Size.Y)

			// Stop object if it hits the ground
			if obj.onTheFloor() && (obj.Velocity.Y < negligibleFloat) {
				obj.Velocity.Y = 0
			}
		}
	}
}

// Clamp a value between a min and max
func clamp(val float64, minValue float64, maxValue float64) float64 {
	if val < minValue {
		return minValue
	}
	if val > maxValue {
		return maxValue
	}
	return val
}
