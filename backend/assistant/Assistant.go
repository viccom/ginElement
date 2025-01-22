package assistant

import (
	"ehang.io/nps/client"
	"github.com/astaxie/beego/logs"
	"time"
)

func NpcRun(serverAddr string, verifyKey string, connType string, disconnectTime int) error {
	go func() {
		for {
			rpClient := client.NewRPClient(serverAddr, verifyKey, connType, "", nil, disconnectTime)
			// Start() doesn't return an error but we can still log if the rpClient is nil
			if rpClient == nil {
				logs.Error("Failed to create NPC rpClient")
			} else {
				rpClient.Start()
			}
			logs.Info("Client closed! It will be reconnected in five seconds")
			time.Sleep(time.Second * 5)
		}
	}()

	return nil
}
