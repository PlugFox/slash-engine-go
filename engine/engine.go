package engine

import (
	"fmt"
	"sync"
	"time"
)

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
	return engine.getWorld()
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
	engine.lastUpdate = time.Now()
	engine.updateTicker = time.NewTicker(time.Duration(tickMS) * time.Millisecond)
	engine.mutex.Unlock()

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
		engine.update(rtt) // Extrapolate object positions based on RTT
	}
	engine.lastUpdate = time.Now() // Set last update time
}

// Get an object or particle by id
func (engine *Engine) GetObject(id int) *Object {
	engine.mutex.RLock()
	defer engine.mutex.RUnlock()
	return engine.getObject(id)
}

// Upsert an object to the world
func (engine *Engine) UpsertObject(obj *Object) {
	engine.mutex.Lock()
	defer engine.mutex.Unlock()
	world := engine.getWorld()
	if world != nil {
		world.Objects[obj.ID] = obj
	}
}

// Upsert objects to the world
func (engine *Engine) UpsertObjects(objects []*Object) {
	engine.mutex.Lock()
	defer engine.mutex.Unlock()
	world := engine.getWorld()
	if world != nil {
		for _, obj := range objects {
			world.Objects[obj.ID] = obj
		}
	}
}

// Add an impulse to an object
func (engine *Engine) AddImpulse(id int, direction Vector, damping float64) {
	engine.mutex.Lock()
	defer engine.mutex.Unlock()
	obj := engine.getObject(id)
	if obj != nil {
		newImpulse := &Impulse{
			Direction: direction,
			Damping:   damping,
			Next:      obj.Impulses,
		}
		obj.Impulses = newImpulse
	}
}

// Set the velocity of an object
func (engine *Engine) SetVelocity(id int, velocity Vector) {
	engine.mutex.Lock()
	defer engine.mutex.Unlock()
	obj := engine.getObject(id)
	if obj != nil {
		obj.Velocity = velocity
	}
}

// Set the position of an object
func (engine *Engine) SetPosition(id int, position Vector) {
	engine.mutex.Lock()
	defer engine.mutex.Unlock()
	obj := engine.getObject(id)
	if obj != nil {
		obj.Position = position
	}
}

// Set the anchor of an object
func (engine *Engine) SetAnchor(id int, anchor Vector) {
	engine.mutex.Lock()
	defer engine.mutex.Unlock()
	obj := engine.getObject(id)
	if obj != nil {
		obj.Anchor = anchor
	}
}

// Remove object by ID
func (engine *Engine) RemoveObject(id int) {
	engine.mutex.Lock()
	defer engine.mutex.Unlock()
	world := engine.getWorld()
	if world != nil {
		delete(world.Objects, id)
	}
}

// Remove objects by IDs
func (engine *Engine) RemoveObjects(ids []int) {
	engine.mutex.Lock()
	defer engine.mutex.Unlock()
	world := engine.getWorld()
	if world != nil {
		for _, id := range ids {
			delete(world.Objects, id)
		}
	}
}

// -- Internal methods -- //

// Get the world instance, can be nil
func (engine *Engine) getWorld() *World {
	return engine.world
}

// Get the object from the world
func (engine *Engine) getObject(id int) *Object {
	world := engine.getWorld()
	if world == nil {
		return nil
	}
	return world.Objects[id]
}
