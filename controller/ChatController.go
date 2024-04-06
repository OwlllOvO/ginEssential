package controller

import (
	"owlllovo/ginessential/common"
	"owlllovo/ginessential/model"
	"owlllovo/ginessential/response"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ChatController struct {
	DB *gorm.DB
}

func NewChatController() *ChatController {
	return &ChatController{DB: common.GetDB()}
}

func (c *ChatController) SendMessage(ctx *gin.Context) {
	sender, _ := ctx.Get("user")
	var req struct {
		ReceiverID uint   `json:"receiver_id"`
		Content    string `json:"content"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Fail(ctx, nil, "Invalid request")
		return
	}

	message := model.Message{
		SenderID:   sender.(model.User).ID,
		ReceiverID: req.ReceiverID,
		Content:    req.Content,
	}

	if err := c.DB.Create(&message).Error; err != nil {
		response.Fail(ctx, nil, "Failed to send message")
		return
	}

	response.Success(ctx, nil, "Message sent successfully")
}

func (c *ChatController) GetMessages(ctx *gin.Context) {
	sender, _ := ctx.Get("user")

	// 从查询参数中获取receiver_id，而不是从请求体中获取
	receiverIDStr := ctx.Query("receiver_id")
	receiverID, err := strconv.ParseUint(receiverIDStr, 10, 32)
	if err != nil {
		response.Fail(ctx, nil, "Invalid receiver ID")
		return
	}

	var messages []model.Message
	if err := c.DB.Where("(sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)",
		sender.(model.User).ID, receiverID, receiverID, sender.(model.User).ID).
		Order("created_at DESC").Find(&messages).Error; err != nil {
		response.Fail(ctx, nil, "Failed to retrieve messages")
		return
	}

	response.Success(ctx, gin.H{"messages": messages}, "Messages retrieved successfully")
}
