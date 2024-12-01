package main

/*
#include <stdint.h>
#include <string.h>
#include <stdlib.h>

typedef enum {
    Other,      // Unknown object type
    Creature,   // Living entity
    Projectile, // Moving entity that can hit others
    Effect,     // Visual effect
    Terrain,    // Static blocking entity
    Structure,  // Static structure
    Item        // Static entity that can be picked up
} ObjectType;

typedef struct {
    double X;
    double Y;
} Vector;

typedef struct Impulse {
    Vector Direction;    // Direction and magnitude of the impulse
    double Damping;      // Damping factor
    struct Impulse* Next; // Pointer to the next impulse in the list
} Impulse;

typedef struct {
    int32_t ID;          // Object ID
    ObjectType Type;     // Type of the object
    uint8_t Client;      // Created by client (1) or server (0)
    Vector Size;         // Current object size (width, height)
    Vector Velocity;     // Current velocity (x, y)
    Vector Position;     // Current position (x, y)
    Vector Anchor;       // Anchor position relative to the object's center
    double GravityFactor; // Gravity factor
    Impulse* Impulses;   // Linked list of active impulses
} Object;

typedef struct {
    double Gravity;
    Vector Boundary;
    Object* Objects;
    int32_t ObjectCount;
} World;

// Forward declarations for exporting
void CreateWorld(double gravity, Vector boundary);
void SetWorld(World* world, double rtt);
void Run(double tickMS);
void Stop();
World* GetWorldPtr();
uint8_t* GetWorldBytes(int32_t* size);
Object* GetObjectPtr(int32_t id);
void UpsertObject(Object* obj);
void UpsertObjects(Object* objects, int32_t count);
void AddImpulse(int32_t id, Vector direction, double damping);
void SetVelocity(int32_t id, Vector velocity);
void SetPosition(int32_t id, Vector position);
void SetAnchor(int32_t id, Vector anchor);
void RemoveObject(int32_t id);
void RemoveObjects(int32_t* ids, int32_t count);
void FreeImpulsePtr(Impulse* impulse);
void FreeObjectPtr(Object* obj);
void FreeWorldPtr(World* world);
*/
import "C"

import (
	"unsafe"

	"github.com/plugfox/slash-engine-go/engine"
)

//nolint:gochecknoglobals
var singleton = &engine.Engine{} // Create a global instance of Engine for the C API

//export CreateWorld
func CreateWorld(gravity C.double, boundary C.Vector) {
	goBoundary := engine.Vector{X: float64(boundary.X), Y: float64(boundary.Y)}
	singleton.CreateWorld(float64(gravity), goBoundary)
}

//export SetWorld
func SetWorld(world *C.World, rtt C.double) {
	goRTT := float64(rtt)
	if world == nil {
		singleton.SetWorld(nil, goRTT)
		return
	}
	goWorld := _convertWorldToGo(world) // Преобразуем C-мир в Go-мир
	singleton.SetWorld(goWorld, goRTT)  // Устанавливаем преобразованный мир в движок
}

//export Run
func Run(tickMS C.double) {
	singleton.Run(float64(tickMS))
}

//export Stop
func Stop() {
	singleton.Stop()
}

//export GetWorldPtr
func GetWorldPtr() *C.World {
	// Получаем указатель на текущий мир
	world := singleton.GetWorld()
	if world == nil {
		return nil
	}

	// Преобразуем Go-мир в C-мир
	return _convertWorldToC(world)
}

//export GetWorldBytes
func GetWorldBytes(size *C.int32_t) *C.uint8_t {
	// Получаем объект world
	world := singleton.GetWorld()
	if world == nil {
		// Если world равен nil, возвращаем null
		if size != nil {
			*size = 0
		}
		return nil
	}

	// Сериализуем мир в байты
	data := world.ToBytes()
	dataSize := len(data)

	// Выделяем память для массива байт
	cData := C.malloc(C.size_t(dataSize))
	if cData == nil {
		// Если память не выделилась, возвращаем null
		if size != nil {
			*size = 0
		}
		return nil
	}

	// Копируем данные в выделенную память
	C.memcpy(cData, unsafe.Pointer(&data[0]), C.size_t(dataSize))

	// Возвращаем размер и указатель
	if size != nil {
		*size = C.int32_t(dataSize)
	}
	return (*C.uint8_t)(cData)
}

//export GetObjectPtr
func GetObjectPtr(id C.int32_t) *C.Object {
	goObj := singleton.GetObject(int(id))
	return _convertObjectToC(goObj)
}

//export UpsertObject
func UpsertObject(obj *C.Object) {
	goObj := _convertObjectToGo(obj)
	singleton.UpsertObject(goObj)
}

