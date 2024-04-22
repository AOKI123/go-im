package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

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
var Sender = conn.Sender{}

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
			if isConnClosed(err) {
				break
			}
			log.Println("Read Json Message Failed", err)
		}
		err = Sender.Send(v)
		if err != nil {
			log.Println("Send Message Failed", err)
		}
	}
}

// 标记消息已读接口
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

// 推送消息接口
func send(ctx *gin.Context) {
	// TODO 签名校验, 确保只能server间调用
	var message model.Message
	// 将请求的JSON绑定到user结构体中
	if err := ctx.ShouldBindJSON(&message); err != nil {
		ctx.AbortWithError(http.StatusBadRequest, err)
		return
	}
	// 获取接收方的连接
	conn := ConnManager.FindConn(message.ToUserID)
	if conn == nil {
		msg := fmt.Sprintf("%d does not online", message.ToUserID)
		ctx.JSON(http.StatusOK, gin.H{"message": msg})
		return
	}
	// 推送消息
	err := conn.WriteJSON(message)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "send success"})
}

func OnDisconnect(user model.User) {
	log.Printf("%v disconnect...\n", user.Name)
	ConnManager.DelConn(user.ID)
}

func isConnClosed(err error) bool {
	_, ok := err.(*websocket.CloseError)
	return ok
}

func main() {
	args := os.Args[1:]
	port, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		log.Fatal("Parse ServerPort Failed:", err)
	}
	// 本地测试，当部署到生产环境时可以通过容器的环境变量获取IP地址
	model.CurrServer.Host = "localhost"
	model.CurrServer.Port = int(port)
	server := gin.Default()
	server.GET("/ws", ws).POST("/markRead", markRead).POST("/send", send)
	server.Run(model.CurrServer.ServerAddr())
}
