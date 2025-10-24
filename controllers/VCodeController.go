package controllers

import (
	"log"
	"net/http"

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
		response.Fail(c, http.StatusBadRequest, nil, "参数格式错误: "+err.Error())
		return
	}

	// 2. 检查邮箱是否已注册
	if req.Usage == "reset-password" {
		// 如果是用于重置密码，需要确保邮箱已经注册用户
		_, err := dao.GetUserByEmail(req.Email)
		if err != nil {
			response.Fail(c, http.StatusNotFound, nil, "该邮箱未注册")
			return
		}
	} else if req.Usage == "register" {
		// 如果是注册，需要检查邮箱是否已被占用
		_, err := dao.GetUserByEmail(req.Email)
		if err == nil {
			response.Fail(c, http.StatusUnprocessableEntity, nil, "邮箱已被注册")
			return
		}

	} else {
		response.Fail(c, http.StatusBadRequest, nil, "无效的验证码业务")
		return
	}

	// 3. 生成验证码
	code := utils.GenerateVerificationCode(6) // 生成6位数字验证码

	// 4. 将验证码存入Redis
	err := utils.StoreVerificationCode(req.Email, req.Usage, code)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, nil, "存储验证码失败")
		return
	}

	// 5. 发送邮件 (这里可以异步执行以提高接口响应速度)
	go func() {
		err := utils.SendVerificationEmail(req.Email, code)
		if err != nil {
			// 记录发送失败的日志，不影响给前端的成功响应
			log.Printf("异步发送验证码到 %s 失败: %v", req.Email, err)
		}
	}()

	// 6. 返回成功响应
	response.Success(c, gin.H{
		"success": true,
	}, "验证码已发送至您的邮箱，请注意查收")
}
