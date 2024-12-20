// ignore_for_file: unused_local_variable

import 'dart:async';
import 'dart:ffi' as ffi;
import 'dart:io' as io;

import 'package:ffi/ffi.dart' as ffi;

/// Vector struct (corresponds to Go's Vector)
final class VectorStruct extends ffi.Struct {
  @ffi.Double()
  external double X;

  @ffi.Double()
  external double Y;
}

/// Impulse struct (corresponds to Go's Impulse)
final class ImpulseStruct extends ffi.Struct {
  external VectorStruct Direction;

  @ffi.Double()
  external double Damping;

  external ffi.Pointer<ImpulseStruct> Next;
}

/// Object struct (corresponds to Go's Object)
final class ObjectStruct extends ffi.Struct {
  @ffi.Int32()
  external int ID;

  external VectorStruct Size;
  external VectorStruct Velocity;
  external VectorStruct Position;

  @ffi.Double()
  external double GravityFactor;

  external ffi.Pointer<ImpulseStruct> Impulses;

  @ffi.Uint8()
  external int Particle; // 0 = false, 1 = true
}

/// World struct (corresponds to Go's World)
final class WorldStruct extends ffi.Struct {
  @ffi.Double()
  external double Gravity;

  external VectorStruct Boundary;

  external ffi.Pointer<ObjectStruct> Objects;

  @ffi.Int32()
  external int ObjectCount;
}

// Function typedefs

// GetWorld function
typedef _GetWorldC = ffi.Pointer<WorldStruct> Function();
typedef GetWorldDart = ffi.Pointer<WorldStruct> Function();

// StopEngine function
typedef _StopEngineC = ffi.Void Function();
typedef StopEngineDart = void Function();

// RunEngine function
typedef _RunEngineC = ffi.Void Function(ffi.Double tickMS);
typedef RunEngineDart = void Function(double tickMS);

// CreateWorld function
typedef _CreateWorldC = ffi.Pointer<WorldStruct> Function(
  ffi.Double gravity,
  ffi.Pointer<VectorStruct> boundary,
);
typedef CreateWorldDart = ffi.Pointer<WorldStruct> Function(
  double gravity,
  ffi.Pointer<VectorStruct> boundary,
);

// SetWorld function
typedef _SetWorldC = ffi.Void Function(
  ffi.Pointer<WorldStruct> world,
  ffi.Double rtt,
);
typedef SetWorldDart = void Function(
  ffi.Pointer<WorldStruct> world,
  double rtt,
);

// AddImpulse function
typedef _AddImpulseC = ffi.Void Function(
  ffi.Int32 id,
  ffi.Pointer<VectorStruct> direction,
  ffi.Double damping,
);
typedef AddImpulseDart = void Function(
  int id,
  ffi.Pointer<VectorStruct> direction,
  double damping,
);

// SetVelocity function
typedef _SetVelocityC = ffi.Void Function(
  ffi.Int32 id,
  ffi.Pointer<VectorStruct> velocity,
);
typedef SetVelocityDart = void Function(
  int id,
  ffi.Pointer<VectorStruct> velocity,
);

// SetPosition function
typedef _SetPositionC = ffi.Void Function(
  ffi.Int32 id,
  ffi.Pointer<VectorStruct> position,
);
typedef SetPositionDart = void Function(
  int id,
  ffi.Pointer<VectorStruct> position,
);

// RemoveObjects function
typedef _RemoveObjectsC = ffi.Void Function(
  ffi.Pointer<ffi.Int32> ids,
  ffi.Int32 count,
);
typedef RemoveObjectsDart = void Function(
  ffi.Pointer<ffi.Int32> ids,
  int count,
);

