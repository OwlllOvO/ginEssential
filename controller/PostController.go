package controller

import (
	"errors"
	"log"
	"net/http"
	"owlllovo/ginessential/common"
	"owlllovo/ginessential/model"
	"owlllovo/ginessential/response"
	"owlllovo/ginessential/vo"
	"path/filepath"
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
	UploadImage(ctx *gin.Context)
	ApprovePost(ctx *gin.Context)
	LikePost(ctx *gin.Context)
	UnlikePost(ctx *gin.Context)
}

type PostController struct {
	DB *gorm.DB
}

func (p PostController) Create(ctx *gin.Context) {
	var requestPost vo.CreatePostRequest
	// validate data

	if err := ctx.ShouldBind(&requestPost); err != nil {
		response.Fail(ctx, gin.H{"error": "Data Error, Please Fill Category Name"}, "")
	}
	log.Println(requestPost)
	var category model.Category
	// 尝试根据分类名称查找分类
	result := p.DB.Where("name = ?", requestPost.CategoryName).First(&category)

	// 如果分类不存在，返回错误信息
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		response.Fail(ctx, gin.H{"error": "Category does not exist"}, "")
		return
	}

	// 如果查找过程中发生了其他错误，也返回错误信息
	if result.Error != nil {
		log.Println(result.Error)
		response.Fail(ctx, gin.H{"error": "An error occurred while retrieving the category"}, "")
		return
	}

	// Get logined user

	user, _ := ctx.Get("user")

	// Create Post

	post := model.Post{
		UserId:     user.(model.User).ID,
		CategoryId: category.ID,
		Title:      requestPost.Title,
		HeadImg:    requestPost.HeadImg,
		Content:    requestPost.Content,
		Status:     "Pending",
	}

	if err := p.DB.Create(&post).Error; err != nil {
		panic(err)
		return
	}

	response.Success(ctx, nil, "Create Success")

	go func() {
		var aiUser model.User
		err := p.DB.Where("name = ?", "AI").FirstOrCreate(&aiUser, model.User{Name: "AI", Role: "AI"}).Error
		if err != nil {
			log.Printf("Failed to ensure AI user exists: %v", err)
			return
		}

		aiComment, err := GetAIComment(requestPost.HeadImg)
		if err != nil {
			log.Printf("Failed to get AI comment: %v", err)
		} else {
			// Assume the AI user has a fixed user ID, e.g., 1
			aiUserComment := model.Comment{
				PostID:  post.ID,
				UserID:  aiUser.ID, // AI用户的ID
				Content: aiComment,
			}

			if err := p.DB.Create(&aiUserComment).Error; err != nil {
				log.Printf("Failed to save AI comment: %v", err)
			}
		}
	}()
}

func (p PostController) Update(ctx *gin.Context) {
	var requestPost vo.CreatePostRequest

	// validate data

	if err := ctx.ShouldBind(&requestPost); err != nil {
		// Tutorial Error, Original:
		// response.Fail(ctx, "Data Error, Please Fill Category Name", nil)
		log.Println(requestPost)
		response.Fail(ctx, gin.H{"error": "Data Error, Please Fill Category Name"}, "")
		return
	}

	var category model.Category
	if err := p.DB.Where("name = ?", requestPost.CategoryName).First(&category).Error; err != nil {
		response.Fail(ctx, gin.H{"error": "Category does not exist"}, "")
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
	result := p.DB.Where("id = ?", postId).First(&post)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		response.Fail(ctx, gin.H{"error": "Post does not exist"}, "")
		return
	} else if result.Error != nil {
		// Handle other potential errors
		response.Fail(ctx, gin.H{"error": "An error occurred"}, "")
		return
	}

	// Check if user is author of post

	// Get logined user

	user, _ := ctx.Get("user")

	userId := user.(model.User).ID
	log.Println(user.(model.User).Role)
	if userId != post.UserId && user.(model.User).Role != "Admin" {
		response.Fail(ctx, gin.H{"error": "Only author and admin can edit posts"}, "")
		return
	}

	// Update post
	/* Tutorial Error, Original:
	if err := p.DB.Model(&post).Update(requestPost).Error; err != nil {
	*/
	if err := p.DB.Model(&model.Post{}).Where("id = ?", postId).Updates(model.Post{
		CategoryId: category.ID,
		Title:      requestPost.Title,
		HeadImg:    requestPost.HeadImg,
		Content:    requestPost.Content,
		Status:     requestPost.Status,
	}).Error; err != nil {
		response.Fail(ctx, gin.H{"error": "Update Failed"}, "")
		return
	}

	response.Success(ctx, gin.H{"post": post}, "Update Success")

}

func (p PostController) Show(ctx *gin.Context) {
	postId := ctx.Params.ByName("id")

	var likeCount int64
	p.DB.Model(&model.Like{}).Where("post_id = ?", postId).Count(&likeCount)
	var post model.Post

	// 使用Preload嵌套加载关联的评论以及评论的用户信息
	result := p.DB.Preload("Category").Preload("Comments.User").Preload("User").Where("id = ?", postId).First(&post)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		response.Fail(ctx, gin.H{"error": "Post does not exist"}, "")
		return
	} else if result.Error != nil {
		log.Println(result.Error)
		response.Fail(ctx, gin.H{"error": "An error occurred while retrieving the post"}, "")
		return
	}

	response.Success(ctx, gin.H{"post": post, "likeCount": likeCount}, "Show Success")
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
	if userId != post.UserId && user.(model.User).Role != "Admin" {
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
	p.DB.Preload("Category").Preload("User").Order("created_at desc").Offset((pageNum - 1) * pageSize).Limit(pageSize).Find(&posts)

	// Total numbers

	var total int64
	p.DB.Model(model.Post{}).Count(&total)

	response.Success(ctx, gin.H{"data": posts, "total": total}, "Success")
}

func (p PostController) UploadImage(ctx *gin.Context) {
	file, err := ctx.FormFile("file")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 生成唯一文件名
	ext := filepath.Ext(file.Filename)
	newFileName := uuid.NewV4().String() + ext

	// 保存文件
	path := "assets/images/" + newFileName
	if err := ctx.SaveUploadedFile(file, path); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 返回保存的文件名
	ctx.JSON(http.StatusOK, gin.H{"filename": newFileName})
}

// 在 PostController 中添加 ApprovePost 方法
func (p PostController) ApprovePost(ctx *gin.Context) {

	postId := ctx.Param("id") // 获取URL参数中的postId

	var post model.Post
	// 查找帖子，确保它存在
	if err := p.DB.Where("id = ?", postId).First(&post).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Fail(ctx, gin.H{"error": "Post not found"}, "Post does not exist")
			return
		}
		response.Fail(ctx, gin.H{"error": err.Error()}, "An error occurred while retrieving the post")
		return
	}

	// 更改帖子的状态为"Approved"
	post.Status = "Approved"
	if err := p.DB.Save(&post).Error; err != nil {
		response.Fail(ctx, gin.H{"error": err.Error()}, "Failed to approve the post")
		return
	}

	response.Success(ctx, gin.H{"post": post}, "Post approved successfully")
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

func NewPostController() IPostController {
	db := common.GetDB()
	db.AutoMigrate(&model.Post{}, &model.Comment{}, &model.Like{})
	return PostController{DB: db}
}
