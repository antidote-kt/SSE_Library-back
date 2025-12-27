package controllers

import (
	"errors"
	"net/http"

	"github.com/antidote-kt/SSE_Library-back/constant"
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
		// 如果绑定失败，调用response层返回错误响应
		response.Fail(c, http.StatusBadRequest, nil, constant.ParamParseError)
		return
	}

	// 2.验证当前要注册的用户名是否已经存在
	_, err := dao.GetUserByUsername(req.Username)
	if err == nil {
		//如果返回为nil,说明数据查到了用户名，出现重复，注册失败
		response.Fail(c, http.StatusBadRequest, nil, constant.UserNameAlreadyExist)
		return
	}

	// 如果 err 不是 nil，我们需要判断它到底是什么错误
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		// 如果错误不是 "记录未找到"，那它就是一个我们没预料到的数据库内部错误
		response.Fail(c, http.StatusInternalServerError, nil, constant.DatabaseError)
		return
	}

	// 如果代码能执行到这里，说明 err 恰好是 gorm.ErrRecordNotFound，意味着用户名可用
	var avatarURL string
	// 3. 上传用户头像（uploaderAvatar会检查是否为空）
	avatarURL, err = utils.UploadAvatar(req.Avatar)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, nil, err.Error())
		return
	}

	// 4.验证邮箱验证码（如果邮箱已注册用户，验证码就无法发出，前端也就无法正常调用此接口，因此我们无需再次额外检查邮箱是否已被注册）
	isValidCode, err := utils.CheckVerificationCode(req.Email, "register", req.VerificationCode)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, nil, constant.VerificationCodeCheckError)
		return
	}
	if !isValidCode {
		response.Fail(c, http.StatusBadRequest, nil, constant.VerificationCodeExpired)
		return
	}

	// 5. 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, nil, constant.PasswordEncryptFailed)
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
		response.Fail(c, http.StatusInternalServerError, nil, constant.UserRegisterFailed)
		return
	}

	// 8.调用response层的结构体组装返回数据
	userBrief := response.UserBriefResponse{
		UserID:     createdUser.ID,
		Username:   createdUser.Username,
		UserAvatar: createdUser.Avatar,
		Status:     createdUser.Status,
		CreateTime: createdUser.CreatedAt.Format("2006-01-02 15:04:05"),
		Email:      createdUser.Email,
		Role:       createdUser.Role,
	}

	// 9. 返回响应
	response.SuccessWithData(c, userBrief, constant.UserRegisterSuccess)
}

// Login 处理用户登录请求
func Login(c *gin.Context) {
	var req dto.LoginDTO

	// 1. 绑定并验证参数
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, nil, constant.ParamParseError)
		return
	}

	// 2. 检查用户是否存在
	user, err := dao.GetUserByEmail(req.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Fail(c, http.StatusUnauthorized, nil, constant.UnauthorizedEmail)
			return
		}
		response.Fail(c, http.StatusInternalServerError, nil, constant.DatabaseError)
		return
	}

	// 3. 校验密码
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		response.Fail(c, http.StatusUnauthorized, nil, constant.PasswordFalse)
		return
	}

	// 4. 生成JWT
	token, err := utils.GenerateToken(user.ID, user.Username, user.Role)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, nil, constant.TokenGenerateFailed)
		return
	}

	// 5.调用response层的结构体组装返回数据
	userBrief := response.UserBriefResponse{
		UserID:     user.ID,
		Username:   user.Username,
		UserAvatar: user.Avatar,
		Status:     user.Status,
		CreateTime: user.CreatedAt.Format("2006-01-02 15:04:05"),
		Email:      user.Email,
		Role:       user.Role,
	}

	// 6. 构造并返回响应信息
	response.Success(c, gin.H{
		"token": token,
		"user":  userBrief,
	}, constant.UserLoginSuccess)
}

// ChangePassword 处理修改密码请求
func ChangePassword(c *gin.Context) {
	var req dto.ChangePasswordDTO

	// 1. 绑定并验证参数
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, nil, constant.ParamParseError)
		return
	}

	// 2. 根据当前邮箱查找用户，用于后续更新用户密码到数据库中
	// 如果邮箱未注册用户，验证码就无法发出，也就无法正常调用此接口，因此这里我们无需再次用err检查邮箱是否已被注册
	// 但是我们需要检查用户状态，是否已停用
	user, _ := dao.GetUserByEmail(req.Email)
	if user.Status == "disabled" {
		response.Fail(c, http.StatusBadRequest, nil, constant.UserBeenSuspended)
		return
	}

	// 3. 验证邮箱验证码
	isValidCode, err := utils.CheckVerificationCode(req.Email, "reset-password", req.VerificationCode)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, nil, constant.VerificationCodeCheckError)
		return
	}
	if !isValidCode {
		response.Fail(c, http.StatusBadRequest, nil, constant.VerificationCodeExpired)
		return
	}

	// 4. 加密新密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, nil, constant.PasswordEncryptFailed)
		return
	}

	// 6. 更新用户密码
	user.Password = string(hashedPassword)
	if err := dao.UpdateUser(user); err != nil {
		response.Fail(c, http.StatusInternalServerError, nil, constant.PasswordUpdateFailed)
		return
	}

	// 7. 返回成功响应
	response.Success(c, gin.H{
		"success": true,
	}, constant.PasswordUpdateSuccess)
}

// GetUsers 聊天界面搜索用户列表
func GetUsers(c *gin.Context) {
	var req dto.SearchUsersDTO

	// 1. 绑定可选的Query参数
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, nil, constant.ParamParseError)
		return
	}

	// 2. 调用DAO层进行查询
	users, err := dao.SearchUsers(req.Username, req.UserID)
	if err != nil {
		// 其他数据库错误
		response.Fail(c, http.StatusInternalServerError, nil, constant.DatabaseError)
		return
	}

	// 3. 将用户列表转换为response层的响应结构列表
	var userResponses []response.UserBriefResponse
	for _, user := range users {
		// 调用已有的 BuildUserBriefResponse 函数处理单个用户
		userBrief := response.UserBriefResponse{
			UserID:     user.ID,
			Username:   user.Username,
			UserAvatar: utils.GetFileURL(user.Avatar),
			Status:     user.Status,
			CreateTime: user.CreatedAt.Format("2006-01-02 15:04:05"),
			Email:      user.Email,
			Role:       user.Role,
		}
		// 将处理后的用户信息添加到列表中
		userResponses = append(userResponses, userBrief)
	}

	// 4. 返回成功的响应
	response.SuccessWithData(c, userResponses, constant.GetUserSuccess)
}
