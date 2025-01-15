package assistant

import (
	"github.com/gorilla/websocket"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

func forwardRequestToLocalServer(localURL string, path string) ([]byte, error) {
	// 将请求转发到本地Web服务
	resp, err := http.Get(localURL + path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func reconnect(wsURL string, conn *websocket.Conn, UUID string, Port string) {
	dialer := websocket.Dialer{}
	for {
		var err error
		conn, _, err = dialer.Dial(wsURL, nil)
		if err != nil {
			log.Println("重连失败，正在重试...")
			time.Sleep(time.Second * 5)
			continue
		}
		// 重新发送注册信息
		regMsg := struct {
			UUID string `json:"uuid"`
			Port string `json:"port"`
		}{
			UUID: UUID,
			Port: Port,
		}
		err = conn.WriteJSON(regMsg)
		if err != nil {
			log.Println("重新发送注册信息失败:", err)
			continue
		}
		log.Println("Reconnected to server")
		break
	}
}
