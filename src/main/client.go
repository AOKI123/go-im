package main

import (
	"bufio"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

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
	header := http.Header{
		"X-TOKEN": []string{u.GenToken()},
	}
	conn, _, err := websocket.DefaultDialer.Dial(uri, header)
	if err != nil {
		log.Fatal("dial failed:", err)
	}
	defer conn.Close()

	go func() {
		for {
			m := &model.Message{}
			err = conn.ReadJSON(m)
			if err != nil {
				log.Println("Read WS message failed:", err)
				return
			}
			log.Printf("Received Message %v From %v \n", m.Data, m.FromUserID)
			// _, msg, err := conn.ReadMessage()
			// if err != nil {
			// 	log.Println("Read WS message failed:", err)
			// 	break
			// }
			// log.Printf("Received Message %v \n", string(msg))
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
