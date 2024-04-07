package controller

import (
	"errors"
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

	postID, err := uuid.FromString(req.PostID)
	if err != nil {
		response.Fail(ctx, nil, "Invalid post ID")
		return
	}

	// 查找是否已存在相同收发者和帖子的 Chat
	var chat model.Chat
	err = c.DB.Where("((sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)) AND post_id = ?",
		sender.(model.User).ID, req.ReceiverID, req.ReceiverID, sender.(model.User).ID, postID).First(&chat).Error

	// 如果找不到，创建新的 Chat
	if errors.Is(err, gorm.ErrRecordNotFound) {
		chat = model.Chat{
			SenderID:   sender.(model.User).ID,
			ReceiverID: req.ReceiverID,
			PostID:     postID,
		}
		if err := c.DB.Create(&chat).Error; err != nil {
			response.Fail(ctx, nil, "Failed to create chat")
			return
		}
	} else if err != nil {
		response.Fail(ctx, nil, "Failed to retrieve chat")
		return
	}

	// 创建新的 Message 并关联到找到或创建的 Chat，同时确保 SenderID 被正确设置
	message := model.Message{
		ChatID:   chat.ID,
		SenderID: sender.(model.User).ID, // 明确设置 SenderID
		Content:  req.Content,
	}

	if err := c.DB.Create(&message).Error; err != nil {
		response.Fail(ctx, nil, "Failed to send message")
		return
	}

	response.Success(ctx, nil, "Message sent successfully")
}

func (c *ChatController) GetMessages(ctx *gin.Context) {
	user, _ := ctx.Get("user")
	receiverIDStr := ctx.Query("receiver_id")
	postIDStr := ctx.Query("post_id")

	receiverID, err := strconv.ParseUint(receiverIDStr, 10, 32)
	if err != nil {
		response.Fail(ctx, nil, "Invalid receiver ID")
		return
	}

	postID, err := uuid.FromString(postIDStr)
	if err != nil {
		response.Fail(ctx, nil, "Invalid post ID")
		return
	}

	// 查找对应的 Chat
	var chat model.Chat
	if err := c.DB.Where("((sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)) AND post_id = ?",
		user.(model.User).ID, receiverID, receiverID, user.(model.User).ID, postID).First(&chat).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Fail(ctx, nil, "Chat not found")
		} else {
			response.Fail(ctx, nil, "Failed to retrieve chat")
		}
		return
	}

	// 获取该 Chat 下的所有 Message
	var messages []model.Message
	if err := c.DB.Where("chat_id = ?", chat.ID).Order("created_at DESC").Find(&messages).Error; err != nil {
		response.Fail(ctx, nil, "Failed to retrieve messages")
		return
	}

	response.Success(ctx, gin.H{"messages": messages}, "Messages retrieved successfully")
}

func (c *ChatController) ChatList(ctx *gin.Context) {
	user, _ := ctx.Get("user")
	userID := user.(model.User).ID

	var conversations []struct {
		ChatID               uuid.UUID  `json:"chat_id"`
		PostID               uuid.UUID  `json:"post_id"`
		PostTitle            string     `json:"post_title"`
		PostHeadImg          string     `json:"post_head_img"`
		OtherParticipantID   uint       `json:"other_participant_id"`
		OtherParticipantName string     `json:"other_participant_name"`
		LastMessageContent   string     `json:"last_message_content"`
		LastMessageTime      model.Time `json:"last_message_time"`
	}

	if err := c.DB.Raw(`
		SELECT
			chats.id AS chat_id,
			chats.post_id,
			posts.title AS post_title,
			posts.head_img AS post_head_img,
			IF(chats.sender_id = ?, chats.receiver_id, chats.sender_id) AS other_participant_id,
			IF(chats.sender_id = ?, users_receiver.name, users_sender.name) AS other_participant_name,
			latest_messages.content AS last_message_content,
			latest_messages.created_at AS last_message_time
		FROM
			chats
		JOIN (
			SELECT
				chat_id,
				MAX(created_at) AS max_created_at
			FROM
				messages
			GROUP BY
				chat_id
		) AS latest_chat_messages ON chats.id = latest_chat_messages.chat_id
		JOIN messages AS latest_messages ON latest_chat_messages.chat_id = latest_messages.chat_id AND latest_chat_messages.max_created_at = latest_messages.created_at
		JOIN posts ON chats.post_id = posts.id
		LEFT JOIN users AS users_sender ON users_sender.id = chats.sender_id
		LEFT JOIN users AS users_receiver ON users_receiver.id = chats.receiver_id
		WHERE
			chats.sender_id = ? OR chats.receiver_id = ?
		ORDER BY
			latest_messages.created_at DESC
	`, userID, userID, userID, userID).Scan(&conversations).Error; err != nil {
		response.Fail(ctx, nil, "Failed to retrieve chat list")
		return
	}

	response.Success(ctx, gin.H{"conversations": conversations}, "Conversations retrieved successfully")
}
