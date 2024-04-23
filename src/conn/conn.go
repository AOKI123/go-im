package conn

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"aoki.com/go-im/src/model"
	"aoki.com/go-im/src/repo"
	"github.com/gorilla/websocket"
)

type ConnManager struct {
	connections sync.Map
}

var m *ConnManager = &ConnManager{
	connections: sync.Map{},
}

func GetConnManager() *ConnManager {
	return m
}

func (m *ConnManager) FindConn(userID int64) *websocket.Conn {
	conn, has := m.connections.Load(userID)
	if has {
		return conn.(*websocket.Conn)
	}
	return nil
}

func (m *ConnManager) AddConn(userID int64, conn *websocket.Conn) {
	// 保存本地
	m.connections.Store(userID, conn)
	// 保存当前连接的服务器信息
	err := SetRedisKVWithTTL(UserConnKey(userID), model.CurrServer.ServerAddr(), 5*time.Second)
	if err != nil {
		log.Printf("Save conn of %d to redis failed:\n %v\n", userID, err)
	}
}

func (m *ConnManager) SetConnTTL(userID int64) {
	conn := m.FindConn(userID)
	if conn == nil {
		return
	}
	err := SetRedisKVWithTTL(UserConnKey(userID), model.CurrServer.ServerAddr(), 5*time.Second)
	if err != nil {
		log.Printf("Save conn of %d to redis failed:\n %v\n", userID, err)
	}
}

func (m *ConnManager) DelConn(userID int64) {
	// 删除本地记录
	m.connections.Delete(userID)
	// 删除Redis中的信息
	err := DelRedisKey(UserConnKey(userID))
	if err != nil {
		log.Printf("Del conn of %d to redis failed:\n %v\n", userID, err)
	}
}

func UserConnKey(userID int64) string {
	return fmt.Sprintf("conn:%d", userID)
}

type Sender struct {
}

func (sender *Sender) Send(msg *model.Message) error {
	// 保存数据库
	repo.MessageRepoOps.Create(msg)
	conn := GetConnManager().FindConn(msg.ToUserID)
	if conn == nil {
		// 本地没有，查询redis
		address, err := GetRedisVal(UserConnKey(msg.ToUserID))
		if err != nil {
			return nil
		}
		// call http send
		if len(address) > 0 {
			return sender.callRemoteSend(msg, address)
		}
		return fmt.Errorf("%v does not online", msg.ToUserID)
	}
	return conn.WriteJSON(msg)
}

func (Sender) SendUnRead(userID int64, conn *websocket.Conn) error {
	// 获取当前用户未读的消息
	messages := repo.MessageRepoOps.FindUnRead(userID)
	if len(messages) == 0 {
		return nil
	}
	log.Printf("Send unread to %d", userID)
	var err error
	for _, msg := range messages {
		err = conn.WriteJSON(msg)
	}
	return err
}

func (Sender) callRemoteSend(msg *model.Message, remoteAddr string) (err error) {
	var (
		bodyBytes []byte
		req       *http.Request
		resp      *http.Response
	)
	bodyBytes, err = json.Marshal(msg)
	if err != nil {
		return err
	}
	// 创建一个请求体，这里使用的是JSON格式的数据
	body := bytes.NewBufferString(string(bodyBytes))
	url := fmt.Sprintf("http://%s/send", remoteAddr)
	log.Printf("call POST %s", url)
	req, err = http.NewRequest("POST", url, body)
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	// 使用http.DefaultClient.Do方法来发送请求
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	// 读取响应体
	bodyBytes, err = io.ReadAll(resp.Body)
	if err != nil {
		return
	}
	log.Printf("call remote send resp: %s \n", string(bodyBytes))
	return
}