//export UpsertObjects
func UpsertObjects(objects *C.Object, count C.int32_t) {
	goObjects := make([]*engine.Object, count)
	objSlice := (*[1 << 30]C.Object)(unsafe.Pointer(objects))[:count:count]
	for i, obj := range objSlice {
		goObjects[i] = _convertObjectToGo(&obj)
	}
	singleton.UpsertObjects(goObjects)
}

//export AddImpulse
func AddImpulse(id C.int32_t, direction C.Vector, damping C.double) {
	goDirection := engine.Vector{X: float64(direction.X), Y: float64(direction.Y)}
	singleton.AddImpulse(int(id), goDirection, float64(damping))
}

//export SetVelocity
func SetVelocity(id C.int32_t, velocity C.Vector) {
	goVelocity := engine.Vector{X: float64(velocity.X), Y: float64(velocity.Y)}
	singleton.SetVelocity(int(id), goVelocity)
}

//export SetPosition
func SetPosition(id C.int32_t, position C.Vector) {
	goPosition := engine.Vector{X: float64(position.X), Y: float64(position.Y)}
	singleton.SetPosition(int(id), goPosition)
}

//export SetAnchor
func SetAnchor(id C.int32_t, anchor C.Vector) {
	goAnchor := engine.Vector{X: float64(anchor.X), Y: float64(anchor.Y)}
	singleton.SetAnchor(int(id), goAnchor)
}

//export RemoveObject
func RemoveObject(id C.int32_t) {
	singleton.RemoveObject(int(id))
}

//export RemoveObjects
func RemoveObjects(ids *C.int32_t, count C.int32_t) {
	goIDs := make([]int, count)
	idSlice := (*[1 << 30]C.int32_t)(unsafe.Pointer(ids))[:count:count]
	for i, id := range idSlice {
		goIDs[i] = int(id)
	}
	singleton.RemoveObjects(goIDs)
}

//export FreeImpulsePtr
func FreeImpulsePtr(cImpulse *C.Impulse) {
	_freeImpulse(cImpulse)
}

//export FreeObjectPtr
func FreeObjectPtr(cObj *C.Object) {
	_freeObject(cObj)
}

//export FreeWorldPtr
func FreeWorldPtr(cWorld *C.World) {
	_freeWorld(cWorld)
}

// Helper function to convert bool to uint8
func _boolToUint8(b bool) C.uint8_t {
	if b {
		return 1
	}
	return 0
}

// Helper function to convert uint8 to bool
func _uint8ToBool(b C.uint8_t) bool {
	return b != 0
}

// Вспомогательная функция для преобразования связного списка импульсов
func _convertImpulsesToC(goImpulse *engine.Impulse) *C.Impulse {
	if goImpulse == nil {
		return nil
	}

	// Рекурсивное преобразование импульсов
	cImpulse := (*C.Impulse)(C.malloc(C.size_t(C.sizeof_Impulse)))
	cImpulse.Direction = C.Vector{
		X: C.double(goImpulse.Direction.X),
		Y: C.double(goImpulse.Direction.Y),
	}
	cImpulse.Damping = C.double(goImpulse.Damping)
	cImpulse.Next = _convertImpulsesToC(goImpulse.Next)

	return cImpulse
}

// Вспомогательная функция для преобразования C-списка импульсов в Go-список
func _convertImpulsesToGo(cImpulse *C.Impulse) *engine.Impulse {
	if cImpulse == nil {
		return nil
	}

	// Рекурсивное преобразование импульсов
	goImpulse := &engine.Impulse{
		Direction: engine.Vector{
			X: float64(cImpulse.Direction.X),
			Y: float64(cImpulse.Direction.Y),
		},
		Damping: float64(cImpulse.Damping),
		Next:    _convertImpulsesToGo(cImpulse.Next),
	}

	return goImpulse
}

// Converts a Go Object to a C Object
func _convertObjectToC(obj *engine.Object) *C.Object {
	if obj == nil {
		return nil
	}

	cObj := (*C.Object)(C.malloc(C.size_t(C.sizeof_Object)))

	// Convert basic fields
	cObj.ID = C.int32_t(obj.ID)
	cObj.Type = C.ObjectType(obj.Type)
	cObj.Client = C.uint8_t(_boolToUint8(obj.Client))
	cObj.Size = C.Vector{
		X: C.double(obj.Size.X),
		Y: C.double(obj.Size.Y),
	}
	cObj.Velocity = C.Vector{
		X: C.double(obj.Velocity.X),
		Y: C.double(obj.Velocity.Y),
	}
	cObj.Position = C.Vector{
		X: C.double(obj.Position.X),
		Y: C.double(obj.Position.Y),
	}
	cObj.Anchor = C.Vector{
		X: C.double(obj.Anchor.X),
		Y: C.double(obj.Anchor.Y),
	}
	cObj.GravityFactor = C.double(obj.GravityFactor)

	// Convert impulses
	cObj.Impulses = _convertImpulsesToC(obj.Impulses)

	return cObj
}