/// Before running this example, make sure to build the library:
/// `make binding_darwin_arm64` or `make binding_windows_amd64`
/// and
/// ```bash
/// cd example
/// dart pub get
/// dart run bin/main.dart
/// ```
void main() {
  final lib = openLib();

  // Load functions
  final getWorld = lib.lookupFunction<_GetWorldC, GetWorldDart>('GetWorld');

  final stopEngine =
      lib.lookupFunction<_StopEngineC, StopEngineDart>('StopEngine');

  final runEngine = lib.lookupFunction<_RunEngineC, RunEngineDart>('RunEngine');

  final createWorld =
      lib.lookupFunction<_CreateWorldC, CreateWorldDart>('CreateWorld');

  final setWorld = lib.lookupFunction<_SetWorldC, SetWorldDart>('SetWorld');

  final addImpulse =
      lib.lookupFunction<_AddImpulseC, AddImpulseDart>('AddImpulse');

  final setVelocity =
      lib.lookupFunction<_SetVelocityC, SetVelocityDart>('SetVelocity');

  final setPosition =
      lib.lookupFunction<_SetPositionC, SetPositionDart>('SetPosition');

  final removeObjects =
      lib.lookupFunction<_RemoveObjectsC, RemoveObjectsDart>('RemoveObjects');

  // Example usage

  {
    // Create world
    final boundary = ffi.calloc<VectorStruct>();
    boundary.ref.X = 6000.0;
    boundary.ref.Y = 480.0;
    final world = createWorld(9.8, boundary);
    ffi.calloc.free(boundary);
    print('World created with gravity 9.8 and boundary 6000x480');

    world.ref.ObjectCount = 5;
    world.ref.Objects = ffi.calloc<ObjectStruct>(5);
    for (var i = 0; i < 5; i++) {
      final object = world.ref.Objects + i;
      object.ref.ID = i + 1;
      object.ref.Size.X = 24.0;
      object.ref.Size.Y = 64.0;
      object.ref.Velocity.X = 0.0;
      object.ref.Velocity.Y = 0.0;
      object.ref.Position.X = 100.0 + i * 32.0;
      object.ref.Position.Y = 100.0;
      object.ref.GravityFactor = 1.0;
      object.ref.Particle = 0;
      object.ref.Impulses = ffi.nullptr;
    }
    setWorld(world, 60); // Set world with RTT 60ms
    print('World objects upserted with initial positions and velocities');
  }

  {
    // Get world info
    final world = getWorld();
    print(
      'Received world with gravity: ${world.ref.Gravity}, '
      'boundary: ${world.ref.Boundary.X}x${world.ref.Boundary.Y}, '
      '${world.ref.ObjectCount} objects',
    );
  }

  {
    // Add impulse
    final direction = ffi.calloc<VectorStruct>();
    direction.ref.X = 10.0;
    direction.ref.Y = 20.0;
    addImpulse(1, direction, 0.95);
    ffi.calloc.free(direction);
    print('Impulse added to object 1 with direction 10x20 and damping 0.95');
  }

  {
    // Set velocity
    final velocity = ffi.calloc<VectorStruct>();
    velocity.ref.X = 5.0;
    velocity.ref.Y = 0.0;
    setVelocity(2, velocity);
    ffi.calloc.free(velocity);
    print('Object 2 velocity set to 5x0');
  }

  {
    // Delete objects
    final ids = ffi.calloc<ffi.Int32>(2);
    ids[0] = 3;
    ids[1] = 4;
    removeObjects(ids, 2);
    print('Objects 3 and 4 removed');
  }

  {
    // Run engine for 16ms (60 FPS)
    runEngine(16.67);
    print('Engine running with 16.67ms (60 FPS) tick interval');
  }

  Timer(const Duration(milliseconds: 200), () {
    {
      // Get world info
      final world = getWorld();
      final buffer = StringBuffer()
        ..writeln('Objects count: ${world.ref.ObjectCount}');
      for (var i = 0; i < world.ref.ObjectCount; i++) {
        final object = (world.ref.Objects + i).ref;
        buffer.writeln(
          'Object #${object.ID}: '
          'position: ${object.Position.X}x${object.Position.Y}, '
          'velocity: ${object.Velocity.X}x${object.Velocity.Y}',
        );
      }
      print(buffer.toString());
    }

    {
      // Stop engine
      stopEngine();
      print('Engine stopped');
    }
  });
}

ffi.DynamicLibrary openLib() {
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
