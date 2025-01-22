package main

import (
	"log"
	"math/rand/v2"
	"time"
)

func dataSim() {
	defer wg.Done()
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			data := map[string]any{
				"value1":    rand.Float64(),
				"value2":    rand.Float64(),
				"value3":    rand.Float64(),
				"value4":    rand.Float64(),
				"value5":    rand.Float64(),
				"timestamp": time.Now().Unix(),
			}
			dataQueue.Enqueue(data)
			log.Printf("Generated new data: %v\n", data)
		case <-simStopChan:
			return
		}
	}
}
