import 'dart:ffi' as ffi;
import 'dart:io' as io;
import 'dart:typed_data';

import 'package:ffi/ffi.dart' as ffi;

import 'models.dart';

/// Vector struct (corresponds to Go's Vector)
final class _VectorStruct extends ffi.Struct {
  /// Converts a _VectorStruct to a Dart Vector
  static Vector convert(_VectorStruct vector) => Vector(vector.X, vector.Y);

  @ffi.Double()
  external double X;

  @ffi.Double()
  external double Y;
}

/// Impulse struct (corresponds to Go's Impulse)
final class _ImpulseStruct extends ffi.Struct {
  /// Converts a _ImpulseStruct to a Dart Impulse
  static Impulse convert(_ImpulseStruct imp) {
    final nextPtr = imp.Next;
    Impulse? next;
    if (nextPtr.address != 0) {
      next = _ImpulseStruct.convert(nextPtr.ref);
    }
    return Impulse(
      _VectorStruct.convert(imp.Direction),
      imp.Damping,
      next,
    );
  }

  external _VectorStruct Direction;

  @ffi.Double()
  external double Damping;

  external ffi.Pointer<_ImpulseStruct> Next;
}

/// Object struct (corresponds to Go's Object)
final class _ObjectStruct extends ffi.Struct {
  /// Converts a _ObjectStruct to a Dart GameObject
  static GameObject convert(_ObjectStruct obj) => GameObject(
        id: obj.ID,
        type: obj.Type,
        client: obj.Client != 0,
        size: _VectorStruct.convert(obj.Size),
        velocity: _VectorStruct.convert(obj.Velocity),
        position: _VectorStruct.convert(obj.Position),
        anchor: _VectorStruct.convert(obj.Anchor),
        gravityFactor: obj.GravityFactor,
        impulses: _ImpulseStruct.convert(obj.Impulses.ref),
      );

  @ffi.Int32()
  external int ID;

  @ffi.Int32()
  external int Type;

  @ffi.Uint8()
  external int Client;

  external _VectorStruct Size;
  external _VectorStruct Velocity;
  external _VectorStruct Position;
  external _VectorStruct Anchor;

  @ffi.Double()
  external double GravityFactor;

  external ffi.Pointer<_ImpulseStruct> Impulses;
}

/// World struct (corresponds to Go's World)
final class _WorldStruct extends ffi.Struct {
  /// Converts a _WorldStruct to a Dart GameWorld
  static GameWorld convert(_WorldStruct world) => GameWorld(
        gravity: world.Gravity,
        boundary: _VectorStruct.convert(world.Boundary),
        objects: world.ObjectCount < 1
            ? const <GameObject>[]
            : Iterable<GameObject?>.generate(
                world.ObjectCount,
                (i) {
                  final ptr = world.Objects + i;
                  if (ptr.address == 0) return null;
                  return _ObjectStruct.convert(ptr.ref);
                },
              ).whereType<GameObject>().toList(growable: false),
      );

  @ffi.Double()
  external double Gravity;

  external _VectorStruct Boundary;

  external ffi.Pointer<_ObjectStruct> Objects;

  @ffi.Int32()
  external int ObjectCount;
}

// CreateWorld function
typedef _CreateWorldC = ffi.Void Function(
  ffi.Double gravity,
  ffi.Pointer<_VectorStruct> boundary,
);
typedef _CreateWorldDart = void Function(
  double gravity,
  ffi.Pointer<_VectorStruct> boundary,
);

// SetWorld function
typedef _SetWorldC = ffi.Void Function(
  ffi.Pointer<_WorldStruct> world,
  ffi.Double rtt,
);
typedef _SetWorldDart = void Function(
  ffi.Pointer<_WorldStruct> world,
  double rtt,
);

// Run function
typedef _RunC = ffi.Void Function(ffi.Double tickMS);
typedef _RunDart = void Function(double tickMS);

// Stop function
typedef _StopC = ffi.Void Function();
typedef _StopDart = void Function();

// GetWorldPtr function
typedef _GetWorldPtrC = ffi.Pointer<_WorldStruct> Function();
typedef _GetWorldPtrDart = ffi.Pointer<_WorldStruct> Function();

// Get world bytes function
typedef _GetWorldBytesC = ffi.Pointer<ffi.Uint8> Function(
    ffi.Pointer<ffi.Int32>);
