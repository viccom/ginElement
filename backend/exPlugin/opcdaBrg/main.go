package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	mqBroker "github.com/mochi-mqtt/server/v2"
)

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

type MQTTClient struct {
	client mqtt.Client
	topic  string
}

var (
	dataQueue   SafeQueue
	simRunning  bool
	pubRunning  bool
	simStopChan chan struct{}
	pubStopChan chan struct{}
	wg          sync.WaitGroup
	mqttClient  *MQTTClient
	hwid        string
	appVer      int = 0
)

const version = "20250123"

var (
	apiServer *http.Server
)

func main() {
	// Start API server in a goroutine
	go startLocalApi()
	defer stopLocalApi()
	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		done <- true
	}()

	// Parse command line arguments
	localBroker := flag.Bool("localbroker", true, "enable local broker")
	broker := flag.String("broker", "tcp://localhost:1883", "MQTT broker address")
	clientID := flag.String("clientID", "opcdaBrg", "MQTT client ID")
	username := flag.String("username", "melon", "MQTT username")
	password := flag.String("password", "password2", "MQTT password")
	path := flag.String("path", "brokerAuth.json", "path to mqttBroker auth file")
	flag.Parse()

	var mqServer *mqBroker.Server
	var err error
	if *localBroker {
		mqServer, err = startLocalBroker(*path) // 使用 = 赋值，修改外部的 mqServer
		if err != nil {
			log.Fatal(err)
		}
	}

	hwid, err = GetHardwareID()
	mqttClient = newMQTTClient(*broker, *clientID, *username, *password)
	// Subscribe to command topic
	if token := mqttClient.client.Subscribe(hwid+"/command", 1, handleCommand); token.Wait() && token.Error() != nil {
		log.Fatal(token.Error())
	}

	<-done
	if mqServer != nil { // 检查 mqServer 是否为 nil
		mqServer.Log.Warn("caught signal, stopping...")
		_ = mqServer.Close()
	}
	mqttClient.client.Disconnect(250)
	log.Println("main.go finished")
}

func handleCommand(client mqtt.Client, msg mqtt.Message) {
	log.Printf("Received command: %s\n", string(msg.Payload()))

	var cmd struct {
		Start     bool                        `json:"start"`
		Ver       int                         `json:"ver"`
		AppConfig map[string]string           `json:"appconfig"`
		Devices   map[string]map[string][]any `json:"devices"`
	}
	if err := json.Unmarshal(msg.Payload(), &cmd); err != nil {
		log.Println("Error parsing command:", err)
		return
	}
	if !cmd.Start {
		var reason string
		if simRunning {
			close(simStopChan)
			close(pubStopChan)
			wg.Wait()
			simRunning = false
			pubRunning = false
			log.Println("Stopped data simulation and publishing")
			reason = "Stopped data simulation and publishing"
		} else {
			log.Println("No running simulation to stop")
			reason = "No running simulation to stop"
		}
		// 发送返回消息
		result := map[string]any{
			"result": "success",
			"reason": reason,
			"cmd":    cmd,
		}
		cmdJSON, _ := json.Marshal(result)
		if token := mqttClient.client.Publish(hwid+"/command/result", 0, false, cmdJSON); token.Wait() && token.Error() != nil {
			log.Println("Error sending command:", token.Error())
		}
		return
	}
	if cmd.Start {
		fmt.Printf("cmd.Ver: %+v, appVer: %+v\n", cmd.Ver, appVer)
		if cmd.Ver > appVer {
			appVer = cmd.Ver
			if simRunning {
				close(simStopChan)
				close(pubStopChan)
				wg.Wait()
				simRunning = false
				pubRunning = false
				log.Println("Restart: So First Stopped data simulation and publishing")
			}
			if !simRunning {
				simRunning = true
				pubRunning = true
				simStopChan = make(chan struct{})
				pubStopChan = make(chan struct{})
				wg.Add(2)
				go dataSim(cmd.AppConfig, cmd.Devices)
				go dataPub(hwid)
				log.Println("Started data simulation and publishing")
			}
			// 发送返回消息
			result := map[string]any{
				"result": "success",
				"reason": "new version is updated",
				"cmd":    cmd,
			}
			cmdJSON, _ := json.Marshal(result)
			if token := mqttClient.client.Publish(hwid+"/command/result", 0, false, cmdJSON); token.Wait() && token.Error() != nil {
				log.Println("Error sending command:", token.Error())
			}
		} else {
			// 发送返回消息
			result := map[string]any{
				"result": "failed",
				"reason": "current version " + strconv.Itoa(appVer) + " is newest",
				"cmd":    cmd,
			}
			cmdJSON, _ := json.Marshal(result)
			if token := mqttClient.client.Publish(hwid+"/command/result", 0, false, cmdJSON); token.Wait() && token.Error() != nil {
				log.Println("Error sending command:", token.Error())
			}
		}
	}
}
