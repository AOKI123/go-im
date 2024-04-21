package model

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"time"
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

type Message struct {
	ID         int64
	Type       string
	FromUserID int64
	ToUserID   int64
	Read       bool
	SendTime   time.Time
	ReadTime   time.Time
	Data       interface{}
}