typedef _GetWorldBytesDart = ffi.Pointer<ffi.Uint8> Function(
    ffi.Pointer<ffi.Int32>);

// _GetObjectPtr function
typedef _GetObjectPtrC = ffi.Pointer<_ObjectStruct> Function(ffi.Int32 id);
typedef _GetObjectPtrDart = ffi.Pointer<_ObjectStruct> Function(int id);

// UpsertObject function
typedef _UpsertObjectC = ffi.Void Function(ffi.Pointer<_ObjectStruct>);
typedef _UpsertObjectDart = void Function(ffi.Pointer<_ObjectStruct>);

// UpsertObjects function
typedef _UpsertObjectsC = ffi.Void Function(
    ffi.Pointer<_ObjectStruct>, ffi.Int32);
typedef _UpsertObjectsDart = void Function(ffi.Pointer<_ObjectStruct>, int);

// AddImpulse function
typedef _AddImpulseC = ffi.Void Function(ffi.Int32, _VectorStruct, ffi.Double);
typedef _AddImpulseDart = void Function(int, _VectorStruct, double);

// SetVelocity function
typedef _SetVelocityC = ffi.Void Function(ffi.Int32, _VectorStruct);
typedef _SetVelocityDart = void Function(int, _VectorStruct);

// SetPosition function
typedef _SetPositionC = ffi.Void Function(ffi.Int32, _VectorStruct);
typedef _SetPositionDart = void Function(int, _VectorStruct);

// SetAnchor function
typedef _SetAnchorC = ffi.Void Function(ffi.Int32, _VectorStruct);
typedef _SetAnchorDart = void Function(int, _VectorStruct);

// RemoveObject function
typedef _RemoveObjectC = ffi.Void Function(ffi.Int32);
typedef _RemoveObjectDart = void Function(int);

// RemoveObjects function
typedef _RemoveObjectsC = ffi.Void Function(ffi.Pointer<ffi.Int32>, ffi.Int32);
typedef _RemoveObjectsDart = void Function(ffi.Pointer<ffi.Int32>, int);

// Utility function to open the shared library
ffi.DynamicLibrary _openEngineLib() {
  if (io.Platform.isMacOS) {
    const lib = 'output/binding/darwin/arm64/slashengine.dylib';
    final path = '${io.Directory.current.parent.path}/$lib';
    assert(io.File(path).existsSync(), 'File not found: $path');
    return ffi.DynamicLibrary.open(path);
  } else if (io.Platform.isLinux) {
    const lib = 'output/binding/linux/aarch64/slashengine.so';
    final path = '${io.Directory.current.parent.path}/$lib';
    assert(io.File(path).existsSync(), 'File not found: $path');
    return ffi.DynamicLibrary.open(path);
  } else if (io.Platform.isWindows) {
    const lib = 'output/binding/windows/amd64/slashengine.dll';
    final path = '${io.Directory.current.parent.path}/$lib';
    assert(io.File(path).existsSync(), 'File not found: $path');
    return ffi.DynamicLibrary.open(path);
  }

  throw UnsupportedError('Unknown platform: ${io.Platform.operatingSystem}');
}

class SlashEngine {
  factory SlashEngine() => _instance ??= SlashEngine._(_openEngineLib());
  static SlashEngine? _instance;
  SlashEngine._(ffi.DynamicLibrary lib)
      : _createWorldDart =
            lib.lookupFunction<_CreateWorldC, _CreateWorldDart>('CreateWorld'),
        _setWorldDart =
            lib.lookupFunction<_SetWorldC, _SetWorldDart>('SetWorld'),
        _runDart = lib.lookupFunction<_RunC, _RunDart>('Run'),
        _stopDart = lib.lookupFunction<_StopC, _StopDart>('Stop'),
        _getWorldPtrDart =
            lib.lookupFunction<_GetWorldPtrC, _GetWorldPtrDart>('GetWorldPtr'),
        _getWorldBytesDart =
            lib.lookupFunction<_GetWorldBytesC, _GetWorldBytesDart>(
                'GetWorldBytes'),
        _getObjectPtrDart = lib
            .lookupFunction<_GetObjectPtrC, _GetObjectPtrDart>('GetObjectPtr'),
        _upsertObjectDart = lib
            .lookupFunction<_UpsertObjectC, _UpsertObjectDart>('UpsertObject'),
        _upsertObjectsDart =
            lib.lookupFunction<_UpsertObjectsC, _UpsertObjectsDart>(
                'UpsertObjects'),
        _addImpulseDart =
            lib.lookupFunction<_AddImpulseC, _AddImpulseDart>('AddImpulse'),
        _setVelocityDart =
            lib.lookupFunction<_SetVelocityC, _SetVelocityDart>('SetVelocity'),
        _setPositionDart =
            lib.lookupFunction<_SetPositionC, _SetPositionDart>('SetPosition'),
        _setAnchorDart =
            lib.lookupFunction<_SetAnchorC, _SetAnchorDart>('SetAnchor'),
        _removeObjectDart = lib
            .lookupFunction<_RemoveObjectC, _RemoveObjectDart>('RemoveObject'),
        _removeObjectsDart =
            lib.lookupFunction<_RemoveObjectsC, _RemoveObjectsDart>(
                'RemoveObjects');