// Converts a C Object to a Go Object
func _convertObjectToGo(cObj *C.Object) *engine.Object {
	if cObj == nil {
		return nil
	}

	obj := &engine.Object{
		ID:            int(cObj.ID),
		Type:          engine.ObjectType(cObj.Type),
		Client:        _uint8ToBool(cObj.Client),
		Size:          engine.Vector{X: float64(cObj.Size.X), Y: float64(cObj.Size.Y)},
		Velocity:      engine.Vector{X: float64(cObj.Velocity.X), Y: float64(cObj.Velocity.Y)},
		Position:      engine.Vector{X: float64(cObj.Position.X), Y: float64(cObj.Position.Y)},
		Anchor:        engine.Vector{X: float64(cObj.Anchor.X), Y: float64(cObj.Anchor.Y)},
		GravityFactor: float64(cObj.GravityFactor),
	}

	// Convert impulses
	obj.Impulses = _convertImpulsesToGo(cObj.Impulses)

	return obj
}

func _convertWorldToC(world *engine.World) *C.World {
	if world == nil {
		return nil
	}

	// Allocate memory for C.World
	cWorld := (*C.World)(C.malloc(C.size_t(C.sizeof_World)))

	// Convert fields
	cWorld.Gravity = C.double(world.Gravity)
	cWorld.Boundary = C.Vector{
		X: C.double(world.Boundary.X),
		Y: C.double(world.Boundary.Y),
	}

	// Convert map of objects to C array
	objectCount := len(world.Objects)
	cWorld.ObjectCount = C.int32_t(objectCount)

	if objectCount > 0 {
		cWorld.Objects = (*C.Object)(C.malloc(C.size_t(objectCount) * C.size_t(C.sizeof_Object)))

		// Convert Go objects to C objects
		cObjects := (*[1 << 30]C.Object)(unsafe.Pointer(cWorld.Objects))[:objectCount:objectCount]

		i := 0
		for _, obj := range world.Objects {
			cObjects[i] = *_convertObjectToC(obj)
			i++
		}
	} else {
		cWorld.Objects = nil
	}

	return cWorld
}

func _convertWorldToGo(cWorld *C.World) *engine.World {
	if cWorld == nil {
		return nil
	}

	// Create a new Go World
	world := &engine.World{
		Gravity:  float64(cWorld.Gravity),
		Boundary: engine.Vector{X: float64(cWorld.Boundary.X), Y: float64(cWorld.Boundary.Y)},
		Objects:  make(map[int]*engine.Object),
	}

	// Convert C array of objects to Go map
	objectCount := int(cWorld.ObjectCount)
	if objectCount > 0 && cWorld.Objects != nil {
		cObjects := (*[1 << 30]C.Object)(unsafe.Pointer(cWorld.Objects))[:objectCount:objectCount]

		for i := 0; i < objectCount; i++ {
			obj := _convertObjectToGo(&cObjects[i])
			if obj != nil {
				world.Objects[obj.ID] = obj
			}
		}
	}

	return world
}

func _freeImpulse(cImpulse *C.Impulse) {
	if cImpulse == nil {
		return
	}
	// Рекурсивное освобождение связного списка
	_freeImpulse(cImpulse.Next)
	C.free(unsafe.Pointer(cImpulse))
}

func _freeObject(cObj *C.Object) {
	if cObj == nil {
		return
	}
	// Освобождение импульсов объекта
	_freeImpulse(cObj.Impulses)
	C.free(unsafe.Pointer(cObj))
}

func _freeWorld(cWorld *C.World) {
	if cWorld == nil {
		return
	}
	// Освобождение массива объектов
	if cWorld.Objects != nil && cWorld.ObjectCount > 0 {
		cObjects := (*[1 << 30]C.Object)(unsafe.Pointer(cWorld.Objects))[:cWorld.ObjectCount:cWorld.ObjectCount]
		for i := 0; i < int(cWorld.ObjectCount); i++ {
			_freeObject(&cObjects[i])
		}
		C.free(unsafe.Pointer(cWorld.Objects))
	}
	C.free(unsafe.Pointer(cWorld))
}

func main() {}
