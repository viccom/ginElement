// plugin.go
package pluginM

import "sync"

// Plugin 接口
type Plugin interface {
	Name() string                                                             // 插件名称
	Execute(appConfig map[string]string, devices map[string]map[string][]any) // 插件功能
}

type SafeQueue struct {
	mu    sync.Mutex
	queue []map[string]any
}

func (q *SafeQueue) Enqueue(item map[string]any) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.queue = append(q.queue, item)
}

func (q *SafeQueue) Dequeue() (map[string]any, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if len(q.queue) == 0 {
		return nil, false
	}
	item := q.queue[0]
	q.queue = q.queue[1:]
	return item, true
}
