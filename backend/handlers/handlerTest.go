package handlers

import (
	"fmt"
	"github.com/nalgeon/redka"
	"math/rand"
	"time"
)

// 判断字符串是否在数组中
func contains(arr []string, target string) bool {
	// 将数组转换为 map
	set := make(map[string]struct{})
	for _, item := range arr {
		set[item] = struct{}{}
	}

	// 判断目标字符串是否在 map 中
	_, exists := set[target]
	return exists
}

// 打印当前时间
func printTime() {
	// 定义时区
	var (
		beijingLocation    = time.FixedZone("CST", 8*60*60)  // 北京时间 (UTC+8)
		washingtonLocation = time.FixedZone("EDT", -4*60*60) // 华盛顿时间 (UTC-4)
		moscowLocation     = time.FixedZone("MSK", 3*60*60)  // 莫斯科时间 (UTC+3)
	)
	// 获取当前 UTC 时间
	now := time.Now().UTC()

	// 转换为不同时区的时间
	beijingTime := now.In(beijingLocation)
	washingtonTime := now.In(washingtonLocation)
	moscowTime := now.In(moscowLocation)

	// 打印时间
	fmt.Printf("北京时间: %s\n", beijingTime.Format("2006-01-02 15:04:05"))
	fmt.Printf("华盛顿时间: %s\n", washingtonTime.Format("2006-01-02 15:04:05"))
	fmt.Printf("莫斯科时间: %s\n", moscowTime.Format("2006-01-02 15:04:05"))
	fmt.Println("----------------------------------------")
}

// periodicPrint函数，周期打印时间
func periodicPrint(id string, stopChan chan struct{}, rtdb *redka.DB) {
	_ = rtdb.Key()
	for {
		select {
		case <-stopChan: // 如果收到停止信号，退出循环
			fmt.Printf("Worker %v stopped\n", id)
			return
		default:
			printTime()
			// 等待 1 秒
			time.Sleep(5 * time.Second)
		}
	}
}

// findmax函数：周期性地从 10 个随机数中找到最大值并打印
func findmax(id string, stopChan chan struct{}, rtdb *redka.DB) {
	// 使用当前时间的纳秒级时间戳作为种子
	_ = rtdb.Key()
	source := rand.NewSource(time.Now().UnixNano())
	r := rand.New(source)
	for {
		select {
		case <-stopChan: // 如果收到停止信号，退出循环
			fmt.Printf("Worker %v stopped\n", id)
			return
		default:
			// 生成 10 个随机数
			nums := make([]int, 10)
			for i := 0; i < 10; i++ {
				nums[i] = r.Intn(100)
			}
			// 找到最大值
			m := nums[0]
			for _, num := range nums {
				if num > m {
					m = num
				}
			}
			// 打印结果
			fmt.Printf("Worker %v: Max number is %d\n", id, m)
			// 等待 1 秒
			time.Sleep(1 * time.Second)
		}
	}
}

// simtodb函数：周期性地从 10 个随机数中找到最大值并打印
func simtodb(id string, stopChan chan struct{}, rtdb *redka.DB) {
	// 使用当前时间的纳秒级时间戳作为种子
	source := rand.NewSource(time.Now().UnixNano())
	r := rand.New(source)
	for {
		select {
		case <-stopChan: // 如果收到停止信号，退出循环
			fmt.Printf("Worker %v stopped\n", id)
			return
		default:
			// 生成 10 个随机数
			nums := r.Intn(100)
			_, err := rtdb.Hash().Set("test", "test", nums)
			if err != nil {
				return
			}
			// 打印结果
			fmt.Printf("Worker %v: Random number is %d\n", id, nums)
			// 等待 1 秒
			time.Sleep(1 * time.Second)
		}
	}
}
