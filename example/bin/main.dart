import 'dart:async';
import 'dart:ffi' as ffi;
import 'dart:io' as io;

import 'package:ffi/ffi.dart' as ffi;

/// Struct for Object (corresponds to Go's C.Object)
final class ObjectStruct extends ffi.Struct {
  @ffi.Int32()
  external int id;

  @ffi.Double()
  external double posX;

  @ffi.Double()
  external double posY;

  @ffi.Double()
  external double velX;

  @ffi.Double()
  external double velY;

  @ffi.Double()
  external double mass;
}

// Function typedefs

typedef _InitWorldWithOptionsC = ffi.Void Function(
  ffi.Double gravityX,
  ffi.Double gravityY,
  ffi.Double boundaryX,
  ffi.Double boundaryY,
  ffi.Double tickMS,
  ffi.Double rtt,
  ffi.Uint8 autoStart,
);

typedef InitWorldWithOptionsDart = void Function(
  double gravityX,
  double gravityY,
  double boundaryX,
  double boundaryY,
  double tickMS,
  double rtt,
  int autoStart,
);

typedef _UpsertObjectsC = ffi.Void Function(
  ffi.Pointer<ObjectStruct> objects,
  ffi.Int32 count,
);

typedef UpsertObjectsDart = void Function(
  ffi.Pointer<ObjectStruct> objects,
  int count,
);

typedef _GetObjectPositionsC = ffi.Pointer<ObjectStruct> Function(
  ffi.Pointer<ffi.Int32>,
);

typedef GetObjectPositionsDart = ffi.Pointer<ObjectStruct> Function(
  ffi.Pointer<ffi.Int32>,
);

typedef _DeleteObjectsC = ffi.Void Function(
  ffi.Pointer<ffi.Int32> ids,
  ffi.Int32 count,
);

typedef DeleteObjectsDart = void Function(
  ffi.Pointer<ffi.Int32> ids,
  int count,
);

typedef _ApplyImpulseC = ffi.Void Function(
  ffi.Int32 objectID,
  ffi.Double impulseX,
  ffi.Double impulseY,
);

typedef ApplyImpulseDart = void Function(
  int objectID,
  double impulseX,
  double impulseY,
);

typedef _StopWorldC = ffi.Void Function();

typedef StopWorldDart = void Function();

/// Before running this example, make sure to build the library:
/// `make binding_darwin_arm64`
void main() {
  final lib = openLib();

  // Load functions
  final initWorld =
      lib.lookupFunction<_InitWorldWithOptionsC, InitWorldWithOptionsDart>(
    'InitWorld',
  );

  final upsertObjects = lib.lookupFunction<_UpsertObjectsC, UpsertObjectsDart>(
    'UpsertObjects',
  );

  final getObjectPositions =
      lib.lookupFunction<_GetObjectPositionsC, GetObjectPositionsDart>(
    'GetObjectPositions',
  );

  final deleteObjects = lib.lookupFunction<_DeleteObjectsC, DeleteObjectsDart>(
    'DeleteObjects',
  );

  final applyImpulse = lib.lookupFunction<_ApplyImpulseC, ApplyImpulseDart>(
    'ApplyImpulse',
  );

  final stopWorld = lib.lookupFunction<_StopWorldC, StopWorldDart>(
    'StopWorld',
  );

  // Example usage

  // Initialize world
  initWorld(
    0, // gravityX
    -9.8, // gravityY
    6000, // boundaryX
    700, // boundaryY
    16.67, // tickMS
    0.1, // rtt
    1, // autoStart
  );

  // Create objects
  final objectList = [
    ffi.Struct.create<ObjectStruct>()
      ..id = 1
      ..posX = 10.0
      ..posY = 50.0
      ..velX = 0.0
      ..velY = 0.0
      ..mass = 1.0,
    for (var i = 2; i <= 1000; i++)
      ffi.Struct.create<ObjectStruct>()
        ..id = i
        ..posX = 20.0
        ..posY = 60.0
        ..velX = 1.0
        ..velY = 1.0
        ..mass = 2.0,
  ];
  final objectPointer = ffi.calloc<ObjectStruct>(objectList.length);
  for (var i = 0; i < objectList.length; i++) {
    objectPointer[i] = objectList[i];
  }
  upsertObjects(objectPointer, objectList.length);
  ffi.malloc.free(objectPointer);

  // Apply impulse to an object
  applyImpulse(1, 5.0, 18.0);
  Timer(Duration(milliseconds: 40), () {
    final countPtr = ffi.calloc<ffi.Int32>();
    final objects = getObjectPositions(countPtr);
    // Get count:
    final count = countPtr.value;
    ffi.calloc.free(countPtr); // Освобождаем память
    print('Object count: $count');
    final map = <int, ObjectStruct>{};
    for (var i = 0; i < count; i++) {
      final object = (objects + i).ref;
      map[object.id] = object;
    }
    final object = map[1]!;
    assert(object.posX >= 0 && object.posY >= 0 && object.id == 1);
    print('Object #${object.id}: (${object.posX}, ${object.posY})');
  });

  // Get object positions
  Timer(Duration(seconds: 2), () {
    for (var i = 0; i < 10; i++) {
      final countPtr = ffi.calloc<ffi.Int32>();
      final objects = getObjectPositions(countPtr);
      final count = countPtr.value;
      ffi.calloc.free(countPtr); // Освобождаем память
      for (var j = 0; j < count; j++) {
        final object = (objects + j).ref;
        //print('Object #${object.id}: (${object.posX}, ${object.posY})');
        assert(object.posX >= 0 && object.posY >= 0);
      }
      ffi.malloc.free(objects);
    }
    final stopWatch = Stopwatch()..start();
    final countPtr = ffi.calloc<ffi.Int32>();
    final objects = getObjectPositions(countPtr);
    final count = countPtr.value;
    ffi.calloc.free(countPtr);
    assert(objects.ref.posX >= 0 && objects.ref.posY >= 0 && count > 0);
    ffi.malloc.free(objects);
    print('Time per call: ${stopWatch.elapsedMicroseconds / 1000} ms');
  });

  // Delete objects
  Timer(Duration(seconds: 3), () {
    final ids = ffi.calloc<ffi.Int32>(2);
    ids[0] = 1;
    ids[1] = 2;
    deleteObjects(ids, 2);
    ffi.malloc.free(ids);
  });

  Timer(Duration(seconds: 5), () {
    stopWorld();
    io.sleep(Duration(milliseconds: 250));
    io.exit(0);
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
