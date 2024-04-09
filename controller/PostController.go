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
	UploadImage(ctx *gin.Context)
	ApprovePost(ctx *gin.Context)
	LikePost(ctx *gin.Context)
	UnlikePost(ctx *gin.Context)
	IsLiked(ctx *gin.Context)
	LikeRank(ctx *gin.Context)
	GetUserPosts(ctx *gin.Context)
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
		err := p.DB.Where("name = ?", "GPT-4").FirstOrCreate(&aiUser, model.User{Name: "GPT-4", Role: "AI"}).Error
		if err != nil {
			log.Printf("Failed to ensure AI user exists: %v", err)
			return
		}

		aiComment, err := GetGPTComment(requestPost.HeadImg, "请对这幅儿童绘画作品给出评价，从作品内容、构图、技巧等方面进行评价，给出不足之处并提出改进建议")

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

func NewPostController() IPostController {
	db := common.GetDB()
	db.AutoMigrate(&model.Post{}, &model.Like{})
	return PostController{DB: db}
}

func (p PostController) GetUserPosts(ctx *gin.Context) {
	var userId string = ctx.Param("id")
	var pageNum int
	var pageSize int
	var err error

	if pageNum, err = strconv.Atoi(ctx.DefaultQuery("pageNum", "1")); err != nil {
		pageNum = 1
	}

	if pageSize, err = strconv.Atoi(ctx.DefaultQuery("pageSize", "10")); err != nil {
		pageSize = 10
	}

	var posts []model.Post
	if err := p.DB.
		Preload("Category").
		Preload("User").
		Where("user_id = ?", userId).
		Offset((pageNum - 1) * pageSize).
		Limit(pageSize).
		Find(&posts).
		Error; err != nil {
		response.Fail(ctx, nil, "Posts not found")
		return
	}
	var user model.User
	if err := p.DB.Where("id = ?", userId).First(&user).Error; err != nil {
		response.Fail(ctx, nil, "User not found")
		return
	}

	var total int64
	p.DB.Model(&model.Post{}).Where("user_id = ?", userId).Count(&total)

	response.Success(ctx, gin.H{"userName": user.Name, "posts": posts, "total": total, "pageNum": pageNum, "pageSize": pageSize}, "User's posts retrieved successfully")
}
