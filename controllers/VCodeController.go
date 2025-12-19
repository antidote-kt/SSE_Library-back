package controllers

import (
	"net/http"

	"github.com/antidote-kt/SSE_Library-back/constant"
	"github.com/antidote-kt/SSE_Library-back/dao"
	"github.com/antidote-kt/SSE_Library-back/dto"
	"github.com/antidote-kt/SSE_Library-back/response"
	"github.com/antidote-kt/SSE_Library-back/utils"
	"github.com/gin-gonic/gin"
)

// SendVerificationCode 处理发送验证码的请求
func SendVerificationCode(c *gin.Context) {
	var req dto.SendCodeDTO

	// 1. 绑定并验证参数
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, nil, constant.ParamParseError+err.Error())
		return
	}

	// 2. 检查邮箱是否已注册
	if req.Usage == "reset-password" {
		// 如果是用于重置密码，需要确保邮箱已经注册用户
		_, err := dao.GetUserByEmail(req.Email)
		if err != nil {
			response.Fail(c, http.StatusNotFound, nil, constant.UnauthorizedEmail)
			return
		}
	} else if req.Usage == "register" {
		// 如果是注册，需要检查邮箱是否已被占用
		_, err := dao.GetUserByEmail(req.Email)
		if err == nil {
			response.Fail(c, http.StatusUnprocessableEntity, nil, constant.EmailHasBeenUsed)
			return
		}

	} else {
		response.Fail(c, http.StatusBadRequest, nil, constant.InvalidTransaction)
		return
	}

	// 3. 生成验证码
	code := utils.GenerateVerificationCode(6) // 生成6位数字验证码

	// 4. 将验证码存入Redis
	err := utils.StoreVerificationCode(req.Email, req.Usage, code)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, nil, constant.VerificationCodeStoreError)
		return
	}

	// 5. 发送邮件 (这里不采用异步执行，以保证前端真实接收邮件发送结果)
	// 如果异步执行，会跳过异步执行的协程继续往下执行成功响应，而一旦成功响应发送给了前端，后续的邮件发送结果就会被堵塞，无法改写为报错，前端也就看不到错误信息
	err = utils.SendVerificationEmail(req.Email, code)
	if err != nil {
		// 给前端返回常量错误响应（而非err的具体错误信息）
		response.Fail(c, http.StatusInternalServerError, nil, constant.VerificationCodeSendFailed)
		return
	}

	// 6. 返回成功响应
	response.Success(c, gin.H{
		"success": true,
	}, constant.VCodeSendSuccess)
}
