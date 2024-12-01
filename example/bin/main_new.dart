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

// GetWorld function
typedef _GetWorldC = ffi.Pointer<_WorldStruct> Function();
typedef _GetWorldDart = ffi.Pointer<_WorldStruct> Function();

// FreeWorld function
typedef _FreeWorldC = ffi.Void Function(
  ffi.Pointer<_WorldStruct>,
);
typedef _FreeWorldDart = void Function(
  ffi.Pointer<_WorldStruct>,
);

// Stop function
typedef _StopC = ffi.Void Function();
typedef _StopDart = void Function();

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
      : _getWorldC = lib.lookupFunction<_GetWorldC, _GetWorldDart>('GetWorld'),
        _freeWorldC =
            lib.lookupFunction<_FreeWorldC, _FreeWorldDart>('FreeWorld'),
        _createWorldC =
            lib.lookupFunction<_CreateWorldC, _CreateWorldDart>('CreateWorld'),
        _stopC = lib.lookupFunction<_StopC, _StopDart>('Stop');

  final _CreateWorldDart _createWorldC;
  final _GetWorldDart _getWorldC;
  final _FreeWorldDart _freeWorldC;
  final _StopDart _stopC;

  /// Create a new world with the given gravity, boundary
  void createWorld(
      {required double x, required double y, required double gravity}) {
    final boundary = ffi.calloc<_VectorStruct>();
    boundary.ref
      ..X = x
      ..Y = y;
    _createWorldC(gravity, boundary);
    ffi.calloc.free(boundary);
  }

  /// Get the current world
  // ignore: unused_element
  ffi.Pointer<_WorldStruct> _getWorld() {
    final result = _getWorldC();
    // _freeWorld(result);
    return result;
  }

  /// Free the given world
  // ignore: unused_element
  void _freeWorld(ffi.Pointer<_WorldStruct> world) {
    _freeWorldC(world);
  }

  /// Stop the engine
  void stop() {
    _stopC();
  }
}

void main() {
  final engine = SlashEngine();

  engine.createWorld(
    gravity: 9.81,
    x: 1000.0,
    y: 500.0,
  );

  final world = engine._getWorld();

  if (world.address != 0) {
    print('World created: Gravity = ${world.ref.Gravity}, '
        'Boundary = (${world.ref.Boundary.X}, ${world.ref.Boundary.Y}), '
        'Objects = ${world.ref.ObjectCount}');
    // Free the allocated world after use
    engine._freeWorld(world);
    print('World freed');
  }

  // Stop the engine
  engine.stop();
  print('Engine stopped');
}
