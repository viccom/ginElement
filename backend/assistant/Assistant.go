package assistant

import (
	"ehang.io/nps/client"
	"github.com/astaxie/beego/logs"
	"time"
)

var (
	serverAddr     = "nps.metme.top:7088"
	verifyKey      = "84dce5a776bf44bba953aaf2f108fbda"
	connType       = "tcp"
	disconnectTime = 60
)

func NpcRun() error {
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
