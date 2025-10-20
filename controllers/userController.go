package controllers

import (
	"errors"
	"net/http"

	"github.com/antidote-kt/SSE_Library-back/dao"
	"github.com/antidote-kt/SSE_Library-back/dto"
	"github.com/antidote-kt/SSE_Library-back/models"
	"github.com/antidote-kt/SSE_Library-back/response"
	"github.com/antidote-kt/SSE_Library-back/utils"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt" // 引入密码加密库
	"gorm.io/gorm"
)

func RegisterUser(c *gin.Context) {
	var req dto.RegisterDTO // 定义一个用于绑定请求参数的结构体
	// 1. 绑定并验证请求参数
	if err := c.ShouldBind(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, nil, "参数绑定失败")
		// 如果绑定失败，调用response层返回错误响应
	}

	// 2.验证当前要注册的用户名是否已经存在
	_, err := dao.GetUserByUsername(req.Username)
	if err == nil {
		//如果返回为nil,说明数据查到了用户名，出现重复，注册失败
		response.Fail(c, http.StatusBadRequest, nil, "用户名已存在")
		return
	}

	// 如果 err 不是 nil，我们需要判断它到底是什么错误
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		// 如果错误不是 "记录未找到"，那它就是一个我们没预料到的数据库内部错误
		response.Fail(c, http.StatusInternalServerError, nil, "数据库查询失败")
		return
	}

	// 3. 检查邮箱是否已存在
	_, err = dao.GetUserByEmail(req.Email)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		response.Fail(c, http.StatusInternalServerError, nil, "数据库查询失败")
		return
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		response.Fail(c, http.StatusUnprocessableEntity, nil, "邮箱已被注册")
		return
	}

	var avatarURL string
	// 4. 上传用户头像（uploaderAvatar会检查是否为空）
	avatarURL, err = utils.UploadAvatar(req.Avatar)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, nil, err.Error())
		return
	}

	// 如果代码能执行到这里，说明 err 恰好是 gorm.ErrRecordNotFound，意味着用户名可用
	// 5. 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, nil, "密码加密失败")
		return
	}

	// 6. 创建用户模型
	newUser := models.User{
		Username: req.Username,
		Password: string(hashedPassword),
		Email:    req.Email,
		Avatar:   avatarURL,
		Role:     "user",   // 默认角色
		Status:   "active", // 默认状态
	}

	// 7. 保存到数据库
	createdUser, err := dao.CreateUser(newUser)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, nil, "用户注册失败")
		return
	}

	// 8. 返回成功响应
	createdUserDTO := buildUserBriefDTO(createdUser) //使用UserDTO中封装的用户简要信息模板，实现统一接口返回格式
	response.Success(c, gin.H{"user:": createdUserDTO}, "注册成功")
}

// Login 处理用户登录请求
func Login(c *gin.Context) {
	var req dto.LoginDTO

	// 1. 绑定并验证参数
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, nil, "参数格式错误: "+err.Error())
		return
	}

	// 2. 检查用户是否存在
	user, err := dao.GetUserByEmail(req.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Fail(c, http.StatusUnauthorized, nil, "用户邮箱错误")
			return
		}
		response.Fail(c, http.StatusInternalServerError, nil, "数据库查询失败")
		return
	}

	// 3. 校验密码
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		response.Fail(c, http.StatusUnauthorized, nil, "密码错误")
		return
	}

	// 4. 生成JWT
	token, err := utils.GenerateToken(user.ID, user.Username, user.Role)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, nil, "Token生成失败")
		return
	}

	// 5. 返回成功响应
	UserDTO := buildUserBriefDTO(user) //使用UserDTO中封装的用户简要信息模板，实现统一接口返回格式
	response.Success(c, gin.H{
		"token": token,
		"user":  UserDTO,
	}, "登录成功")
}
