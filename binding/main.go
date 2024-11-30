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
	"unsafe"

	"github.com/plugfox/slash-engine-go/pkg"
)

//nolint:gochecknoglobals
var manager = &pkg.WorldManager{} // Create a global instance of WorldManager for the C API

// Initialize the world with options
//
//export InitWorld
func InitWorld(
	gravityX C.double,
	gravityY C.double,
	boundaryX C.double,
	boundaryY C.double,
	tickMS C.double,
	rtt C.double,
	autoStart C.int,
) {
	manager.InitWorld(
		float64(gravityX),
		float64(gravityY),
		float64(boundaryX),
		float64(boundaryY),
		float64(tickMS),
		float64(rtt),
		autoStart != 0, // Convert C.int to Go bool
	)
}

// Add or update objects
//
//export UpsertObjects
func UpsertObjects(cObjects *C.Object, count C.int) {
	objects := make([]*pkg.Object, count)
	for idx := range objects {
		obj := (*C.Object)(unsafe.Pointer(uintptr(unsafe.Pointer(cObjects)) + uintptr(idx)*unsafe.Sizeof(*cObjects)))
		objects[idx] = &pkg.Object{
			ID:       int(obj.id),
			Position: pkg.Vector{X: float64(obj.posX), Y: float64(obj.posY)},
			Velocity: pkg.Vector{X: float64(obj.velX), Y: float64(obj.velY)},
			Mass:     float64(obj.mass),
		}
	}
	manager.UpsertObjects(objects)
}

// Delete objects by IDs
//
//export DeleteObjects
func DeleteObjects(ids *C.int, count C.int) {
	// Safely convert the C array to a Go slice
	idSlice := unsafe.Slice((*C.int)(unsafe.Pointer(ids)), count)

	// Convert C.int slice directly to []int without extra copying
	toDelete := make([]int, count)
	for i := range toDelete {
		toDelete[i] = int(idSlice[i])
	}

	// Call manager to delete objects
	manager.DeleteObjects(toDelete)
}

// Apply impulse to an object
//
//export ApplyImpulse
func ApplyImpulse(objectID C.int, impulseX C.double, impulseY C.double) {
	manager.ApplyImpulse(int(objectID), float64(impulseX), float64(impulseY))
}

// Get all object positions
//
//export GetObjectPositions
func GetObjectPositions() *C.Object {
	objects := manager.GetObjectPositions()
	count := len(objects)
	cObjects := C.malloc(C.size_t(count) * C.size_t(C.sizeof_Object))

	i := 0
	for _, obj := range objects {
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
	manager.StopWorld()
}

func main() {}
