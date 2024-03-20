package controller

import (
	"log"
	"owlllovo/ginessential/model"
	"owlllovo/ginessential/repository"
	"owlllovo/ginessential/response"
	"owlllovo/ginessential/vo"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ICategoryController interface {
	RestController
	ListAll(ctx *gin.Context)
}

type CategoryController struct {
	Repository repository.CategoryRepository
}

func NewCategoryController() ICategoryController {
	repository := repository.NewCategoryRepository()
	repository.DB.AutoMigrate(model.Category{})

	return CategoryController{Repository: repository}
}

func (c CategoryController) Create(ctx *gin.Context) {
	var requestCategory vo.CreateCategoryRequest

	if err := ctx.ShouldBind(&requestCategory); err != nil {
		log.Println(err.Error())
		// Tutorial Error, Original:
		// response.Fail(ctx, "Data Error, Please Fill Category Name", nil)
		response.Fail(ctx, gin.H{"error": "Data Error, Please Fill Category Name"}, "")
		return
	}

	category, err := c.Repository.Create(requestCategory.Name)
	if err != nil {
		panic(err)
		return
	}

	response.Success(ctx, gin.H{"category": category}, "")
}

func (c CategoryController) Update(ctx *gin.Context) {
	// Bind parameters in body

	var requestCategory vo.CreateCategoryRequest

	if err := ctx.ShouldBind(&requestCategory); err != nil {
		// Tutorial Error, Original:
		// response.Fail(ctx, "Data Error, Please Fill Category Name", nil)
		response.Fail(ctx, gin.H{"error": "Data Error, Please Fill Category Name"}, "")
		return
	}

	// Get parameters in path
	categoryId, _ := strconv.Atoi(ctx.Params.ByName("id"))

	updateCategory, err := c.Repository.SelectById(categoryId)
	if err != nil {
		// Tutotial Error, Original:
		// response.Fail(ctx, "Category does not exist", nil)
		response.Fail(ctx, gin.H{"error": "Category does not exist"}, "")
		return
	}

	// Update Category
	category, err := c.Repository.Update(*updateCategory, requestCategory.Name)
	if err != nil {
		panic(err)
	}

	response.Success(ctx, gin.H{"category": category}, "Update Success")
}

func (c CategoryController) Show(ctx *gin.Context) {
	// Get parameters in path
	categoryId, _ := strconv.Atoi(ctx.Params.ByName("id"))

	category, err := c.Repository.SelectById(categoryId)
	if err != nil {
		// Tutotial Error, Original:
		// response.Fail(ctx, "Category does not exist", nil)
		response.Fail(ctx, gin.H{"error": "Category does not exist"}, "")
		return
	}

	response.Success(ctx, gin.H{"category": category}, "")
}

func (c CategoryController) Delete(ctx *gin.Context) {
	// Get parameters in path
	categoryId, _ := strconv.Atoi(ctx.Params.ByName("id"))

	if err := c.Repository.DeleteById(categoryId); err != nil {
		response.Fail(ctx, gin.H{"error": "Category does not exist"}, "")
		return
	}

	response.Success(ctx, nil, "")
}

func (c CategoryController) ListAll(ctx *gin.Context) {
	categories, err := c.Repository.ListAll()
	if err != nil {
		response.Fail(ctx, gin.H{"error": "Failed to retrieve categories"}, "")
		return
	}

	response.Success(ctx, gin.H{"categories": categories}, "")
}
