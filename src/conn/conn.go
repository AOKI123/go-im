package conn

import (
	"fmt"
	"sync"

	"aoki.com/go-im/src/model"
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
}

type LocalSender struct {
}

func (sender *LocalSender) Send(msg model.Message) error {
	conn := GetConnManager().FindConn(msg.ToUserID)
	if conn == nil {
		return fmt.Errorf("%v does not online", msg.ToUserID)
	}
	return conn.WriteJSON(msg)
}
