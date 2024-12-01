package engine_test

import (
	"testing"
	"time"

	"github.com/plugfox/slash-engine-go/engine"
)

var e = &engine.Engine{}

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

func TestInitWorld(t *testing.T) {
	runWithTimeout(t, func(t *testing.T) {
		e.CreateWorld(9.8, engine.Vector{X: 6000, Y: 480})
		t.Cleanup(e.Stop) // Ensure StopWorld is called after test

		world := e.GetWorld()
		if world == nil {
			t.Fatal("World not initialized")
		}
		if world.Gravity != 9.8 {
			t.Errorf("Expected gravity Y to be 9.8, got %f", world.Gravity)
		}
		if world.Boundary.X != 6000 || world.Boundary.Y != 480 {
			t.Errorf("Expected boundary to be 6000x480, got %fx%f", world.Boundary.X, world.Boundary.Y)
		}
	})
}

func TestWorldToBytes(t *testing.T) {
	runWithTimeout(t, func(t *testing.T) {
		e.CreateWorld(9.8, engine.Vector{X: 6000, Y: 480})
		t.Cleanup(e.Stop) // Ensure StopWorld is called after test

		world := e.GetWorld()
		data := world.ToBytes()
		if len(data) == 0 {
			t.Fatal("Expected non-empty data")
		}
	})
}
