package main

import (
	"fmt"
	"log"
	"math/rand/v2"
	"time"
)

func dataSim(appConfig map[string]string, devices map[string]map[string][]any) {
	defer wg.Done()
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	fmt.Printf("appConfig: %+v\n", appConfig)
	fmt.Printf("devices: %+v\n", devices)
	// 模拟数据生成逻辑
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

//start cmd
//{
//    "start": true,
//    "ver": 1,
//    "appconfig": {
//        "instid": "opcda@434uyjhgwqe"
//    },
//    "devices": {
//        "DEV_12345678": {
//            "tags": [
//                [
//                    "tag1",
//                    "tag1",
//                    "uuid",
//                    1,
//                    3,
//                    "int16"
//                ],
//                [
//                    "tag2",
//                    "tag2",
//                    "uuid",
//                    1,
//                    3,
//                    "int16"
//                ]
//            ]
//        }
//    }
//}
//stop cmd
//{
//  "start": false
//}
