package controller

import (
	"errors"
	"log"
	"math"
	"net/http"
	"owlllovo/ginessential/common"
	"owlllovo/ginessential/dto"
	"owlllovo/ginessential/model"
	"owlllovo/ginessential/response"
	"owlllovo/ginessential/util"
	"strconv"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func Register(ctx *gin.Context) {
	DB := common.GetDB()

	// get parameter via map
	// var requestMap = make(map[string]string)
	// json.NewDecoder(ctx.Request.Body).Decode(&requestMap)

	// get parameter via struct and gin-bind
	var requestUser = model.User{}
	// json.NewDecoder(ctx.Request.Body).Decode(&requestUser)
	ctx.Bind(&requestUser)

	name := requestUser.Name
	telephone := requestUser.Telephone
	password := requestUser.Password
	role := requestUser.Role

	// data verify

	if len(telephone) != 11 {
		response.Response(ctx, http.StatusUnprocessableEntity, 422, nil, "Phone number should be 11 digits")
		return
	}
	if len(password) < 6 {
		response.Response(ctx, http.StatusUnprocessableEntity, 422, nil, "Password too weak")
		return
	}
	if len(name) == 0 {
		name = util.RandomString(10)
	}
	if len(role) == 0 {
		role = "user"
	}

	log.Println(name, telephone, password, role)

	// check phone number

	if isTelephoneExist(DB, telephone) {
		response.Response(ctx, http.StatusUnprocessableEntity, 422, nil, "User exist")
		return
	}

	// create user

	hasedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		response.Response(ctx, http.StatusUnprocessableEntity, 500, nil, "Encryption error")
		return
	}
	newUser := model.User{
		Name:      name,
		Telephone: telephone,
		Password:  string(hasedPassword),
		Role:      role,
	}
	DB.Create(&newUser)

	// send token

	token, err := common.ReleaseToken(newUser)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "System Error"})
		log.Printf("token generate error: %v", err)
		return
	}

	// return result
	response.Success(ctx, gin.H{"token": token}, "Register Success")
}

func Login(ctx *gin.Context) {

	DB := common.GetDB()

	// get parameter via struct and gin-bind
	var requestUser = model.User{}
	// json.NewDecoder(ctx.Request.Body).Decode(&requestUser)
	ctx.Bind(&requestUser)

	telephone := requestUser.Telephone
	password := requestUser.Password

	// data verify

	if len(telephone) != 11 {
		response.Response(ctx, http.StatusUnprocessableEntity, 422, nil, "Phone number should be 11 digits")
		return
	}
	if len(password) < 6 {
		response.Response(ctx, http.StatusUnprocessableEntity, 422, nil, "Password too weak")
		return
	}

	// check phone number exist

	var user model.User
	DB.Where("telephone = ?", telephone).First(&user)
	if user.ID == 0 {
		ctx.JSON(http.StatusUnprocessableEntity, gin.H{"code": 422, "msg": "User does not exist"})
		return

	}

	// check password correct

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "Password wrong"})
		return
	}

	// send token

	token, err := common.ReleaseToken(user)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "System Error"})
		log.Printf("token generate error: %v", err)
		return
	}

	// return result
	response.Success(ctx, gin.H{"token": token, "userid": user.ID}, "Login Success")
}

func Info(ctx *gin.Context) {
	user, _ := ctx.Get("user")

	ctx.JSON(http.StatusOK, gin.H{"code": 200, "data": gin.H{"user": dto.ToUserDto(user.(model.User))}})
}

func isTelephoneExist(db *gorm.DB, telephone string) bool {
	var user model.User
	db.Where("telephone = ?", telephone).First(&user)
	if user.ID != 0 {
		return true
	}
	return false
}

