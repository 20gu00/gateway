package controller

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/20gu00/gateway/common"
	"github.com/20gu00/gateway/common/lib"
	"github.com/20gu00/gateway/dao"
	"github.com/20gu00/gateway/dto"
	"github.com/20gu00/gateway/middleware"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
)

type AdminController struct{}

func AdminRegister(group *gin.RouterGroup) {
	adminLogin := &AdminController{}
	group.GET("/admin_info", adminLogin.AdminInfo)
	group.POST("/change_pwd", adminLogin.ChangePwd)
}

// AdminInfo godoc
// @Summary admin信息
// @Description admin信息
// @Tags admin接口
// @ID /admin/admin_info
// @Accept  json
// @Produce  json
// @Success 200 {object} middleware.Response{data=dto.AdminInfoOutput} "success"
// @Router /admin/admin_info [get]
func (adminlogin *AdminController) AdminInfo(c *gin.Context) {
	session := sessions.Default(c)                //新建Session
	sessionInfo := session.Get(common.SessionKey) //从context中获取session的key对应的value

	//新建一个管理员的session信息的结构体,将上下文中拿到的sessioninfo解码放入adminsessioninfo中
	adminSessionInfo := &dto.AdminSessionInfo{}
	if err := json.Unmarshal([]byte(fmt.Sprint(sessionInfo)), adminSessionInfo); err != nil {
		middleware.ResponseError(c, 2000, err)
		return
	}

	output := &dto.AdminInfoOutput{
		ID:           adminSessionInfo.ID,
		Name:         adminSessionInfo.UserName,
		LoginTime:    adminSessionInfo.LoginTime,
		Avatar:       "https://camo.githubusercontent.com/2b507540e2681c1a25698f246b9dca69c30548ed66a7323075b0224cbb1bf058/68747470733a2f2f676f6c616e672e6f72672f646f632f676f706865722f6669766579656172732e6a7067",
		Introduction: "admin超管",
		Roles:        []string{"admin"},
	}
	//成功响应,数据在context中传输,有处理响应的中间件ResponseSuccess包装resp响应请求
	middleware.ResponseSuccess(c, output)
}

// ChangePwd godoc
// @Summary ChangePwd
// @Description ChangePwd
// @Tags admin接口
// @ID /admin/change_pwd
// @Accept  json
// @Produce  json
// @Param body body dto.ChangePwdInput true "body"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /admin/change_pwd [post]
func (a *AdminController) ChangePwd(ctx *gin.Context) {
	in := &dto.ChangePwdInput{}
	if err := in.BindValidParam(ctx); err != nil {
		middleware.ResponseError(ctx, 2000, err)
		return
	}

	//修改密码逻辑,需要用户已经登录,能获取用户信息(获取session)说明用户处于登录状态
	session := sessions.Default(ctx)
	sessionInfo := session.Get(common.SessionKey)
	adminSessionInfo := &dto.AdminSessionInfo{}
	if err := json.Unmarshal([]byte(fmt.Sprint(sessionInfo)), adminSessionInfo); err != nil {
		middleware.ResponseError(ctx, 2000, errors.New("修改密码接口使用报错,确保用户已经登录"))
		return
	}

	tx, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(ctx, 2001, err)
		return
	}
	adminInfo := &dao.Admin{}
	adminInfo, err = adminInfo.Find(ctx, tx, (&dao.Admin{UserName: adminSessionInfo.UserName}))
	if err != nil {
		middleware.ResponseError(ctx, 2002, err)
		return
	}

	saltPassword := common.SaltPassword(adminInfo.Salt, in.Password)
	adminInfo.Password = saltPassword

	//数据写入数据库
	if err := adminInfo.Save(ctx, tx); err != nil {
		middleware.ResponseError(ctx, 2003, err)
		return
	}

	middleware.ResponseSuccess(ctx, "修改密码成功")
}
