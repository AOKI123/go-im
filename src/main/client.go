package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"aoki.com/go-im/src/model"
	"github.com/gorilla/websocket"
)

func main() {
	uri := "ws://localhost:8848/ws"
	args := os.Args[1:]
	userID, er := strconv.ParseInt(args[0], 10, 64)
	if er != nil {
		log.Fatal("Parse UserId Failed:", er)
	}
	u := model.User{
		ID:   userID,
		Name: args[1],
	}
	token := u.GenToken()
	header := http.Header{
		"X-TOKEN": []string{token},
	}
	conn, err := connect(uri, header)
	if err != nil {
		log.Fatal("dial failed:", err)
	}
	defer conn.Close()

	go func() {
		for {
			m := &model.Message{}
			err = conn.ReadJSON(m)
			if err != nil {
				// server shutdown or socket closed
				if isClosed(err) {
					// reconnect
					log.Println("Trying reconnect ...")
					conn, err = connect(uri, header)
					if err != nil {
						log.Fatalln("Reconnect Failed...", err)
						return
					}
				}
				log.Println("Read WS message failed:", err)
				continue
			}
			log.Printf("Received Message %v From %v \n", m.Data, m.FromUserID)
			err = markMsgRead(*m, token)
			if err != nil {
				log.Println("MarkRead failed:", err)
			}
		}
	}()

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		str := strings.Split(scanner.Text(), " ")
		ToUserID, er := strconv.ParseInt(str[0], 10, 64)
		if er != nil {
			log.Println("cannot recognize ToUserId:", err)
			break
		}
		m := model.Message{
			FromUserID: u.ID,
			ToUserID:   ToUserID,
			Type:       "Text",
			Data:       str[1],
		}
		if err := conn.WriteJSON(m); err != nil {
			log.Println("Send WS message failed:", err)
		}
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

func connect(url string, header http.Header) (*websocket.Conn, error) {
	var (
		conn *websocket.Conn
		err  error
	)
	maxRetries := 5
	for attempt := 0; attempt < maxRetries; attempt++ {
		conn, _, err = websocket.DefaultDialer.Dial(url, header)
		if err == nil {
			log.Println("connected to server ...")
			return conn, nil
		}
		if attempt < maxRetries-1 {
			// 失败后的延迟
			time.Sleep(time.Second)
		}
	}
	return nil, err
}

func isClosed(err error) bool {
	_, ok := err.(*websocket.CloseError)
	return ok
}

func markMsgRead(m model.Message, token string) (err error) {
	var (
		bodyBytes []byte
		req       *http.Request
		resp      *http.Response
	)
	bodyBytes, err = json.Marshal(m)
	if err != nil {
		return err
	}
	// 创建一个请求体，这里使用的是JSON格式的数据
	body := bytes.NewBufferString(string(bodyBytes))
	req, err = http.NewRequest("POST", "http://localhost:8848/markRead", body)
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-TOKEN", token)
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
	log.Printf("markRead resp: %s \n", string(bodyBytes))
	return
}
