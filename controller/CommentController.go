package controller

import (
	"owlllovo/ginessential/common"
	"owlllovo/ginessential/model"
	"owlllovo/ginessential/response"
	"owlllovo/ginessential/vo"

	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"
)

type ICommentController interface {
	AddComment(ctx *gin.Context)
	GetComments(ctx *gin.Context)
}

func NewCommentController() ICommentController {
	db := common.GetDB()
	db.AutoMigrate(&model.Comment{})
	return PostController{DB: db}
}

func (p PostController) AddComment(ctx *gin.Context) {
	// 获取postId字符串
	postIdStr := ctx.Params.ByName("id")
	// 将字符串转换为UUID
	postId, err := uuid.FromString(postIdStr)
	if err != nil {
		response.Fail(ctx, nil, "Invalid post ID")
		return
	}

	// 定义接收数据的结构体
	var commentVo vo.CreateCommentRequest
	// 绑定数据
	if err := ctx.ShouldBindJSON(&commentVo); err != nil {
		response.Fail(ctx, nil, "Data parsing error")
		return
	}
	user, _ := ctx.Get("user")
	userID := user.(model.User).ID
	// 创建评论
	comment := model.Comment{
		PostID:  postId, // 使用转换后的UUID
		UserID:  userID, // 设置评论者的UserID
		Content: commentVo.Content,
	}

	if err := p.DB.Create(&comment).Error; err != nil {
		response.Fail(ctx, nil, "Failed to add comment")
		return
	}

	response.Success(ctx, nil, "Comment added successfully")
}

func (p PostController) GetComments(ctx *gin.Context) {
	postId := ctx.Params.ByName("id")

	var comments []model.Comment
	// 预加载User关联，以获取每条评论的用户信息
	if err := p.DB.Where("post_id = ?", postId).Preload("User").Find(&comments).Error; err != nil {
		response.Fail(ctx, nil, "Failed to retrieve comments")
		return
	}

	response.Success(ctx, gin.H{"comments": comments}, "Comments retrieved successfully")
}
