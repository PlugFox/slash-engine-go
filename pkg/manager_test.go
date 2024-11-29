package pkg_test

import (
	"testing"
	"time"

	"github.com/plugfox/slash-engine-go/pkg"
)

var manager = &pkg.WorldManager{} // Create a global instance of WorldManager for testing

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
		manager.InitWorld(0, -9.8, 100, 100, 16.67, 0.1, false)
		t.Cleanup(manager.StopWorld) // Ensure StopWorld is called after test

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
