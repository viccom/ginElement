// plugin_manager.go
package pluginM

import (
	"fmt"
	"sync"
)

var (
	StopChan = make(chan struct{})
	plugins  = make(map[string]Plugin)
	mu       sync.Mutex
)

// RegisterPlugin 注册插件
func RegisterPlugin(plugin Plugin) {
	mu.Lock()
	defer mu.Unlock()
	plugins[plugin.Name()] = plugin
	fmt.Printf("Plugin registered: %s\n", plugin.Name())
}

// GetPlugin 获取插件
func GetPlugin(name string) (Plugin, bool) {
	mu.Lock()
	defer mu.Unlock()
	plugin, exists := plugins[name]
	return plugin, exists
}
