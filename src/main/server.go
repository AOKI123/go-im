package main

import (
	"errors"
	"log"
	"net/http"

	"aoki.com/go-im/src/conn"
	"aoki.com/go-im/src/model"
	"aoki.com/go-im/src/repo"
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
	// 将用户的未读消息发回去
	log.Printf("%s connected... \n", user.Name)
	defer OnDisconnect(*user)
	err = Sender.SendUnRead(user.ID, conn)
	if err != nil {
		log.Printf("Send Unread to %s failed... \n", user.Name, err)
	}
	for {
		v := &model.Message{}
		err = conn.ReadJSON(v)
		if err != nil {
			log.Println("Read Json Message Failed", err)
		}
		err = Sender.Send(v)
		if err != nil {
			log.Println("Send Message Failed", err)
		}
	}
}

func markRead(ctx *gin.Context) {
	user := model.ResolveUser(ctx.Request)
	if user == nil {
		ctx.AbortWithError(http.StatusUnauthorized,
			errors.New("ResolveUser failed..."))
		return
	}
	var message model.Message
	// 将请求的JSON绑定到user结构体中
	if err := ctx.ShouldBindJSON(&message); err != nil {
		ctx.AbortWithError(http.StatusBadRequest, err)
		return
	}
	repo.MessageRepoOps.MarkRead(&message)
	ctx.JSON(http.StatusOK, gin.H{"message": "mark read success"})
}

func OnDisconnect(user model.User) {
	log.Printf("%v disconnect...\n", user.Name)
	ConnManager.DelConn(user.ID)
}

func main() {
	server := gin.Default()
	server.GET("/ws", ws).POST("/markRead", markRead)
	server.Run("localhost:8848")
}
