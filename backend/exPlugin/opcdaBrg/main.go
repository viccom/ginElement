package main

import (
	"flag"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	mqBroker "github.com/mochi-mqtt/server/v2"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
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
)

func main() {
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

	mqttClient = newMQTTClient(*broker, *clientID, *username, *password)
	// Subscribe to command topic
	if token := mqttClient.client.Subscribe("command", 1, handleCommand); token.Wait() && token.Error() != nil {
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
