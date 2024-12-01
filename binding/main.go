package main

/*
#include <stdint.h>
#include <stdlib.h>

typedef struct {
    double X;
    double Y;
} Vector;

typedef struct Impulse {
    Vector Direction;
    double Damping;
    struct Impulse* Next;
} Impulse;

typedef struct {
    int32_t ID;
    Vector Size;
    Vector Velocity;
    Vector Position;
    double GravityFactor;
    Impulse* Impulses;
    uint8_t Particle;
} Object;

typedef struct {
    double Gravity;
    Vector Boundary;
    Object* Objects;
    int32_t ObjectCount;
} World;

// Forward declarations for exporting
World* GetWorld();
void StopEngine();
void RunEngine(double tickMS);
World* CreateWorld(double gravity, Vector boundary);
void SetWorld(World* world);
void SetRTT(double rtt);
void AddImpulse(int32_t id, Vector direction, double damping);
void SetVelocity(int32_t id, Vector velocity);
void SetPosition(int32_t id, Vector position);
void RemoveObjects(int32_t* ids, int32_t count);
*/
import "C"

import (
	"unsafe"

	"github.com/plugfox/slash-engine-go/engine"
)

//nolint:gochecknoglobals
var singleton = &engine.Engine{} // Create a global instance of Engine for the C API

//export GetWorld
func GetWorld() *C.World {
	// Получаем указатель на текущий мир
	world := singleton.GetWorld()
	if world == nil {
		return nil
	}

	// Создаем C-структуру World
	cWorld := (*C.World)(C.malloc(C.size_t(C.sizeof_World)))

	// Заполняем поля World
	cWorld.Gravity = C.double(world.Gravity)
	cWorld.Boundary = C.Vector{
		X: C.double(world.Boundary.X),
		Y: C.double(world.Boundary.Y),
	}

	// Преобразуем карту Objects в массив объектов
	objectCount := len(world.Objects)
	cWorld.ObjectCount = C.int32_t(objectCount)

	if objectCount > 0 {
		cWorld.Objects = (*C.Object)(C.malloc(C.size_t(objectCount) * C.size_t(C.sizeof_Object)))

		// Получаем массив C-объектов
		cObjects := (*[1 << 30]C.Object)(unsafe.Pointer(cWorld.Objects))[:objectCount:objectCount]

		i := 0
		for _, obj := range world.Objects {
			// Конвертируем Go-объект в C-объект
			cObjects[i] = C.Object{
				ID:            C.int32_t(obj.ID),
				Size:          C.Vector{X: C.double(obj.Size.X), Y: C.double(obj.Size.Y)},
				Velocity:      C.Vector{X: C.double(obj.Velocity.X), Y: C.double(obj.Velocity.Y)},
				Position:      C.Vector{X: C.double(obj.Position.X), Y: C.double(obj.Position.Y)},
				GravityFactor: C.double(obj.GravityFactor),
				Particle:      C.uint8_t(0),
			}

			// Если это частица, устанавливаем флаг
			if obj.Particle {
				cObjects[i].Particle = C.uint8_t(1)
			}

			// Преобразуем импульсы
			cObjects[i].Impulses = convertImpulsesToC(obj.Impulses)

			i++
		}
	} else {
		cWorld.Objects = nil
	}

	return cWorld
}

// Вспомогательная функция для преобразования связного списка импульсов
func convertImpulsesToC(goImpulse *engine.Impulse) *C.Impulse {
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
	cImpulse.Next = convertImpulsesToC(goImpulse.Next)

	return cImpulse
}

//export StopEngine
func StopEngine() {
	singleton.Stop()
}

//export RunEngine
func RunEngine(tickMS C.double) {
	singleton.Run(float64(tickMS))
}

//export CreateWorld
func CreateWorld(gravity C.double, boundary C.Vector) *C.World {
	goBoundary := engine.Vector{X: float64(boundary.X), Y: float64(boundary.Y)}
	world := singleton.CreateWorld(float64(gravity), goBoundary)

	cWorld := (*C.World)(C.malloc(C.size_t(C.sizeof_World)))
	cWorld.Gravity = C.double(world.Gravity)
	cWorld.Boundary = C.Vector{
		X: C.double(world.Boundary.X),
		Y: C.double(world.Boundary.Y),
	}
	cWorld.ObjectCount = C.int32_t(len(world.Objects))
	return cWorld
}

//export SetWorld
func SetWorld(world *C.World) {
	if world == nil {
		singleton.SetWorld(nil)
		return
	}

	// Преобразуем границы мира
	goBoundary := engine.Vector{
		X: float64(world.Boundary.X),
		Y: float64(world.Boundary.Y),
	}

	// Создаём новый объект мира
	goWorld := &engine.World{
		Gravity:  float64(world.Gravity),
		Boundary: goBoundary,
		Objects:  make(map[int]*engine.Object),
	}

	// Преобразуем C-объекты в Go-объекты
	if world.Objects != nil && world.ObjectCount > 0 {
		cObjects := (*[1 << 30]C.Object)(unsafe.Pointer(world.Objects))[:world.ObjectCount:world.ObjectCount]
		for _, cObj := range cObjects {
			// Преобразуем каждый C-объект
			goObj := &engine.Object{
				ID:            int(cObj.ID),
				Size:          engine.Vector{X: float64(cObj.Size.X), Y: float64(cObj.Size.Y)},
				Velocity:      engine.Vector{X: float64(cObj.Velocity.X), Y: float64(cObj.Velocity.Y)},
				Position:      engine.Vector{X: float64(cObj.Position.X), Y: float64(cObj.Position.Y)},
				GravityFactor: float64(cObj.GravityFactor),
				Particle:      cObj.Particle != 0,
				Impulses:      convertImpulsesToGo(cObj.Impulses),
			}
			goWorld.Objects[goObj.ID] = goObj
		}
	}

	// Устанавливаем преобразованный мир в движок
	singleton.SetWorld(goWorld)
}

// Вспомогательная функция для преобразования C-списка импульсов в Go-список
func convertImpulsesToGo(cImpulse *C.Impulse) *engine.Impulse {
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
		Next:    convertImpulsesToGo(cImpulse.Next),
	}

	return goImpulse
}

//export SetRTT
func SetRTT(rtt C.double) {
	singleton.SetRTT(float64(rtt))
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

//export RemoveObjects
func RemoveObjects(ids *C.int32_t, count C.int32_t) {
	goIDs := make([]int, count)
	idSlice := (*[1 << 30]C.int32_t)(unsafe.Pointer(ids))[:count:count]
	for i, id := range idSlice {
		goIDs[i] = int(id)
	}
	singleton.RemoveObjects(goIDs)
}

func main() {}
