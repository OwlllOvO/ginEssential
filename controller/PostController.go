package controller

import (
	"errors"
	"log"
	"owlllovo/ginessential/common"
	"owlllovo/ginessential/model"
	"owlllovo/ginessential/response"
	"owlllovo/ginessential/vo"
	"strconv"

	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"
	"gorm.io/gorm"
)

type IPostController interface {
	RestController
	PageList(ctx *gin.Context)
	AddComment(ctx *gin.Context)  // 添加评论的方法
	GetComments(ctx *gin.Context) // 获取评论的方法
}

type PostController struct {
	DB *gorm.DB
}

func (p PostController) Create(ctx *gin.Context) {
	var requestPost vo.CreatePostRequest

	// validate data

	if err := ctx.ShouldBind(&requestPost); err != nil {
		log.Println(err.Error())
		// Tutorial Error, Original:
		// response.Fail(ctx, "Data Error, Please Fill Category Name", nil)
		response.Fail(ctx, gin.H{"error": "Data Error, Please Fill Category Name"}, "")
		return
	}

	// Get logined user

	user, _ := ctx.Get("user")

	// Create Post

	post := model.Post{
		UserId:     user.(model.User).ID,
		CategoryId: requestPost.CategoryId,
		Title:      requestPost.Title,
		HeadImg:    requestPost.HeadImg,
		Content:    requestPost.Content,
	}

	if err := p.DB.Create(&post).Error; err != nil {
		panic(err)
		return
	}

	response.Success(ctx, nil, "Create Success")
}

func (p PostController) Update(ctx *gin.Context) {
	var requestPost vo.CreatePostRequest

	// validate data

	if err := ctx.ShouldBind(&requestPost); err != nil {
		// Tutorial Error, Original:
		// response.Fail(ctx, "Data Error, Please Fill Category Name", nil)
		response.Fail(ctx, gin.H{"error": "Data Error, Please Fill Category Name"}, "")
		return
	}

	// Get postId in path

	postId := ctx.Params.ByName("id")

	/* Tutorial Error, Original:
	var post model.Post
	if p.DB.Where("id = ?", postId).First(&post).RecordNotFound() {
		response.Fail(ctx, gin.H{"error": "Post does not exist"}, "")
		return
	}
	*/

	var post model.Post
	result := p.DB.Preload("Category").Where("id = ?", postId).First(&post)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		response.Fail(ctx, gin.H{"error": "Post does not exist"}, "")
		return
	} else if result.Error != nil {
		// Handle other potential errors
		response.Fail(ctx, gin.H{"error": "An error occurred"}, "")
		return
	}

	// Chech if user is author of post

	// Get logined user

	user, _ := ctx.Get("user")

	userId := user.(model.User).ID
	if userId != post.UserId {
		response.Fail(ctx, gin.H{"error": "Post does not belong to you, access denied"}, "")
		return
	}

	// Update post
	/* Tutorial Error, Original:
	if err := p.DB.Model(&post).Update(requestPost).Error; err != nil {
	*/
	if err := p.DB.Model(&post).Updates(requestPost).Error; err != nil {
		response.Fail(ctx, gin.H{"error": "Update Failed"}, "")
		return
	}

	response.Success(ctx, gin.H{"post": post}, "Update Success")

}

func (p PostController) Show(ctx *gin.Context) {
	postId := ctx.Params.ByName("id")

	var post model.Post

	// 使用Preload嵌套加载关联的评论以及评论的用户信息
	result := p.DB.Preload("Comments.User").Where("id = ?", postId).First(&post)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		response.Fail(ctx, gin.H{"error": "Post does not exist"}, "")
		return
	} else if result.Error != nil {
		log.Println(result.Error)
		response.Fail(ctx, gin.H{"error": "An error occurred while retrieving the post"}, "")
		return
	}

	response.Success(ctx, gin.H{"post": post}, "Show Success")
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

func (p PostController) Delete(ctx *gin.Context) {
	// Get postId in path

	postId := ctx.Params.ByName("id")

	var post model.Post

	/* Tutorial Error, Original:
	if p.DB.Where("id = ?", postId).First(&post).RecordNotFound() {
		response.Fail(ctx, gin.H{"error": "Post does not exist"}, "")
		return
	}
	*/

	result := p.DB.Where("id = ?", postId).First(&post)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		response.Fail(ctx, gin.H{"error": "Post does not exist"}, "")
		return
	} else if result.Error != nil {
		// Handle other possible errors
		response.Fail(ctx, gin.H{"error": "An unexpected error occurred"}, "")
		return
	}

	// Chech if user is author of post

	// Get logined user

	user, _ := ctx.Get("user")

	userId := user.(model.User).ID
	if userId != post.UserId {
		response.Fail(ctx, gin.H{"error": "Post does not belong to you, access denied"}, "")
		return
	}

	// Delete Post

	p.DB.Delete(&post)
	response.Success(ctx, gin.H{"post": post}, "Delete Success")
}

func (p PostController) PageList(ctx *gin.Context) {
	// Get page parameters
	pageNum, _ := strconv.Atoi(ctx.DefaultQuery("pageNum", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("pageSize", "20"))

	// Split into pages

	var posts []model.Post
	p.DB.Order("created_at desc").Offset((pageNum - 1) * pageSize).Limit(pageSize).Find(&posts)

	// Total numbers

	var total int64
	p.DB.Model(model.Post{}).Count(&total)

	response.Success(ctx, gin.H{"data": posts, "total": total}, "Success")
}

func NewPostController() IPostController {
	db := common.GetDB()
	db.AutoMigrate(&model.Post{}, &model.Comment{})
	return PostController{DB: db}
}
