package main

import (
	"fmt"
	"log"
	"time"

	"github.com/benharold/libdrag/pkg/api"
	"github.com/benharold/libdrag/pkg/events"
)

// RaceMonitor demonstrates how to use libdrag's event system
// to build a real-time race monitoring application
type RaceMonitor struct {
	api        *api.LibDragAPI
	raceData   map[string]*RaceData
	currentRace string
}

type RaceData struct {
	RaceID         string
	StartTime      time.Time
	Lane1RT        float64
	Lane2RT        float64
	Lane1Complete  bool
	Lane2Complete  bool
	Winner         int
}

func NewRaceMonitor() *RaceMonitor {
	return &RaceMonitor{
		api:      api.NewLibDragAPI(),
		raceData: make(map[string]*RaceData),
	}
}

func (rm *RaceMonitor) Initialize() error {
	if err := rm.api.Initialize(); err != nil {
		return err
	}

	// Subscribe to all events we care about
	rm.setupEventHandlers()
	return nil
}

func (rm *RaceMonitor) setupEventHandlers() {
	// Race lifecycle events
	rm.api.Subscribe(events.EventRaceStart, func(e events.Event) {
		rm.raceData[e.RaceID] = &RaceData{
			RaceID:    e.RaceID,
			StartTime: e.Timestamp,
		}
		rm.currentRace = e.RaceID
		fmt.Printf("\n🏁 NEW RACE STARTED: %s\n", e.RaceID[:8])
		fmt.Println("================================")
	})

	// Staging events
	rm.api.Subscribe(events.EventTreePreStage, func(e events.Event) {
		fmt.Printf("   Lane %d: Pre-staged ⚪\n", e.Lane)
	})

	rm.api.Subscribe(events.EventTreeStage, func(e events.Event) {
		fmt.Printf("   Lane %d: Staged 🟡\n", e.Lane)
	})

	rm.api.Subscribe(events.EventTreeArmed, func(e events.Event) {
		fmt.Println("\n   🔥 TREE ARMED - Get Ready!")
	})

	// Light sequence events
	rm.api.Subscribe(events.EventTreeSequenceStart, func(e events.Event) {
		sequenceType := e.Data["sequence_type"].(string)
		fmt.Printf("\n   Starting %s Tree Sequence...\n", sequenceType)
	})

	rm.api.Subscribe(events.EventTreeAmberOn, func(e events.Event) {
		if count, ok := e.Data["count"].(int); ok {
			// Pro tree - all three at once
			fmt.Println("   💡💡💡 All ambers ON!")
		} else if number, ok := e.Data["amber_number"].(int); ok {
			// Sportsman tree - sequential
			fmt.Printf("   💡 Amber %d ON\n", number)
		}
	})

	rm.api.Subscribe(events.EventTreeGreenOn, func(e events.Event) {
		fmt.Println("   🟢 GREEN LIGHT! GO GO GO!")
		fmt.Println()
	})

	// Timing events
	rm.api.Subscribe(events.EventTimingReaction, func(e events.Event) {
		rt := e.Data["reaction_time"].(float64)
		if data, ok := rm.raceData[e.RaceID]; ok {
			if e.Lane == 1 {
				data.Lane1RT = rt
			} else {
				data.Lane2RT = rt
			}
		}
		
		if rt < 0 {
			fmt.Printf("   ❌ Lane %d: RED LIGHT! (RT: %.3fs)\n", e.Lane, rt)
		} else {
			fmt.Printf("   ✅ Lane %d: Reaction Time: %.3fs\n", e.Lane, rt)
		}
	})

	rm.api.Subscribe(events.EventTiming60Foot, func(e events.Event) {
		time := e.Data["time"].(float64)
		fmt.Printf("   Lane %d: 60ft - %.3fs\n", e.Lane, time)
	})

	rm.api.Subscribe(events.EventTimingEighthMile, func(e events.Event) {
		time := e.Data["time"].(float64)
		fmt.Printf("   Lane %d: 1/8 mile - %.3fs\n", e.Lane, time)
	})

	rm.api.Subscribe(events.EventTimingQuarterMile, func(e events.Event) {
		time := e.Data["time"].(float64)
		speed := e.Data["trap_speed"].(float64)
		
		if data, ok := rm.raceData[e.RaceID]; ok {
			if e.Lane == 1 {
				data.Lane1Complete = true
			} else {
				data.Lane2Complete = true
			}
			
			// Determine winner if both finished
			if data.Lane1Complete && data.Lane2Complete {
				if data.Lane1RT < 0 && data.Lane2RT >= 0 {
					data.Winner = 2 // Lane 1 red light
				} else if data.Lane2RT < 0 && data.Lane1RT >= 0 {
					data.Winner = 1 // Lane 2 red light
				} else if data.Lane1RT < 0 && data.Lane2RT < 0 {
					data.Winner = 0 // Both red lights
				} else {
					// Compare ETs (simplified - should include reaction time)
					data.Winner = 1 // Would need to track full ET
				}
			}
		}
		
		fmt.Printf("   🏁 Lane %d: FINISH! ET: %.3fs @ %.1f mph\n", e.Lane, time, speed)
	})

	rm.api.Subscribe(events.EventRaceComplete, func(e events.Event) {
		fmt.Println("\n================================")
		fmt.Println("🏆 RACE COMPLETE!")
		
		if data, ok := rm.raceData[e.RaceID]; ok {
			rm.printRaceSummary(data)
		}
	})

	// Foul events
	rm.api.Subscribe(events.EventRaceFoul, func(e events.Event) {
		reason := e.Data["reason"].(string)
		fmt.Printf("   ⚠️  Lane %d: FOUL - %s\n", e.Lane, reason)
	})
}

func (rm *RaceMonitor) printRaceSummary(data *RaceData) {
	fmt.Println("\nRACE SUMMARY:")
	fmt.Printf("Race ID: %s\n", data.RaceID[:8])
	fmt.Printf("Lane 1 RT: %.3fs\n", data.Lane1RT)
	fmt.Printf("Lane 2 RT: %.3fs\n", data.Lane2RT)
	
	if data.Winner > 0 {
		fmt.Printf("\n🥇 WINNER: Lane %d!\n", data.Winner)
	} else if data.Winner == 0 {
		fmt.Println("\n❌ NO WINNER - Both lanes fouled")
	}
}

func (rm *RaceMonitor) StartRace() (string, error) {
	return rm.api.StartRaceWithID()
}

func (rm *RaceMonitor) Shutdown() error {
	return rm.api.Stop()
}

func main() {
	fmt.Println("🏁 LIBDRAG RACE MONITOR")
	fmt.Println("======================")
	fmt.Println("Real-time drag racing event monitor")
	fmt.Println()

	monitor := NewRaceMonitor()
	
	// Initialize the system
	if err := monitor.Initialize(); err != nil {
		log.Fatal("Failed to initialize:", err)
	}
	defer monitor.Shutdown()

	// Start a race
	raceID, err := monitor.StartRace()
	if err != nil {
		log.Fatal("Failed to start race:", err)
	}

	fmt.Printf("Started race: %s\n", raceID[:8])

	// Let the race run
	time.Sleep(10 * time.Second)

	fmt.Println("\n👋 Thanks for using libdrag!")
}
