package conn

import (
	"fmt"
	"log"
	"sync"

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
	m.connections.Store(userID, conn)
}

func (m *ConnManager) DelConn(userID int64) {
	m.connections.Delete(userID)
}

type Sender interface {
	Send(msg model.Message)
	SendUnRead(userID int64)
}

type LocalSender struct {
}

func (sender *LocalSender) Send(msg *model.Message) error {
	// 保存数据库
	repo.MessageRepoOps.Create(msg)
	conn := GetConnManager().FindConn(msg.ToUserID)
	if conn == nil {
		return fmt.Errorf("%v does not online", msg.ToUserID)
	}
	return conn.WriteJSON(msg)
}

func (LocalSender) SendUnRead(userID int64, conn *websocket.Conn) error {
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
