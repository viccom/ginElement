package handlers

import (
	"fmt"
	"github.com/nalgeon/redka"
	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"
	"time"
)

// handlermobus 函数：周期性地从 10 个随机数中找到最大值并打印
func handlermobus(id string, stopChan chan struct{}, cfgdb *redka.DB, rtdb *redka.DB) {
	_ = rtdb.Key()
	for {
		select {
		case <-stopChan: // 如果收到停止信号，退出循环
			fmt.Printf("Worker %v stopped\n", id)
			return
		default:
			// 打印结果
			fmt.Printf("Worker %v\n", id)
			// 等待 1 秒
			time.Sleep(1 * time.Second)
		}
	}
}
