package main

import (
	"encoding/json"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	mqBroker "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/hooks/auth"
	"github.com/mochi-mqtt/server/v2/listeners"
	"log"
	"os"
	"time"
)

func dataPub(pubID string) {
	defer wg.Done()
	for {
		select {
		case <-pubStopChan:
			return
		default:
			if data, ok := dataQueue.Dequeue(); ok {
				jsonData, err := json.Marshal(data)
				if err != nil {
					log.Println("Error marshaling data:", err)
					continue
				}

				if token := mqttClient.client.Publish(pubID+"/data", 0, false, jsonData); token.Wait() && token.Error() != nil {
					log.Println("Error publishing data:", token.Error())
				} else {
					log.Printf("Published data to %s topic: %s\n", mqttClient.topic, string(jsonData))
				}
			}
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func startLocalBroker(path string) (*mqBroker.Server, error) {
	// Get ledger from yaml file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	server := mqBroker.New(nil)
	err = server.AddHook(new(auth.Hook), &auth.Options{
		Data: data, // build ledger from byte slice, yaml or json
	})
	if err != nil {
		return nil, err
	}

	tcp := listeners.NewTCP(listeners.Config{
		ID:      "t1",
		Address: ":1883",
	})
	err = server.AddListener(tcp)
	if err != nil {
		return nil, err
	}

	go func() {
		err := server.Serve()
		if err != nil {
			log.Fatal(err)
		}
	}()

	return server, nil
}

func newMQTTClient(broker string, clientID string, userName string, passWord string) *MQTTClient {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(broker)
	opts.SetClientID(clientID)
	opts.SetUsername(userName)
	opts.SetPassword(passWord)

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatal(token.Error())
	}

	return &MQTTClient{
		client: client,
	}
}
