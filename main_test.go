package main

import (
	"testing"
	"time"
)

// runWithTimeout runs a test function with a timeout.
func runWithTimeout(t *testing.T, testFunc func(t *testing.T)) {
	done := make(chan struct{})
	go func() {
		defer close(done)
		testFunc(t)
	}()

	select {
	case <-done:
		// Test completed within time limit
	case <-time.After(1 * time.Second):
		t.Fatal("Test timed out")
	}
}

func TestInitWorldWithOptions(t *testing.T) {
	runWithTimeout(t, func(t *testing.T) {
		manager.InitWorldWithOptions(0, -9.8, 100, 100, 16.67, 0.1, false)
		t.Cleanup(StopWorld) // Ensure StopWorld is called after test

		world := manager.GetWorld()
		if world == nil {
			t.Fatal("World not initialized")
		}
		if world.Gravity.Y != -9.8 {
			t.Errorf("Expected gravity Y to be -9.8, got %f", world.Gravity.Y)
		}
		if world.Boundary.X != 100 || world.Boundary.Y != 100 {
			t.Errorf("Expected boundary to be 100x100, got %fx%f", world.Boundary.X, world.Boundary.Y)
		}
	})
}

func TestUpsertObjectsWithoutCGO(t *testing.T) {
	runWithTimeout(t, func(t *testing.T) {
		manager.InitWorldWithOptions(0, -9.8, 100, 100, 16.67, 0.1, false)
		t.Cleanup(StopWorld)

		world := manager.GetWorld()

		objects := []*Object{
			{ID: 1, Position: Vector{X: 10, Y: 50}, Velocity: Vector{X: 2, Y: 3}, Mass: 1.0},
			{ID: 2, Position: Vector{X: 20, Y: 60}, Velocity: Vector{X: -1, Y: -2}, Mass: 2.0},
		}
		for _, obj := range objects {
			world.mutex.Lock()
			world.Objects[obj.ID] = obj
			world.mutex.Unlock()
		}

		if len(world.Objects) != 2 {
			t.Fatalf("Expected 2 objects, got %d", len(world.Objects))
		}
		if obj, ok := world.Objects[1]; !ok || obj.Position.X != 10 || obj.Position.Y != 50 {
			t.Errorf("Object 1 position mismatch, got %fx%f", obj.Position.X, obj.Position.Y)
		}
		if obj, ok := world.Objects[2]; !ok || obj.Mass != 2.0 {
			t.Errorf("Object 2 mass mismatch, got %f", obj.Mass)
		}
	})
}

func TestExtrapolateObjectWithoutCGO(t *testing.T) {
	runWithTimeout(t, func(t *testing.T) {
		manager.InitWorldWithOptions(0, 0, 100, 100, 16.67, 0.1, false)
		t.Cleanup(StopWorld)

		world := manager.GetWorld()

		obj := &Object{
			ID:         1,
			Position:   Vector{X: 10, Y: 50},
			Velocity:   Vector{X: 2, Y: 0}, // Moving horizontally
			Mass:       1.0,
			LastUpdate: time.Now(),
		}

		world.mutex.Lock()
		world.Objects[obj.ID] = obj
		world.mutex.Unlock()

		// Wait for some time to simulate elapsed time
		time.Sleep(100 * time.Millisecond)

		// Capture elapsed time and extrapolate
		world.mutex.Lock()
		before := world.Objects[1].Position.X
		elapsed := time.Since(world.Objects[1].LastUpdate).Seconds()
		world.extrapolateObject(world.Objects[1])
		after := world.Objects[1].Position.X
		world.mutex.Unlock()

		// Calculate expected position with RTT
		rtt := 0.1 // RTT from InitWorldWithOptions
		totalElapsed := elapsed + rtt
		expectedX := 10 + 2*totalElapsed

		// Debugging information
		t.Logf("Before: %f, After: %f, Elapsed: %f, RTT: %f, Total Elapsed: %f, Expected: %f", before, after, elapsed, rtt, totalElapsed, expectedX)

		if after < expectedX-0.01 || after > expectedX+0.01 {
			t.Errorf("Object 1 extrapolated position mismatch, got %f, expected %f", after, expectedX)
		}
	})
}

func TestApplyImpulseWithoutCGO(t *testing.T) {
	runWithTimeout(t, func(t *testing.T) {
		manager.InitWorldWithOptions(0, 0, 100, 100, 16.67, 0.1, false)
		t.Cleanup(StopWorld)

		world := manager.GetWorld()

		obj := &Object{ID: 1, Position: Vector{X: 10, Y: 50}, Velocity: Vector{X: 0, Y: 0}, Mass: 1.0}
		world.mutex.Lock()
		world.Objects[obj.ID] = obj
		world.mutex.Unlock()

		world.mutex.Lock()
		if obj, ok := world.Objects[1]; ok {
			obj.Velocity.X += 5 / obj.Mass
			obj.Velocity.Y += -3 / obj.Mass
		}
		world.mutex.Unlock()

		if obj := world.Objects[1]; obj.Velocity.X != 5 || obj.Velocity.Y != -3 {
			t.Errorf("Object 1 velocity mismatch after impulse, got %fx%f", obj.Velocity.X, obj.Velocity.Y)
		}
	})
}
