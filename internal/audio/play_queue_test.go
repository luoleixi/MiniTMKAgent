package audio

import (
	"testing"
	"time"
)

func TestPlayQueue(t *testing.T) {
	queue := &PlayQueue{
		queue:       make(chan QueuedAudio, 100),
		pending:     make(map[int][]byte),
		nextSeq:     1,
		stopChan:    make(chan struct{}),
		sequenceGen: NewSequenceGenerator(),
	}

	t.Run("Sequence Generator", func(t *testing.T) {
		gen := NewSequenceGenerator()

		// Test sequential generation
		if got := gen.Next(); got != 1 {
			t.Errorf("First Next() = %d, want 1", got)
		}
		if got := gen.Next(); got != 2 {
			t.Errorf("Second Next() = %d, want 2", got)
		}
		if got := gen.Current(); got != 2 {
			t.Errorf("Current() = %d, want 2", got)
		}

		// Test reset
		gen.Reset()
		if got := gen.Current(); got != 0 {
			t.Errorf("After Reset(), Current() = %d, want 0", got)
		}
	})

	t.Run("Enqueue without sequence", func(t *testing.T) {
		seq := queue.EnqueueWithoutSeq([]byte("test"))
		if seq < 1 {
			t.Errorf("EnqueueWithoutSeq returned invalid sequence: %d", seq)
		}
	})

	t.Run("Queue size", func(t *testing.T) {
		initialSize := queue.QueueSize()
		queue.Enqueue(100, []byte("data"))
		if queue.QueueSize() != initialSize+1 {
			t.Errorf("Queue size should increase after enqueue")
		}
	})
}

func TestSequenceGenerator_Concurrent(t *testing.T) {
	gen := NewSequenceGenerator()

	// Test concurrent access
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 10; j++ {
				gen.Next()
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		select {
		case <-done:
		case <-time.After(1 * time.Second):
			t.Fatal("Timeout waiting for goroutines")
		}
	}

	// Check final value (should be 100)
	if got := gen.Current(); got != 100 {
		t.Errorf("After concurrent access, Current() = %d, want 100", got)
	}
}
