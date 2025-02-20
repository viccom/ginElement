package plugins

import (
	"fmt"
	"log"
	"math/rand/v2"
	"opcdaBrg/pluginM"
	"sync"
	"time"
)

var (
	dataQueue   pluginM.SafeQueue
	simStopChan chan struct{}
	wg          sync.WaitGroup
)

// dataSim 插件
type dataSim struct{}

func (p *dataSim) Name() string {
	return "dataSim"
}

func (p *dataSim) Execute(appConfig map[string]string, devices map[string]map[string][]any) {
	xxx(appConfig, devices)
}

// 初始化时注册插件
func init() {
	pluginM.RegisterPlugin(&dataSim{})
}

func xxx(appConfig map[string]string, devices map[string]map[string][]any) {
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