  final _CreateWorldDart _createWorldDart;
  final _SetWorldDart _setWorldDart;
  final _RunDart _runDart;
  final _GetWorldPtrDart _getWorldPtrDart;
  final _GetWorldBytesDart _getWorldBytesDart;
  final _GetObjectPtrDart _getObjectPtrDart;
  final _UpsertObjectDart _upsertObjectDart;
  final _UpsertObjectsDart _upsertObjectsDart;
  final _AddImpulseDart _addImpulseDart;
  final _SetVelocityDart _setVelocityDart;
  final _SetPositionDart _setPositionDart;
  final _SetAnchorDart _setAnchorDart;
  final _RemoveObjectDart _removeObjectDart;
  final _RemoveObjectsDart _removeObjectsDart;
  final _StopDart _stopDart;

  /// Create a new world with the given gravity, boundary
  void createWorld({
    required double x,
    required double y,
    required double gravity,
  }) {
    final boundary = ffi.calloc<_VectorStruct>();
    try {
      boundary.ref
        ..X = x
        ..Y = y;
      _createWorldDart(gravity, boundary);
    } finally {
      ffi.calloc.free(boundary);
    }
  }

  /// Set the world with the given round-trip time
  void setWorld(ffi.Pointer<_WorldStruct> world, double rtt) {
    _setWorldDart(world, rtt);
  }

  /// Run the engine with the given tick interval
  void run(double tickMS) {
    _runDart(tickMS);
  }

  /// Get the current world by reference
  GameWorld? getWorldPtr() {
    final ptr = _getWorldPtrDart();
    if (ptr.address == 0) return null;
    return _WorldStruct.convert(ptr.ref);
  }

  /// Get the current world
  Uint8List? getWorldBytes() {
    // Запрашиваем размер данных
    final sizePtr = ffi.calloc<ffi.Int32>();
    final dataPtr = _getWorldBytesDart(sizePtr);

    final size = sizePtr.value;
    ffi.calloc.free(sizePtr);

    // Если размер равен 0, возвращаем null
    if (size <= 0 || dataPtr.address == 0) {
      return null;
    }

    // Копируем данные в List<int>
    final bytes = dataPtr.asTypedList(size);

    // Освобождаем выделенную память
    ffi.calloc.free(dataPtr);

    return bytes;
  }

  /// Get the current object
  GameObject? getObjectPtr(int id) {
    final ptr = _getObjectPtrDart(id);
    if (ptr.address == 0) return null;
    return _ObjectStruct.convert(ptr.ref);
  }

  /// Add or update a single object
  void upsertObjectPtr(GameObject object) {
    final ptr = ffi.calloc<_ObjectStruct>();
    try {
      ptr.ref
        ..ID = object.id
        ..Type = object.type
        ..Client = object.client ? 1 : 0
        ..Size.X = object.size.x
        ..Size.Y = object.size.y
        ..Velocity.X = object.velocity.x
        ..Velocity.Y = object.velocity.y
        ..Position.X = object.position.x
        ..Position.Y = object.position.y
        ..Anchor.X = object.anchor.x
        ..Anchor.Y = object.anchor.y
        ..GravityFactor = object.gravityFactor;
      _upsertObjectDart(ptr);
    } finally {
      ffi.calloc.free(ptr);
    }
  }

