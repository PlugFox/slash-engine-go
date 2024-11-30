package pkg

import (
	"sync"
	"time"
)

// Vector represents a 2D vector
type Vector struct {
	X, Y float64
}

// Object represents a game object
type Object struct {
	ID            int
	Position      Vector
	Velocity      Vector
	Mass          float64
	LastServerPos Vector
	LastUpdate    time.Time
}

// World represents the game world
type World struct {
	Gravity      Vector
	Boundary     Vector
	RTT          float64
	Objects      map[int]*Object
	mutex        sync.RWMutex
	running      bool
	stopChannel  chan struct{}
	updateTicker *time.Ticker
	updateSignal chan struct{}
}

// WorldManager manages the world instance
type WorldManager struct {
	once  sync.Once
	world *World
}

// Get the world instance, can be nil
func (wm *WorldManager) GetWorld() *World {
	return wm.world
}

// Initialize the world with options
func (wm *WorldManager) InitWorld(
	gravityX float64,
	gravityY float64,
	boundaryX float64,
	boundaryY float64,
	tickMS float64,
	rtt float64,
	autoStart bool,
) {
	wm.once.Do(func() {
		wm.StopWorld() // Stop and clear the world
		wm.world = &World{
			Gravity:      Vector{X: gravityX, Y: gravityY},
			Boundary:     Vector{X: boundaryX, Y: boundaryY},
			RTT:          rtt,
			Objects:      make(map[int]*Object),
			stopChannel:  make(chan struct{}),
			updateSignal: make(chan struct{}, 1),
		}
		if autoStart {
			go wm.world.run(tickMS)
		}
	})
}

// Add or update objects
func (wm *WorldManager) UpsertObjects(objects []*Object) {
	world := wm.GetWorld()
	if world == nil {
		return
	}
	world.mutex.Lock()
	defer world.mutex.Unlock()
	for _, obj := range objects {
		if existingObj, ok := world.Objects[obj.ID]; ok {
			// Update existing object
			existingObj.Position = obj.Position
			existingObj.Velocity = obj.Velocity
			existingObj.Mass = obj.Mass
			existingObj.LastServerPos = obj.Position
			existingObj.LastUpdate = time.Now()
		} else {
			// Add new object
			obj.LastUpdate = time.Now()
			world.Objects[obj.ID] = obj
		}
	}
}

// Delete objects by IDs
func (wm *WorldManager) DeleteObjects(ids []int) {
	world := wm.GetWorld()
	if world == nil {
		return
	}
	world.mutex.Lock()
	defer world.mutex.Unlock()
	for _, id := range ids {
		delete(world.Objects, id)
	}
}

// Apply impulse to an object
func (wm *WorldManager) ApplyImpulse(objectID int, impulseX float64, impulseY float64) {
	world := wm.GetWorld()
	if world == nil {
		return
	}
	world.mutex.Lock()
	defer world.mutex.Unlock()
	if obj, ok := world.Objects[objectID]; ok {
		obj.Velocity.X += impulseX / obj.Mass
		obj.Velocity.Y += impulseY / obj.Mass
	}
}

// Get all object positions
func (wm *WorldManager) GetObjectPositions() []Object {
	world := wm.GetWorld()
	if world == nil {
		return make([]Object, 0)
	}
	world.mutex.RLock()
	defer world.mutex.RUnlock()
	count := len(world.Objects)
	objects := make([]Object, count)
	idx := 0
	for _, obj := range world.Objects {
		objects[idx] = *obj
		idx++
	}
	return objects
}

// Stop and clear the world
func (wm *WorldManager) StopWorld() {
	world := wm.GetWorld()
	if world == nil {
		return
	}
	world.mutex.Lock()
	defer world.mutex.Unlock()
	if world.running {
		world.running = false
		close(world.stopChannel)
		world.updateTicker.Stop()
	}
	world.Objects = make(map[int]*Object)
}

// Run the world update loop
func (w *World) run(tickMS float64) {
	w.running = true
	w.updateTicker = time.NewTicker(time.Duration(tickMS) * time.Millisecond)
	defer w.updateTicker.Stop()

	for {
		select {
		case <-w.updateTicker.C:
			select {
			case w.updateSignal <- struct{}{}:
			default: // Skip if already updating
			}
		case <-w.updateSignal:
			w.update()
		case <-w.stopChannel:
			return
		}
	}
}

// Update object positions
func (w *World) update() {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	for _, obj := range w.Objects {
		// Extrapolate object position
		w.extrapolateObject(obj)

		// Apply gravity
		obj.Velocity.X += w.Gravity.X * obj.Mass
		obj.Velocity.Y += w.Gravity.Y * obj.Mass

		// Update position
		obj.Position.X += obj.Velocity.X
		obj.Position.Y += obj.Velocity.Y

		// Clamp to world boundaries
		obj.Position.X = clamp(obj.Position.X, 0, w.Boundary.X)
		obj.Position.Y = clamp(obj.Position.Y, 0, w.Boundary.Y)
	}
}

// Extrapolate object position based on velocity
func (w *World) extrapolateObject(obj *Object) {
	now := time.Now()
	elapsed := now.Sub(obj.LastUpdate).Seconds()
	if elapsed > 0 {
		totalElapsed := elapsed + w.RTT
		obj.Position.X += obj.Velocity.X * totalElapsed
		obj.Position.Y += obj.Velocity.Y * totalElapsed
		obj.LastUpdate = now
	}
}

// Clamp a value between a min and max
func clamp(val, minValue, maxValue float64) float64 {
	if val < minValue {
		return minValue
	}
	if val > maxValue {
		return maxValue
	}
	return val
}