func UpdateUser(ctx *gin.Context) {
	DB := common.GetDB()
	userIdStr := ctx.Param("id")              // 假设 userId 是从 URL 参数中获取的字符串
	userIdInt, err := strconv.Atoi(userIdStr) // 尝试将字符串转换为整数
	if err != nil {
		// 转换失败，处理错误，例如返回错误响应
		response.Response(ctx, http.StatusBadRequest, 400, nil, "Invalid user ID format")
		return
	}
	var requestUser = model.User{}
	ctx.Bind(&requestUser)
	if len(requestUser.Telephone) != 11 {
		response.Response(ctx, http.StatusUnprocessableEntity, 422, nil, "Phone number should be 11 digits")
		return
	}
	if len(requestUser.Password) < 6 {
		response.Response(ctx, http.StatusUnprocessableEntity, 422, nil, "Password too weak")
		return
	}
	if len(requestUser.Name) == 0 {
		requestUser.Name = util.RandomString(10)
	}
	if len(requestUser.Role) == 0 {
		requestUser.Role = "user"
	}
	log.Println(requestUser.Name, requestUser.Telephone, requestUser.Password, requestUser.Role)

	if isNewTelephoneExist(DB, requestUser.Telephone, uint(userIdInt)) {
		response.Response(ctx, http.StatusUnprocessableEntity, 422, nil, "Telephone already in use by another user")
		return
	}

	hasedPassword, err := bcrypt.GenerateFromPassword([]byte(requestUser.Password), bcrypt.DefaultCost)
	if err != nil {
		response.Response(ctx, http.StatusUnprocessableEntity, 500, nil, "Encryption error")
		return
	}
	requestUser.Password = string(hasedPassword)
	var user model.User
	DB.First(&user, userIdStr)
	if user.ID == 0 {
		response.Response(ctx, http.StatusNotFound, 404, nil, "User not found")
		return
	}
	log.Println("LOG2:", user)
	// 更新用户信息
	DB.Model(&user).Updates(requestUser)

	response.Success(ctx, nil, "User updated successfully")
}

func isNewTelephoneExist(db *gorm.DB, telephone string, excludeUserId uint) bool {
	var user model.User
	db.Where("telephone = ? AND id <> ?", telephone, excludeUserId).First(&user)
	if user.ID != 0 {
		return true
	}
	return false
}

func DeleteUser(ctx *gin.Context) {
	DB := common.GetDB()
	userId := ctx.Param("id")

	var user model.User
	DB.First(&user, userId)
	if user.ID == 0 {
		response.Response(ctx, http.StatusNotFound, 404, nil, "User not found")
		return
	}

	DB.Delete(&user)

	response.Success(ctx, nil, "User deleted successfully")
}

func UserList(ctx *gin.Context) {
	DB := common.GetDB()

	// 获取分页参数
	pageNum, _ := strconv.Atoi(ctx.DefaultQuery("pageNum", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("pageSize", "10"))

	if pageNum <= 0 {
		pageNum = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	var users []model.User
	var total int64

	// 分页查询用户数据
	DB.Select("id", "name", "telephone", "role", "created_at", "updated_at").Order("created_at desc").Offset((pageNum - 1) * pageSize).Limit(pageSize).Find(&users)

	// 查询总用户数以计算总页数
	DB.Model(&model.User{}).Count(&total)

	// 计算总页数
	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))

	// 返回分页的用户数据和总页数
	ctx.JSON(http.StatusOK, gin.H{
		"code":       200,
		"data":       users,
		"total":      total,
		"totalPages": totalPages,
		"pageNum":    pageNum,
		"pageSize":   pageSize,
	})
}

func GetUser(ctx *gin.Context) {
	DB := common.GetDB()
	// 从URL路径中获取用户ID
	userId := ctx.Param("id")

	var user model.User
	// 根据ID查询用户，注意不要返回密码字段
	result := DB.Select("ID", "CreatedAt", "UpdatedAt", "Name", "Telephone", "Role").First(&user, userId)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "User does not exist"})
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "Database error"})
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"code": 200, "data": user})
}