  /// Add or update multiple objects
  void upsertObjectsPtr(List<GameObject> objects) {
    final count = objects.length;
    if (count == 0) return;
    final ptr = ffi.calloc<_ObjectStruct>(count);
    try {
      for (var i = 0; i < count; i++) {
        final object = objects[i];
        final ptrObject = ptr + i;
        ptrObject.ref
          ..ID = object.id
          ..Type = object.type
          ..Client = object.client ? 1 : 0
          ..Size.X = object.size.x
          ..Size.Y = object.size.y
          ..Velocity.X = object.velocity.x
          ..Velocity.Y = object.velocity.y
          ..Position.X = object.position.x
          ..Position.Y = object.position.y
          ..Anchor.X = object.anchor.x
          ..Anchor.Y = object.anchor.y
          ..GravityFactor = object.gravityFactor;
      }
      _upsertObjectsDart(ptr, count);
    } finally {
      ffi.calloc.free(ptr);
    }
  }

  /// Add an impulse to an object
  void addImpulse(int id, Vector direction, double damping) {
    final vector = ffi.calloc<_VectorStruct>();
    try {
      vector.ref
        ..X = direction.x
        ..Y = direction.y;
      _addImpulseDart(id, vector.ref, damping);
    } finally {
      ffi.calloc.free(vector);
    }
  }

  /// Set velocity for an object
  void setVelocity(int id, Vector velocity) {
    final vector = ffi.calloc<_VectorStruct>();
    try {
      vector.ref
        ..X = velocity.x
        ..Y = velocity.y;
      _setVelocityDart(id, vector.ref);
    } finally {
      ffi.calloc.free(vector);
    }
  }

  /// Set position for an object
  void setPosition(int id, Vector position) {
    final vector = ffi.calloc<_VectorStruct>();
    try {
      vector.ref
        ..X = position.x
        ..Y = position.y;
      _setPositionDart(id, vector.ref);
    } finally {
      ffi.calloc.free(vector);
    }
  }

  /// Set anchor for an object
  void setAnchor(int id, Vector anchor) {
    final vector = ffi.calloc<_VectorStruct>();
    try {
      vector.ref
        ..X = anchor.x
        ..Y = anchor.y;
      _setAnchorDart(id, vector.ref);
    } finally {
      ffi.calloc.free(vector);
    }
  }

  /// Remove a single object
  void removeObject(int id) {
    _removeObjectDart(id);
  }

  /// Remove multiple objects
  void removeObjects(Iterable<int> ids) {
    final count = ids.length;
    final ptr = ffi.calloc<ffi.Int32>(count);
    try {
      var i = 0;
      for (final id in ids) {
        ptr[i] = id;
        i++;
      }
      _removeObjectsDart(ptr, count);
    } finally {
      ffi.calloc.free(ptr);
    }
  }

  /// Stop the engine
  void stop() {
    _stopDart();
  }
}

void main() async {
  final engine = SlashEngine();

  engine.createWorld(
    gravity: 9.81,
    x: 1000.0,
    y: 500.0,
  );

  engine.upsertObjectsPtr([
    GameObject(
      id: 1,
      type: 1,
      client: true,
      size: Vector(24.0, 64.0),
      velocity: Vector(0.0, 0.0),
      position: Vector(100.0, 100.0),
      anchor: Vector(0.0, 0.0),
      gravityFactor: 1.0,
    ),
    GameObject(
      id: 2,
      type: 1,
      client: true,
      size: Vector(24.0, 64.0),
      velocity: Vector(0.0, 0.0),
      position: Vector(132.0, 100.0),
      anchor: Vector(0.0, 0.0),
      gravityFactor: 1.0,
    ),
    GameObject(
      id: 3,
      type: 1,
      client: true,
      size: Vector(24.0, 64.0),
      velocity: Vector(0.0, 0.0),
      position: Vector(164.0, 100.0),
      anchor: Vector(0.0, 0.0),
      gravityFactor: 1.0,
    ),
    GameObject(
      id: 4,
      type: 1,
      client: true,
      size: Vector(24.0, 64.0),
      velocity: Vector(0.0, 0.0),
      position: Vector(196.0, 100.0),
      anchor: Vector(0.0, 0.0),
      gravityFactor: 1.0,
    ),
  ]);

  var bytes = engine.getWorldBytes();

  // Run the engine
  engine.run(16.0);

  await Future.delayed(const Duration(seconds: 1));

  bytes = engine.getWorldBytes();

  await Future.delayed(const Duration(seconds: 1));

  // Stop the engine
  engine.stop();
  print('Engine stopped');
}
