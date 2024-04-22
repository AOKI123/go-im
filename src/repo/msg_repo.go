package repo

import (
	"time"

	"aoki.com/go-im/src/model"
)

type MessageRepo struct {
}

var MessageRepoOps = MessageRepo{}

func (MessageRepo) Create(m *model.Message) {
	DatabaseOps.Create(m)
}

// 查询用户未读消息
func (MessageRepo) FindUnRead(userID int64) []*model.Message {
	messages := make([]*model.Message, 0)
	DatabaseOps.Where("to_user_id = ? AND `read` = 0", userID).Find(&messages)
	return messages
}

// 标记已读
func (MessageRepo) MarkRead(m *model.Message) {
	now := time.Now()
	DatabaseOps.Model(&m).Updates(model.Message{Read: true, ReadTime: &now})
}
