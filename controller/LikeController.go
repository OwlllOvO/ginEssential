package controller

import (
	"errors"
	"log"
	"owlllovo/ginessential/common"
	"owlllovo/ginessential/model"
	"owlllovo/ginessential/response"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ILikeController interface {
	RestController
	LikePost(ctx *gin.Context)
	UnlikePost(ctx *gin.Context)
	IsLiked(ctx *gin.Context)
	LikeRank(ctx *gin.Context)
}

func NewLikeController() ILikeController {
	db := common.GetDB()
	db.AutoMigrate(&model.Like{})
	return PostController{DB: db}
}

func (p PostController) LikePost(ctx *gin.Context) {
	user, _ := ctx.Get("user")
	userID := user.(model.User).ID
	postID := ctx.Param("id")
	log.Println(userID)
	var post model.Post

	// 检查帖子是否存在
	if err := p.DB.Where("id = ?", postID).First(&post).Error; err != nil {
		response.Fail(ctx, nil, "帖子不存在")
		return
	}
	var existingLike model.Like
	if err := p.DB.Where("user_id = ? AND post_id = ?", userID, postID).First(&existingLike).Error; err == nil {
		response.Fail(ctx, nil, "已经点过赞了")
		return
	}
	like := model.Like{UserId: userID, PostId: post.ID}
	p.DB.Create(&like)
	response.Success(ctx, nil, "点赞成功")
}

func (p PostController) UnlikePost(ctx *gin.Context) {
	user, _ := ctx.Get("user")
	userID := user.(model.User).ID // 获取当前登录用户的ID
	postID := ctx.Param("id")      // 从请求路径中获取帖子ID

	var like model.Like
	// 查找点赞记录，确保记录存在
	err := p.DB.Where("user_id = ? AND post_id = ?", userID, postID).First(&like).Error
	if err != nil {
		// 如果找不到记录，返回错误信息
		response.Fail(ctx, nil, "点赞记录不存在")
		return
	}

	// 删除找到的点赞记录
	if err := p.DB.Delete(&like).Error; err != nil {
		// 如果删除失败，返回错误信息
		response.Fail(ctx, nil, "取消点赞失败")
		return
	}

	// 成功取消点赞
	response.Success(ctx, nil, "取消点赞成功")
}

func (p PostController) IsLiked(ctx *gin.Context) {
	user, _ := ctx.Get("user")
	userID := user.(model.User).ID
	postID := ctx.Param("id")

	var like model.Like
	err := p.DB.Where("user_id = ? AND post_id = ?", userID, postID).First(&like).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		// 没有找到点赞记录，说明用户没有点赞过该帖子
		response.Success(ctx, gin.H{"isLiked": false}, "Post is not liked by the user")
	} else if err != nil {
		// 数据库查询出错
		response.Fail(ctx, gin.H{"error": err.Error()}, "Failed to check the like status")
	} else {
		// 找到点赞记录，说明用户已经点赞过该帖子
		response.Success(ctx, gin.H{"isLiked": true}, "Post is liked by the user")
	}
}

func (p PostController) LikeRank(ctx *gin.Context) {
	pageNum, _ := strconv.Atoi(ctx.DefaultQuery("pageNum", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("pageSize", "20"))

	// 修改结构体，加入排名字段
	type PostWithLikeCount struct {
		model.Post
		LikeCount int `json:"like_count"`
		Rank      int `json:"rank"` // 加入排名字段
	}

	var postsWithLikeCount []PostWithLikeCount

	// 查询帖子以及对应的点赞数量
	p.DB.Model(&model.Post{}).
		Preload("Category").
		Preload("User").
		Select("posts.*, COUNT(likes.id) as like_count").
		Joins("LEFT JOIN likes ON likes.post_id = posts.id").
		Group("posts.id").
		Order("like_count DESC, posts.created_at DESC").
		Offset((pageNum - 1) * pageSize).Limit(pageSize).
		Find(&postsWithLikeCount)

	// 设置排名
	for i, post := range postsWithLikeCount {
		// 排名从1开始，所以需要加1
		post.Rank = i + 1 + (pageNum-1)*pageSize
		postsWithLikeCount[i] = post // 更新切片中的元素
	}

	// 查询总帖子数，用于分页
	var total int64
	p.DB.Model(&model.Post{}).Count(&total)

	response.Success(ctx, gin.H{"data": postsWithLikeCount, "total": total}, "Success")
}
