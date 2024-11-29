package main

/*
#include <stdint.h>
#include <stdlib.h>
typedef struct {
    int id;
    double posX;
    double posY;
    double velX;
    double velY;
    double mass;
} Object;
*/
import "C"

import (
	"sync"
	"time"
	"unsafe"
)

type Vector struct {
	X, Y float64
}

type Object struct {
	ID            int
	Position      Vector
	Velocity      Vector
	Mass          float64
	LastServerPos Vector
	LastUpdate    time.Time
}

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

type WorldManager struct {
	once  sync.Once
	world *World
}

//nolint:gochecknoglobals
var manager = &WorldManager{}

func (wm *WorldManager) GetWorld() *World {
	return wm.world
}

func (wm *WorldManager) InitWorldWithOptions(gravityX, gravityY, boundaryX, boundaryY, tickMS, rtt float64, autoStart bool) {
	wm.once.Do(func() {
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
//
//export UpsertObjects
func UpsertObjects(cObjects *C.Object, count C.int) {
	world := manager.GetWorld()
	objects := make([]*Object, count)
	for idx := range objects {
		obj := (*C.Object)(unsafe.Pointer(uintptr(unsafe.Pointer(cObjects)) + uintptr(idx)*unsafe.Sizeof(*cObjects)))
		objects[idx] = &Object{
			ID:       int(obj.id),
			Position: Vector{X: float64(obj.posX), Y: float64(obj.posY)},
			Velocity: Vector{X: float64(obj.velX), Y: float64(obj.velY)},
			Mass:     float64(obj.mass),
		}
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
//
//export DeleteObjects
func DeleteObjects(ids *C.int, count C.int) {
	world := manager.GetWorld()
	idSlice := (*[1 << 30]C.int)(unsafe.Pointer(ids))[:count:count]
	world.mutex.Lock()
	defer world.mutex.Unlock()
	for _, id := range idSlice {
		delete(world.Objects, int(id))
	}
}

// Apply impulse to an object
//
//export ApplyImpulse
func ApplyImpulse(objectID C.int, impulseX, impulseY C.double) {
	world := manager.GetWorld()
	world.mutex.Lock()
	defer world.mutex.Unlock()
	if obj, ok := world.Objects[int(objectID)]; ok {
		obj.Velocity.X += float64(impulseX) / obj.Mass
		obj.Velocity.Y += float64(impulseY) / obj.Mass
	}
}

// Get all object positions
//
//export GetObjectPositions
func GetObjectPositions() *C.Object {
	world := manager.GetWorld()
	world.mutex.RLock()
	defer world.mutex.RUnlock()

	count := len(world.Objects)
	cObjects := C.malloc(C.size_t(count) * C.size_t(C.sizeof_Object))

	i := 0
	for _, obj := range world.Objects {
		cObj := (*C.Object)(unsafe.Pointer(uintptr(cObjects) + uintptr(i)*C.sizeof_Object))
		cObj.id = C.int(obj.ID)
		cObj.posX = C.double(obj.Position.X)
		cObj.posY = C.double(obj.Position.Y)
		cObj.velX = C.double(obj.Velocity.X)
		cObj.velY = C.double(obj.Velocity.Y)
		cObj.mass = C.double(obj.Mass)
		i++
	}

	return (*C.Object)(cObjects)
}

// Stop and clear the world
//
//export StopWorld
func StopWorld() {
	world := manager.GetWorld()
	world.mutex.Lock()
	defer world.mutex.Unlock()
	if world.running {
		world.running = false
		close(world.stopChannel)
		world.updateTicker.Stop()
	}
	world.Objects = make(map[int]*Object)
}

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

func clamp(val, minValue, maxValue float64) float64 {
	if val < minValue {
		return minValue
	}
	if val > maxValue {
		return maxValue
	}
	return val
}

func main() {}
