package handlers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nalgeon/redka"
	"net/http"
	"sync"
)

type testFunc func(id string, stopChan chan struct{})
type iotFunc func(id string, stopChan chan struct{}, cfgdb *redka.DB, rtdb *redka.DB)

// 全局变量
var (
	Workers     = make(map[string]chan struct{}) // 存储所有子线程的停止通道
	workersLock sync.Mutex                       // 用于保护 Workers 的并发访问
	//nextID      = 1                              // 用于生成唯一的子线程 ID

	// 定义redka数据库中的表名
	InstListKey  = "inst@router"
	DevAtInstKey = "dev@inst"
	// 定义字符串数组
	funcCode = []string{"findmax", "periodicPrint"}
	funcMap  = map[string]testFunc{
		"findmax":       findmax,
		"periodicPrint": PeriodicPrint,
	}
	// 定义字符串数组
	iotappCode = []string{"simulator", "modbus", "opcda", "opcua", "mqttpub"}
	IotappMap  = map[string]iotFunc{
		"simulator": Simulator,
		"modbus":    ModbusRead,
		"opcda":     OpcDARead,
		"opcua":     OpcUARead,
		"mqttpub":   mqttPubData,
	}
)

// 定义 DataQueue 结构
type DataQueue struct {
	data  []string
	mutex sync.Mutex
	cond  *sync.Cond
}

// 初始化 DataQueue
func NewDataQueue() *DataQueue {
	q := &DataQueue{
		data: make([]string, 0),
	}
	q.cond = sync.NewCond(&q.mutex)
	return q
}

// 入队操作
func (q *DataQueue) Enqueue(d string) {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	q.data = append(q.data, d)
	q.cond.Signal() // 通知等待的消费者
}

// 出队操作
func (q *DataQueue) Dequeue() (string, bool) {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	// 如果队列为空，等待数据
	for len(q.data) == 0 {
		q.cond.Wait()
	}

	val := q.data[0]
	q.data = q.data[1:]
	return val, true
}

// 获取队列长度
func (q *DataQueue) Len() int {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	return len(q.data)
}

// @Summary 启动线程运行函数接口
// @Description 这是一个启动线程的接口
// @Tags 示例
// @Accept json
// @Produce json
// @Param appCode path string true "功能appcode"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /api/v1/startWorker/{appcode} [post]
func StartWorker(c *gin.Context) {
	workersLock.Lock()
	defer workersLock.Unlock()

	appcode := c.Param("appCode")
	// 检查 appcode 是否有效
	if !contains(funcCode, appcode) {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid appCode",
			"details": fmt.Sprintf("appCode '%s' is not supported", appcode),
		})
		return
	}

	// 检查 funcMap 中是否存在对应的函数
	fn, exists := funcMap[appcode]
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Function not found",
			"details": fmt.Sprintf("appCode '%s' has no associated function", appcode),
		})
		return
	}

	// 生成一个新的 UUID
	newUUID := uuid.New()
	uuidstr := appcode + "@" + newUUID.String()

	// 创建停止通道
	stopChan := make(chan struct{})

	// 启动子线程
	go func() {
		defer func() {
			// 使用 select 检查 channel 是否已关闭
			select {
			case <-stopChan:
				// channel 已关闭，无需再次关闭
			default:
				close(stopChan) // 关闭 channel
			}
		}()
		fn(uuidstr, stopChan) // 调用对应的函数
	}()
	//fn(uuidstr, stopChan) // 调用对应的函数
	// 将子线程的停止通道存储到全局变量中
	Workers[uuidstr] = stopChan

	// 返回子线程 ID
	c.JSON(http.StatusOK, gin.H{
		"message": "Worker started",
		"id":      uuidstr,
	})
}

// @Summary 停止指定线程接口stopWorker
// @Description 这是一个停止线程的接口
// @Tags 示例
// @Accept json
// @Produce json
// @Param workerid path string true "线程 workerid"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /api/v1/stopWorker/{workerid} [post]
func StopWorker(c *gin.Context) {
	workersLock.Lock()
	defer workersLock.Unlock()
	// 获取子线程 ID
	workerid := c.Param("workerid")
	var workerID string
	_, err := fmt.Sscanf(workerid, "%v", &workerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid ID",
		})
		return
	}
	// 查找子线程的停止通道
	stopChan, exists := Workers[workerID]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "Worker not found",
		})
		return
	}
	// 发送停止信号
	close(stopChan)
	// 从全局变量中移除子线程
	delete(Workers, workerID)
	// 返回成功消息
	c.JSON(http.StatusOK, gin.H{
		"message": "Worker stopped",
		"id":      workerID,
	})
}

// @Summary 查询后台运行的所有线程ID
// @Description 这是一个查询线程ID的接口
// @Tags 示例
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /api/v1/listWorkers [get]
func ListWorkers(c *gin.Context) {
	workersLock.Lock()
	defer workersLock.Unlock()

	// 获取所有子线程的 ID
	ids := make([]string, 0, len(Workers))
	for id := range Workers {
		ids = append(ids, id)
	}
	if len(ids) == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "No workers running",
		})
		return
	}
	// 返回子线程 ID 列表
	c.JSON(http.StatusOK, gin.H{
		"message": "Workers running",
		"data":    ids,
	})
}

// @Summary 查询软件当前支持的Appcode
// @Description 这是一个查询Appcode的接口
// @Tags APP Manager
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /api/v1/listAppcode [get]
func ListAppcode(c *gin.Context) {
	// 返回 funcCode	列表
	c.JSON(http.StatusOK, gin.H{
		"message": "appCode",
		"appCode": iotappCode,
	})
}
