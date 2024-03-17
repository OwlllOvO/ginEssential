package controller

import (
	"errors"
	"owlllovo/ginessential/common"
	"owlllovo/ginessential/model"
	"owlllovo/ginessential/response"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ICategoryController interface {
	RestController
}

type CategoryController struct {
	DB *gorm.DB
}

func NewCategoryController() ICategoryController {
	db := common.GetDB()
	db.AutoMigrate(model.Category{})

	return CategoryController{DB: db}
}

func (c CategoryController) Create(ctx *gin.Context) {
	var requestCategory model.Category
	ctx.Bind(&requestCategory)

	if requestCategory.Name == "" {
		// Tutorial Error, Original:
		// response.Fail(ctx, "Data Error, Please Fill Category Name", nil)
		response.Fail(ctx, gin.H{"error": "Data Error, Please Fill Category Name"}, "")
	}

	c.DB.Create(&requestCategory)

	response.Success(ctx, gin.H{"category": requestCategory}, "")
}

func (c CategoryController) Update(ctx *gin.Context) {
	// Bind parameters in body

	var requestCategory model.Category
	ctx.Bind(&requestCategory)

	if requestCategory.Name == "" {
		// Tutorial Error, Original:
		// response.Fail(ctx, "Data Error, Please Fill Category Name", nil)
		response.Fail(ctx, gin.H{"error": "Data Error, Please Fill Category Name"}, "")
	}

	// Get parameters in path
	categoryId, _ := strconv.Atoi(ctx.Params.ByName("id"))

	var updateCategory model.Category

	/* Tutorial Error, Original:
	if c.DB.First(&updateCategory, categoryId).RecordNotFound() {
		// Tutorial Error, Original:
		// response.Fail(ctx, "Category does not exist", nil)
		response.Fail(ctx, gin.H{"error": "Category does not exist"}, "")
	}
	*/

	result := c.DB.First(&updateCategory, categoryId)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		// Record not found, handle the error.
		response.Fail(ctx, gin.H{"error": "Category does not exist"}, "")
	} else if result.Error != nil {
		// Some other error occurred.
		response.Fail(ctx, gin.H{"error": "An unexpected error occurred"}, "")
	}
	// Update Category
	c.DB.Model(&updateCategory).Update("name", requestCategory.Name)

	response.Success(ctx, gin.H{"category": updateCategory}, "Update Success")
}

func (c CategoryController) Show(ctx *gin.Context) {
	// Get parameters in path
	categoryId, _ := strconv.Atoi(ctx.Params.ByName("id"))

	var category model.Category

	/* Tutorial Error, Original:
	if c.DB.First(&updateCategory, categoryId).RecordNotFound() {
		// Tutorial Error, Original:
		// response.Fail(ctx, "Category does not exist", nil)
		response.Fail(ctx, gin.H{"error": "Category does not exist"}, "")
	}
	*/

	result := c.DB.First(&category, categoryId)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		// Record not found, handle the error.
		response.Fail(ctx, gin.H{"error": "Category does not exist"}, "")
	} else if result.Error != nil {
		// Some other error occurred.
		response.Fail(ctx, gin.H{"error": "An unexpected error occurred"}, "")
	}

	response.Success(ctx, gin.H{"category": category}, "")
}

func (c CategoryController) Delete(ctx *gin.Context) {
	// Get parameters in path
	categoryId, _ := strconv.Atoi(ctx.Params.ByName("id"))

	if err := c.DB.Delete(model.Category{}, categoryId).Error; err != nil {
		response.Fail(ctx, gin.H{"error": "Category does not exist"}, "")
		return
	}

	response.Success(ctx, nil, "")
}
