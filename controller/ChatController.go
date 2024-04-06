package controller

import (
	"owlllovo/ginessential/common"
	"owlllovo/ginessential/model"
	"owlllovo/ginessential/response"
	"strconv"

	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"
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
		PostID     string `json:"post_id"`
		Content    string `json:"content"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Fail(ctx, nil, "Invalid request")
		return
	}

	postID, err := uuid.FromString(req.PostID) // 将字符串格式的 PostID 转换为 uuid.UUID 类型
	if err != nil {
		response.Fail(ctx, nil, "Invalid post ID")
		return
	}
	message := model.Message{
		SenderID:   sender.(model.User).ID,
		ReceiverID: req.ReceiverID,
		PostID:     postID,
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
	postIDStr := ctx.Query("post_id")
	receiverID, err := strconv.ParseUint(receiverIDStr, 10, 32)
	if err != nil {
		response.Fail(ctx, nil, "Invalid receiver ID")
		return
	}

	postID, err := uuid.FromString(postIDStr) // 将字符串格式的 PostID 转换为 uuid.UUID 类型
	if err != nil {
		response.Fail(ctx, nil, "Invalid post ID")
		return
	}

	var messages []model.Message
	if err := c.DB.Where("(sender_id = ? AND receiver_id = ? AND post_id = ?) OR (sender_id = ? AND receiver_id = ? AND post_id = ?)",
		sender.(model.User).ID, receiverID, postID, receiverID, sender.(model.User).ID, postID).
		Order("created_at DESC").Find(&messages).Error; err != nil {
		response.Fail(ctx, nil, "Failed to retrieve messages")
		return
	}

	response.Success(ctx, gin.H{"messages": messages}, "Messages retrieved successfully")
}

func (c *ChatController) ChatList(ctx *gin.Context) {
	user, _ := ctx.Get("user")
	userID := user.(model.User).ID

	var conversations []struct {
		SenderID           uint       `json:"sender_id"`
		SenderName         string     `json:"sender_name"`
		PostID             uuid.UUID  `json:"post_id"`
		PostTitle          string     `json:"post_title"`
		LastMessageContent string     `json:"last_message_content"`
		LastMessageTime    model.Time `json:"last_message_time"`
		PostHeadImg        string     `json:"post_head_img"` // 添加封面图字段
	}

	// 查询所有与当前用户有过往来的会话，及其最新的消息，包括帖子的封面图
	if err := c.DB.Raw(`
        SELECT m.sender_id, u.name AS sender_name, m.post_id, p.title AS post_title, 
               m.content AS last_message_content, m.created_at AS last_message_time, 
               p.head_img AS post_head_img
        FROM messages m
        JOIN users u ON m.sender_id = u.id
        JOIN posts p ON m.post_id = p.id
        INNER JOIN (
            SELECT sender_id, post_id, MAX(created_at) AS max_time
            FROM messages
            WHERE receiver_id = ? OR sender_id = ?
            GROUP BY sender_id, post_id
        ) AS mm ON m.sender_id = mm.sender_id AND m.post_id = mm.post_id AND m.created_at = mm.max_time
        ORDER BY m.created_at DESC
    `, userID, userID).Scan(&conversations).Error; err != nil {
		response.Fail(ctx, nil, "Failed to retrieve conversations")
		return
	}

	response.Success(ctx, gin.H{"conversations": conversations}, "Conversations retrieved successfully")
}
