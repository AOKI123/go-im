package main

import (
	"log"
	"net/http"

	"aoki.com/go-im/src/conn"
	"aoki.com/go-im/src/model"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	// Allow Cross Origin
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var ConnManager *conn.ConnManager = conn.GetConnManager()

var Sender = conn.LocalSender{}

func ws(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	user := model.ResolveUser(c.Request)
	if user == nil {
		log.Println("ResolveUser failed...")
		return
	}
	ConnManager.AddConn(user.ID, conn)
	log.Printf("%s connected... \n", user.Name)
	defer OnDisconnect(*user)
	for {
		v := &model.Message{}
		err = conn.ReadJSON(v)
		if err != nil {
			log.Println("Read Json Message Failed", err)
		}
		err = Sender.Send(*v)
		if err != nil {
			log.Println("Send Message Failed", err)
		}
	}
}

func OnDisconnect(user model.User) {
	log.Printf("%v disconnect...\n", user.Name)
	ConnManager.DelConn(user.ID)
}

func main() {
	server := gin.Default()
	server.GET("/ws", ws)
	server.Run("localhost:8848")
}
