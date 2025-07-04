package events

import (
	"sync"
	"testing"
	"time"
)

func TestEventBusSync(t *testing.T) {
	eb := NewEventBus(false) // Sync mode

	received := false
	unsubscribe := eb.Subscribe(EventTreeGreenOn, func(event Event) {
		received = true
		if event.Type != EventTreeGreenOn {
			t.Errorf("Expected event type %s, got %s", EventTreeGreenOn, event.Type)
		}
	})
	defer unsubscribe()

	// Publish event
	eb.Publish(NewEvent(EventTreeGreenOn).Build())

	if !received {
		t.Error("Handler was not called")
	}
}

func TestEventBusAsync(t *testing.T) {
	eb := NewEventBus(true) // Async mode
	defer eb.Stop()

	var wg sync.WaitGroup
	wg.Add(1)

	eb.Subscribe(EventRaceStart, func(event Event) {
		defer wg.Done()
		if event.Type != EventRaceStart {
			t.Errorf("Expected event type %s, got %s", EventRaceStart, event.Type)
		}
	})

	// Publish event
	eb.Publish(NewEvent(EventRaceStart).WithRaceID("test-race").Build())

	// Wait for async delivery
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Success
	case <-time.After(time.Second):
		t.Error("Async event not delivered within timeout")
	}
}

func TestEventBuilder(t *testing.T) {
	event := NewEvent(EventTiming60Foot).
		WithRaceID("race-123").
		WithLane(1).
		WithData("time", 1.234).
		Build()

	if event.Type != EventTiming60Foot {
		t.Errorf("Expected type %s, got %s", EventTiming60Foot, event.Type)
	}
	if event.RaceID != "race-123" {
		t.Errorf("Expected race ID 'race-123', got %s", event.RaceID)
	}
	if event.Lane != 1 {
		t.Errorf("Expected lane 1, got %d", event.Lane)
	}
	if event.Data["time"] != 1.234 {
		t.Errorf("Expected time 1.234, got %v", event.Data["time"])
	}
}

func TestSubscribeAll(t *testing.T) {
	eb := NewEventBus(false)

	receivedEvents := make([]EventType, 0)
	eb.SubscribeAll(func(event Event) {
		receivedEvents = append(receivedEvents, event.Type)
	})

	// Publish different event types
	eb.Publish(NewEvent(EventTreeArmed).Build())
	eb.Publish(NewEvent(EventRaceStart).Build())
	eb.Publish(NewEvent(EventTimingReaction).Build())

	if len(receivedEvents) != 3 {
		t.Errorf("Expected 3 events, received %d", len(receivedEvents))
	}

	expectedTypes := []EventType{EventTreeArmed, EventRaceStart, EventTimingReaction}
	for i, eventType := range expectedTypes {
		if receivedEvents[i] != eventType {
			t.Errorf("Expected event %d to be %s, got %s", i, eventType, receivedEvents[i])
		}
	}
}

func TestMultipleHandlers(t *testing.T) {
	eb := NewEventBus(false)

	count := 0
	handler1 := func(event Event) { count++ }
	handler2 := func(event Event) { count++ }
	handler3 := func(event Event) { count++ }

	eb.Subscribe(EventTreeGreenOn, handler1)
	eb.Subscribe(EventTreeGreenOn, handler2)
	eb.Subscribe(EventTreeGreenOn, handler3)

	eb.Publish(NewEvent(EventTreeGreenOn).Build())

	if count != 3 {
		t.Errorf("Expected 3 handlers to be called, but count is %d", count)
	}
}

func TestEventBusStop(t *testing.T) {
	eb := NewEventBus(true)

	// Subscribe to an event
	received := make(chan bool, 1)
	eb.Subscribe(EventRaceComplete, func(event Event) {
		received <- true
	})

	// Publish event
	eb.Publish(NewEvent(EventRaceComplete).Build())

	// Wait for event
	select {
	case <-received:
		// Good
	case <-time.After(100 * time.Millisecond):
		t.Error("Event not received before stop")
	}

	// EmergencyStop the event bus
	eb.Stop()

	// Try to publish after stop - should not panic
	eb.Publish(NewEvent(EventRaceComplete).Build())
}

func BenchmarkEventBusSync(b *testing.B) {
	eb := NewEventBus(false)
	eb.Subscribe(EventTimingBeamTrigger, func(event Event) {
		// Do nothing
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		eb.Publish(NewEvent(EventTimingBeamTrigger).Build())
	}
}

func BenchmarkEventBusAsync(b *testing.B) {
	eb := NewEventBus(true)
	defer eb.Stop()

	eb.Subscribe(EventTimingBeamTrigger, func(event Event) {
		// Do nothing
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		eb.Publish(NewEvent(EventTimingBeamTrigger).Build())
	}
}
