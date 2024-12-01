package engine

import (
	flatbuffers "github.com/google/flatbuffers/go"
	"github.com/plugfox/slash-engine-go/generated/Game"
)

// Конвертация Vector в FlatBuffers
func serializeVector(builder *flatbuffers.Builder, vec Vector) flatbuffers.UOffsetT {
	return Game.CreateVector(builder, vec.X, vec.Y)
}

// Конвертация Impulse в FlatBuffers
func serializeImpulse(builder *flatbuffers.Builder, impulse *Impulse) flatbuffers.UOffsetT {
	if impulse == nil {
		return 0
	}

	next := serializeImpulse(builder, impulse.Next) // Рекурсивная обработка
	dir := serializeVector(builder, impulse.Direction)

	Game.ImpulseStart(builder)
	Game.ImpulseAddDirection(builder, dir)
	Game.ImpulseAddDamping(builder, impulse.Damping)
	Game.ImpulseAddNext(builder, next)
	return Game.ImpulseEnd(builder)
}

// Конвертация Object в FlatBuffers
func serializeObject(builder *flatbuffers.Builder, obj *Object) flatbuffers.UOffsetT {
	size := serializeVector(builder, obj.Size)
	velocity := serializeVector(builder, obj.Velocity)
	position := serializeVector(builder, obj.Position)
	anchor := serializeVector(builder, obj.Anchor)
	impulses := serializeImpulse(builder, obj.Impulses)

	Game.ObjectStart(builder)
	Game.ObjectAddID(builder, int32(obj.ID))
	Game.ObjectAddType(builder, Game.ObjectType(obj.Type))
	Game.ObjectAddClient(builder, obj.Client)
	Game.ObjectAddSize(builder, size)
	Game.ObjectAddVelocity(builder, velocity)
	Game.ObjectAddPosition(builder, position)
	Game.ObjectAddAnchor(builder, anchor)
	Game.ObjectAddGravityFactor(builder, obj.GravityFactor)
	Game.ObjectAddImpulses(builder, impulses)
	return Game.ObjectEnd(builder)
}

// Конвертация World в FlatBuffers
func serializeWorldToBytes(world *World) []byte {
	builder := flatbuffers.NewBuilder(1024)

	// Преобразуем объекты
	objects := make([]flatbuffers.UOffsetT, 0, len(world.Objects))
	for _, obj := range world.Objects {
		objects = append(objects, serializeObject(builder, obj))
	}
	Game.WorldStartObjectsVector(builder, len(objects))
	for i := len(objects) - 1; i >= 0; i-- {
		builder.PrependUOffsetT(objects[i])
	}
	objectsVector := builder.EndVector(len(objects))

	// Преобразуем гравитацию и границы
	boundary := serializeVector(builder, world.Boundary)

	// Создаём мир
	Game.WorldStart(builder)
	Game.WorldAddGravity(builder, world.Gravity)
	Game.WorldAddBoundary(builder, boundary)
	Game.WorldAddObjects(builder, objectsVector)
	worldOffset := Game.WorldEnd(builder)

	builder.Finish(worldOffset)
	return builder.FinishedBytes()
}

// Декодируем Vector из FlatBuffers
func deserializeVector(vec *Game.Vector) Vector {
	return Vector{
		X: vec.X(),
		Y: vec.Y(),
	}
}

// Декодируем Impulse из FlatBuffers
func deserializeImpulse(impulse *Game.Impulse) *Impulse {
	if impulse == nil {
		return nil
	}

	return &Impulse{
		Direction: deserializeVector(impulse.Direction(nil)),
		Damping:   impulse.Damping(),
		Next:      deserializeImpulse(impulse.Next(nil)),
	}
}

// Декодируем Object из FlatBuffers
func deserializeObject(obj *Game.Object) *Object {
	return &Object{
		ID:            int(obj.ID()),
		Type:          ObjectType(obj.Type()),
		Client:        obj.Client(),
		Size:          deserializeVector(obj.Size(nil)),
		Velocity:      deserializeVector(obj.Velocity(nil)),
		Position:      deserializeVector(obj.Position(nil)),
		Anchor:        deserializeVector(obj.Anchor(nil)),
		GravityFactor: obj.GravityFactor(),
		Impulses:      deserializeImpulse(obj.Impulses(nil)),
	}
}

// Декодируем World из FlatBuffers
func deserializeWorldFromBytes(data []byte) *World {
	if data == nil || len(data) == 0 {
		return nil
	}

	world := Game.GetRootAsWorld(data, 0)

	// Декодируем объекты
	objects := make(map[int]*Object, world.ObjectsLength())
	for i := 0; i < world.ObjectsLength(); i++ {
		var obj Game.Object
		if world.Objects(&obj, i) {
			goObj := deserializeObject(&obj)
			objects[goObj.ID] = goObj
		}
	}

	return &World{
		Gravity:  world.Gravity(),
		Boundary: deserializeVector(world.Boundary(nil)),
		Objects:  objects,
	}
}
