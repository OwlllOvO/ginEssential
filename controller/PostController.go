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
	"gorm.io/gorm"
)

type IPostController interface {
	RestController
	PageList(ctx *gin.Context)
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
	// Get postId in path

	postId := ctx.Params.ByName("id")

	var post model.Post

	/* Tutorial Error, Original:
	if p.DB.Preload("Category").Where("id = ?", postId).First(&post).RecordNotFound() {
		response.Fail(ctx, gin.H{"error": "Post does not exist"}, "")
		return
	}
	*/

	result := p.DB.Preload("Category").Where("id = ?", postId).First(&post)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		response.Fail(ctx, gin.H{"error": "Post does not exist"}, "")
		return
	}

	response.Success(ctx, gin.H{"post": post}, "Show Success")
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
	db.AutoMigrate(model.Post{})
	return PostController{DB: db}
}
