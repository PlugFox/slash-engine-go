import 'dart:ffi' as ffi;
import 'dart:io' as io;

import 'package:ffi/ffi.dart' as ffi;

/// Vector struct (corresponds to Go's Vector)
final class _VectorStruct extends ffi.Struct {
  @ffi.Double()
  external double X;

  @ffi.Double()
  external double Y;
}

/// Impulse struct (corresponds to Go's Impulse)
final class _ImpulseStruct extends ffi.Struct {
  external _VectorStruct Direction;

  @ffi.Double()
  external double Damping;

  external ffi.Pointer<_ImpulseStruct> Next;
}

/// Object struct (corresponds to Go's Object)
final class _ObjectStruct extends ffi.Struct {
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

// FreeImpulsePtr function
typedef _FreeImpulsePtrC = ffi.Void Function(ffi.Pointer<_ImpulseStruct>);
typedef _FreeImpulsePtrDart = void Function(ffi.Pointer<_ImpulseStruct>);

// FreeObjectPtr function
typedef _FreeObjectPtrC = ffi.Void Function(ffi.Pointer<_ObjectStruct>);
typedef _FreeObjectPtrDart = void Function(ffi.Pointer<_ObjectStruct>);

// FreeWorldPtr function
typedef _FreeWorldPtrC = ffi.Void Function(ffi.Pointer<_WorldStruct>);
typedef _FreeWorldPtrDart = void Function(ffi.Pointer<_WorldStruct>);

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
                'RemoveObjects'),
        _freeImpulsePtrDart =
            lib.lookupFunction<_FreeImpulsePtrC, _FreeImpulsePtrDart>(
                'FreeImpulsePtr'),
        _freeObjectPtrDart =
            lib.lookupFunction<_FreeObjectPtrC, _FreeObjectPtrDart>(
                'FreeObjectPtr'),
        _freeWorldPtrDart = lib
            .lookupFunction<_FreeWorldPtrC, _FreeWorldPtrDart>('FreeWorldPtr');

  final _CreateWorldDart _createWorldDart;
  final _SetWorldDart _setWorldDart;
  final _RunDart _runDart;
  final _GetWorldPtrDart _getWorldPtrDart;
  final _GetObjectPtrDart _getObjectPtrDart;
  final _UpsertObjectDart _upsertObjectDart;
  final _UpsertObjectsDart _upsertObjectsDart;
  final _AddImpulseDart _addImpulseDart;
  final _SetVelocityDart _setVelocityDart;
  final _SetPositionDart _setPositionDart;
  final _SetAnchorDart _setAnchorDart;
  final _RemoveObjectDart _removeObjectDart;
  final _RemoveObjectsDart _removeObjectsDart;
  final _FreeImpulsePtrDart _freeImpulsePtrDart;
  final _FreeObjectPtrDart _freeObjectPtrDart;
  final _FreeWorldPtrDart _freeWorldPtrDart;
  final _StopDart _stopDart;

  /// Create a new world with the given gravity, boundary
  void createWorld({
    required double x,
    required double y,
    required double gravity,
  }) {
    final boundary = ffi.calloc<_VectorStruct>();
    boundary.ref
      ..X = x
      ..Y = y;
    _createWorldDart(gravity, boundary);
    ffi.calloc.free(boundary);
  }

  /// Set the world with the given round-trip time
  void setWorld(ffi.Pointer<_WorldStruct> world, double rtt) {
    _setWorldDart(world, rtt);
  }

  /// Run the engine with the given tick interval
  void run(double tickMS) {
    _runDart(tickMS);
  }

  /// Get the current world
  ffi.Pointer<_WorldStruct> getWorldPtr() {
    return _getWorldPtrDart();
  }

  /// Get the current object
  ffi.Pointer<_ObjectStruct> getObjectPtr(int id) {
    return _getObjectPtrDart(id);
  }

  /// Add or update a single object
  void upsertObject(ffi.Pointer<_ObjectStruct> object) {
    _upsertObjectDart(object);
  }

  /// Add or update multiple objects
  void upsertObjects(ffi.Pointer<_ObjectStruct> objects, int count) {
    _upsertObjectsDart(objects, count);
  }

  /// Add an impulse to an object
  void addImpulse(int id, _VectorStruct direction, double damping) {
    _addImpulseDart(id, direction, damping);
  }

  /// Set velocity for an object
  void setVelocity(int id, _VectorStruct velocity) {
    _setVelocityDart(id, velocity);
  }

  /// Set position for an object
  void setPosition(int id, _VectorStruct position) {
    _setPositionDart(id, position);
  }

  /// Set anchor for an object
  void setAnchor(int id, _VectorStruct anchor) {
    _setAnchorDart(id, anchor);
  }

  /// Remove a single object
  void removeObject(int id) {
    _removeObjectDart(id);
  }

  /// Remove multiple objects
  void removeObjects(ffi.Pointer<ffi.Int32> ids, int count) {
    _removeObjectsDart(ids, count);
  }

  /// Free an impulse pointer
  void freeImpulsePtr(ffi.Pointer<_ImpulseStruct> impulse) {
    _freeImpulsePtrDart(impulse);
  }

  /// Free the given object
  void freeObjectPtr(ffi.Pointer<_ObjectStruct> object) {
    _freeObjectPtrDart(object);
  }

  /// Free the given world
  void freeWorldPtr(ffi.Pointer<_WorldStruct> world) {
    _freeWorldPtrDart(world);
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

  var world = engine.getWorldPtr();

  if (world.address != 0) {
    print('World created: Gravity = ${world.ref.Gravity}, '
        'Boundary = (${world.ref.Boundary.X}, ${world.ref.Boundary.Y}), '
        'Objects = ${world.ref.ObjectCount}');
    // Free the allocated world after use
    engine.freeWorldPtr(world);
    print('World freed');
  }

  // Run the engine
  engine.run(16.0);

  await Future.delayed(const Duration(seconds: 1));

  world = engine.getWorldPtr();
  engine.setWorld(world, 24.0);
  engine.freeWorldPtr(world);

  await Future.delayed(const Duration(seconds: 1));

  // Stop the engine
  engine.stop();
  print('Engine stopped');
}
