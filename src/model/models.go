package model

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID   int64
	Name string
}

// Resolve User From HttpHeader -> X-TOKEN
func ResolveUser(request *http.Request) *User {
	token := request.Header.Get("X-TOKEN")
	if len(token) == 0 {
		return nil
	}
	decoded_token, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return nil
	}
	u := &User{}
	err = json.Unmarshal(decoded_token, u)
	if err != nil {
		return nil
	}
	return u
}

func (u User) GenToken() string {
	jsonBytes, err := json.Marshal(u)
	if err != nil {
		return ""
	}
	return base64.StdEncoding.EncodeToString(jsonBytes)
}

type ServerInfo struct {
	Host string
	Port int
}

var CurrServer ServerInfo = ServerInfo{}

func (s *ServerInfo) ServerAddr() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}

type Message struct {
	ID         int64 `gorm:"primarykey"`
	Type       string
	FromUserID int64
	ToUserID   int64
	Read       bool
	SendTime   *time.Time
	ReadTime   *time.Time
	Data       string
}

// 通过为结构体添加 TableName 字段来指定表名
func (Message) TableName() string {
	return "instant_message"
}

// gorm钩子函数，insert之前触发
func (m *Message) BeforeCreate(tx *gorm.DB) error {
	// 自动填充发送时间
	if m.SendTime == nil {
		Now := time.Now()
		m.SendTime = &Now
	}
	return nil
}
