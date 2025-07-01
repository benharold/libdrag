# libdrag API Examples

## Basic Usage Examples

### Single Race Example

```go
package main

import (
    "encoding/json"
    "fmt"
    "log"
    "time"
    
    "github.com/benharold/libdrag/pkg/api"
)

func main() {
    // Initialize the API
    dragAPI := api.NewLibDragAPI()
    if err := dragAPI.Initialize(); err != nil {
        log.Fatal("Failed to initialize:", err)
    }
    defer dragAPI.Stop()

    // Start a race
    raceID, err := dragAPI.StartRaceWithID()
    if err != nil {
        log.Fatal("Failed to start race:", err)
    }
    
    fmt.Printf("🏁 Race %s started!\n", raceID)

    // Monitor the race in real-time
    for !dragAPI.IsRaceCompleteByID(raceID) {
        // Get current status
        statusJSON := dragAPI.GetRaceStatusJSONByID(raceID)
        
        var status map[string]interface{}
        json.Unmarshal([]byte(statusJSON), &status)
        
        fmt.Printf("State: %s\n", status["state"])
        time.Sleep(100 * time.Millisecond)
    }

    // Get final results
    resultsJSON := dragAPI.GetResultsJSONByID(raceID)
    fmt.Printf("🏆 Final Results: %s\n", resultsJSON)
}
```

### Multiple Concurrent Races

```go
package main

import (
    "fmt"
    "log"
    "sync"
    "time"
    
    "github.com/benharold/libdrag/pkg/api"
)

func main() {
    dragAPI := api.NewLibDragAPI()
    if err := dragAPI.Initialize(); err != nil {
        log.Fatal(err)
    }
    defer dragAPI.Stop()

    // Set higher concurrent race limit
    dragAPI.SetMaxConcurrentRaces(5)

    var wg sync.WaitGroup
    raceCount := 3

    // Start multiple races concurrently
    for i := 0; i < raceCount; i++ {
        wg.Add(1)
        go func(raceNum int) {
            defer wg.Done()
            
            raceID, err := dragAPI.StartRaceWithID()
            if err != nil {
                log.Printf("Race %d failed to start: %v", raceNum, err)
                return
            }
            
            fmt.Printf("🏁 Race %d started with ID: %s\n", raceNum, raceID)
            
            // Wait for completion
            for !dragAPI.IsRaceCompleteByID(raceID) {
                time.Sleep(200 * time.Millisecond)
            }
            
            results := dragAPI.GetResultsJSONByID(raceID)
            fmt.Printf("🏆 Race %d completed: %s\n", raceNum, results)
        }(i + 1)
    }

    wg.Wait()
    fmt.Printf("All races completed. Active races: %d\n", dragAPI.GetActiveRaceCount())
}
```

### Real-time Race Monitoring

```go
package main

import (
    "encoding/json"
    "fmt"
    "log"
    "time"
    
    "github.com/benharold/libdrag/pkg/api"
)

type RaceStatus struct {
    RaceID    string `json:"race_id"`
    State     string `json:"state"`
    StartTime string `json:"start_time,omitempty"`
    Lane1     struct {
        Staged         bool    `json:"staged"`
        ReactionTime   float64 `json:"reaction_time"`
        CurrentPosition float64 `json:"current_position"`
    } `json:"lane_1"`
    Lane2     struct {
        Staged         bool    `json:"staged"`
        ReactionTime   float64 `json:"reaction_time"`
        CurrentPosition float64 `json:"current_position"`
    } `json:"lane_2"`
}

func main() {
    dragAPI := api.NewLibDragAPI()
    if err := dragAPI.Initialize(); err != nil {
        log.Fatal(err)
    }
    defer dragAPI.Stop()

    raceID, err := dragAPI.StartRaceWithID()
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("🏁 Monitoring Race: %s\n", raceID)
    fmt.Println("=" * 50)

    ticker := time.NewTicker(100 * time.Millisecond)
    defer ticker.Stop()

    for !dragAPI.IsRaceCompleteByID(raceID) {
        select {
        case <-ticker.C:
            statusJSON := dragAPI.GetRaceStatusJSONByID(raceID)
            
            var status RaceStatus
            if err := json.Unmarshal([]byte(statusJSON), &status); err != nil {
                continue
            }

            // Display race progress
            fmt.Printf("\rState: %-10s | Lane 1: %6.1fm | Lane 2: %6.1fm", 
                status.State, 
                status.Lane1.CurrentPosition, 
                status.Lane2.CurrentPosition)
        }
    }

    fmt.Println("\n🏆 Race Complete!")
    
    // Show final results
    results := dragAPI.GetResultsJSONByID(raceID)
    fmt.Printf("Results: %s\n", results)
}
```

### Christmas Tree Monitoring

