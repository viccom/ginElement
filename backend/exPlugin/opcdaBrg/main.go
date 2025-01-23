package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
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
)

const version = "20250123"

var (
	apiServer *http.Server
)

func startLocalApi() {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/sysinfo", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		hwid, err := GetHardwareID()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		sysinfo := map[string]string{
			"hwid":    hwid,
			"version": version,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(sysinfo)
	})

	mux.HandleFunc("/api/v1/sysstatus", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		status := map[string]bool{
			"simRunning": simRunning,
			"pubRunning": pubRunning,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(status)
	})

	apiServer = &http.Server{
		Addr:    ":7780",
		Handler: mux,
	}

	go func() {
		log.Println("Starting local API server on :7780")
		if err := apiServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Could not start API server: %v", err)
		}
	}()
}

func stopLocalApi() {
	if apiServer != nil {
		log.Println("Shutting down API server...")
		if err := apiServer.Close(); err != nil {
			log.Printf("Error closing API server: %v", err)
		}
	}
}

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
