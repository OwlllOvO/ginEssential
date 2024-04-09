package main

import (
	"owlllovo/ginessential/controller"
	"owlllovo/ginessential/middleware"

	"github.com/gin-gonic/gin"
)

func CollectRoute(r *gin.Engine) *gin.Engine {
	r.Use(middleware.CORSMiddleware(), middleware.RecoveryMiddleware())
	r.POST("/api/auth/register", controller.Register)
	r.POST("/api/auth/login", controller.Login)
	r.GET("/api/auth/info", middleware.AuthMiddleware(), controller.Info)

	categoryRoutes := r.Group("/categories")
	categoryController := controller.NewCategoryController()
	categoryRoutes.POST("", categoryController.Create)
	categoryRoutes.PUT("/:id", categoryController.Update)
	categoryRoutes.GET("/:id", categoryController.Show)
	categoryRoutes.DELETE("/:id", categoryController.Delete)
	categoryRoutes.GET("", categoryController.ListAll)

	postRoutes := r.Group("/posts")
	postRoutes.Use(middleware.AuthMiddleware())
	postController := controller.NewPostController()
	postRoutes.POST("", postController.Create)
	postRoutes.PUT("/:id", postController.Update)
	postRoutes.GET("/:id", postController.Show)
	postRoutes.DELETE("/:id", postController.Delete)
	postRoutes.POST("/page/list", postController.PageList)
	postRoutes.POST("/upload", postController.UploadImage)
	postRoutes.POST("/:id/like", postController.LikePost)
	postRoutes.POST("/:id/unlike", postController.UnlikePost)
	postRoutes.GET("/:id/isliked", postController.IsLiked)
	postRoutes.GET("/rank", postController.LikeRank)

	r.GET("/user/:id", postController.GetUserPosts)

	// 添加评论相关的路由
	CommentController := controller.NewCommentController()
	postRoutes.POST("/:id/comments", CommentController.AddComment) // 添加评论
	postRoutes.GET("/:id/comments", CommentController.GetComments) // 获取特定图书的所有评论

	adminRoutes := r.Group("/admin")
	adminRoutes.Use(middleware.AdminAuthMiddleware())
	adminRoutes.POST("/users", controller.Register)         // 创建用户
	adminRoutes.PUT("/users/:id", controller.UpdateUser)    // 修改用户
	adminRoutes.DELETE("/users/:id", controller.DeleteUser) // 删除用户
	adminRoutes.GET("/users", controller.UserList)          // 用户列表
	adminRoutes.GET("/users/:id", controller.GetUser)       // 单个用户
	adminRoutes.POST("/posts/:id/approve", postController.ApprovePost)

	chatController := controller.NewChatController()
	r.POST("/message", middleware.AuthMiddleware(), chatController.SendMessage)
	r.GET("/messages", middleware.AuthMiddleware(), chatController.GetMessages)
	r.GET("/chatlist", middleware.AuthMiddleware(), chatController.ChatList)

	return r
}
