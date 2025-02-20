package plugins

import (
	"fmt"
	"log"
	"math/rand/v2"
	"opcdaBrg/pluginM"
	"time"
)

// HelloPlugin 插件
type HelloPlugin struct{}

func (p *HelloPlugin) Name() string {
	return "HelloPlugin"
}

func (p *HelloPlugin) Execute(appConfig map[string]string, devices map[string]map[string][]any) {
	wg.Add(1) // 改为在 Execute 内部添加
	defer wg.Done()
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()
	fmt.Printf("appConfig: %+v\n", appConfig)
	fmt.Printf("devices: %+v\n", devices)
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
		case <-pluginM.StopChan:
			log.Println("Shutting down HelloPlugin...")
			return
		}
	}

}

// 初始化时注册插件
func init() {
	pluginM.RegisterPlugin(&HelloPlugin{})
}