```go
package main

import (
    "encoding/json"
    "fmt"
    "log"
    "time"
    
    "github.com/benharold/libdrag/pkg/api"
)

type TreeStatus struct {
    IsArmed      bool `json:"is_armed"`
    IsRunning    bool `json:"is_running"`
    SequenceType string `json:"sequence_type"`
    LightStates  map[string]map[string]string `json:"light_states"`
}

func displayTree(status TreeStatus) {
    fmt.Println("🎄 Christmas Tree Status:")
    fmt.Printf("   Armed: %t | Running: %t | Type: %s\n", 
        status.IsArmed, status.IsRunning, status.SequenceType)
    
    for lane, lights := range status.LightStates {
        fmt.Printf("   Lane %s: ", lane)
        for light, state := range lights {
            if state == "on" {
                switch light {
                case "pre_stage":
                    fmt.Print("🟡")
                case "stage":
                    fmt.Print("🟡")
                case "amber_1", "amber_2", "amber_3":
                    fmt.Print("🟠")
                case "green":
                    fmt.Print("🟢")
                case "red":
                    fmt.Print("🔴")
                }
            } else {
                fmt.Print("⚫")
            }
        }
        fmt.Println()
    }
}

func main() {
    dragAPI := api.NewLibDragAPI()
    if err := dragAPI.Initialize(); err != nil {
        log.Fatal(err)
    }
    defer dragAPI.Stop()

    raceID, err := dragAPI.StartRaceWithID()
    if err != nil {
        log.Fatal(err)
    }

    for !dragAPI.IsRaceCompleteByID(raceID) {
        treeJSON := dragAPI.GetTreeStatusJSONByID(raceID)
        
        var treeStatus TreeStatus
        if err := json.Unmarshal([]byte(treeJSON), &treeStatus); err == nil {
            fmt.Print("\033[H\033[2J") // Clear screen
            displayTree(treeStatus)
        }
        
        time.Sleep(50 * time.Millisecond)
    }
}
```

### Error Handling Example

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/benharold/libdrag/pkg/api"
)

func main() {
    dragAPI := api.NewLibDragAPI()
    
    // Attempting to start race before initialization
    _, err := dragAPI.StartRaceWithID()
    if err != nil {
        fmt.Printf("Expected error: %v\n", err)
    }

    // Initialize properly
    if err := dragAPI.Initialize(); err != nil {
        log.Fatal(err)
    }
    defer dragAPI.Stop()

    // Set very low race limit for testing
    dragAPI.SetMaxConcurrentRaces(1)

    // Start first race
    race1, err := dragAPI.StartRaceWithID()
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Race 1 started: %s\n", race1)

    // Try to start second race (should fail)
    _, err = dragAPI.StartRaceWithID()
    if err != nil {
        fmt.Printf("Expected limit error: %v\n", err)
    }

    // Check non-existent race
    if !dragAPI.RaceExists("invalid-race-id") {
        fmt.Println("Race validation working correctly")
    }

    // Get status for non-existent race
    status := dragAPI.GetRaceStatusJSONByID("invalid-race-id")
    fmt.Printf("Non-existent race status: %s\n", status)
}
```

### Performance Monitoring

```go
package main

import (
    "fmt"
    "log"
    "runtime"
    "time"
    
    "github.com/benharold/libdrag/pkg/api"
)

func main() {
    dragAPI := api.NewLibDragAPI()
    if err := dragAPI.Initialize(); err != nil {
        log.Fatal(err)
    }
    defer dragAPI.Stop()

    // Monitor system resources
    go func() {
        ticker := time.NewTicker(1 * time.Second)
        defer ticker.Stop()
        
        for {
            select {
            case <-ticker.C:
                var m runtime.MemStats
                runtime.ReadMemStats(&m)
                
                fmt.Printf("Active Races: %d | Memory: %.2f MB | Goroutines: %d\n",
                    dragAPI.GetActiveRaceCount(),
                    float64(m.Alloc)/1024/1024,
                    runtime.NumGoroutine())
            }
        }
    }()

    // Start multiple races to test performance
    dragAPI.SetMaxConcurrentRaces(10)
    
    for i := 0; i < 5; i++ {
        raceID, err := dragAPI.StartRaceWithID()
        if err != nil {
            log.Printf("Failed to start race %d: %v", i, err)
            continue
        }
        
        fmt.Printf("Started race %d: %s\n", i+1, raceID)
        time.Sleep(500 * time.Millisecond)
    }

    // Wait for all races to complete
    for dragAPI.GetActiveRaceCount() > 0 {
        time.Sleep(1 * time.Second)
    }
    
    fmt.Println("All races completed!")
}
```

## Integration Examples

### REST API Server Integration

```go
package main

import (
    "encoding/json"
    "net/http"
    "log"
    
    "github.com/benharold/libdrag/pkg/api"
)

var dragAPI *api.LibDragAPI

func startRaceHandler(w http.ResponseWriter, r *http.Request) {
    raceID, err := dragAPI.StartRaceWithID()
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    response := map[string]string{"race_id": raceID}
    json.NewEncoder(w).Encode(response)
}

func getRaceStatusHandler(w http.ResponseWriter, r *http.Request) {
    raceID := r.URL.Query().Get("race_id")
    if raceID == "" {
        http.Error(w, "race_id required", http.StatusBadRequest)
        return
    }
    
    status := dragAPI.GetRaceStatusJSONByID(raceID)
    w.Header().Set("Content-Type", "application/json")
    w.Write([]byte(status))
}

func main() {
    dragAPI = api.NewLibDragAPI()
    if err := dragAPI.Initialize(); err != nil {
        log.Fatal(err)
    }
    defer dragAPI.Stop()

    http.HandleFunc("/start", startRaceHandler)
    http.HandleFunc("/status", getRaceStatusHandler)
    
    log.Println("Server starting on :8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}
```

These examples demonstrate the key features and usage patterns of the libdrag API, from basic single races to complex concurrent scenarios and system integration.
