/// Dart representation of Vector
class Vector {
  final double x;
  final double y;

  Vector(this.x, this.y);

  @override
  String toString() => 'Vector(x: $x, y: $y)';
}

/// Dart representation of Impulse
class Impulse {
  final Vector direction;
  final double damping;
  final Impulse? next;

  Impulse(this.direction, this.damping, this.next);

  @override
  String toString() =>
      'Impulse(direction: $direction, damping: $damping, next: $next)';
}

/// Dart representation of Object
class GameObject {
  final int id;
  final int type;
  final bool client;
  final Vector size;
  final Vector velocity;
  final Vector position;
  final Vector anchor;
  final double gravityFactor;
  final Impulse? impulses;

  GameObject({
    required this.id,
    required this.type,
    required this.client,
    required this.size,
    required this.velocity,
    required this.position,
    required this.anchor,
    required this.gravityFactor,
    this.impulses,
  });

  @override
  String toString() {
    return 'GameObject(id: $id, type: $type, client: $client, '
        'size: $size, velocity: $velocity, position: $position, '
        'anchor: $anchor, gravityFactor: $gravityFactor, impulses: $impulses)';
  }
}

/// Dart representation of World
class GameWorld {
  final double gravity;
  final Vector boundary;
  final List<GameObject> objects;

  GameWorld({
    required this.gravity,
    required this.boundary,
    required this.objects,
  });

  @override
  String toString() {
    return 'GameWorld(gravity: $gravity, boundary: $boundary, objects: $objects)';
  }
}
