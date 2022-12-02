package controller

import (
	"encoding/json"
	"errors"
	"github.com/20gu00/gateway/common"
	"github.com/20gu00/gateway/common/lib"
	"github.com/20gu00/gateway/dao"
	"github.com/20gu00/gateway/dto"
	"github.com/20gu00/gateway/middleware"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"time"
)

type AdminLoginController struct{}

func AdminLoginRegister(group *gin.RouterGroup) {
	adminLogin := &AdminLoginController{}
	group.POST("/login", adminLogin.Login)
	group.GET("/logout", adminLogin.LoginOut)
}

// Login godoc
// @Summary 登录
// @Description 登录
// @Tags admin接口
// @ID /admin_login/login
// @Accept  json
// @Produce  json
// @Param body body dto.AdminLoginInput true "body"
// @Success 200 {object} middleware.Response{data=dto.AdminLoginOutput} "success"
// @Router /admin_login/login [post]
func (a *AdminLoginController) Login(ctx *gin.Context) {
	in := &dto.AdminLoginInput{}
	//参数校验
	if err := in.BindValidParam(ctx); err != nil {
		middleware.ResponseError(ctx, 2000, err)
		return
	}

	//admininfo.salt + params.Password sha256 => saltPassword
	tx, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(ctx, 2001, err)
		return
	}

	admin := &dao.Admin{}
	admin, err = admin.LoginCheck(ctx, tx, in)
	if err != nil {
		middleware.ResponseError(ctx, 2002, err)
		return
	}

	//登录成功的用户(管理员)设置session(服务端保存)
	sessionInfo := &dto.AdminSessionInfo{
		//使用管理员的信息
		ID:        admin.Id,
		UserName:  admin.UserName,
		LoginTime: time.Now(),
	}
	sessionBts, err := json.Marshal(sessionInfo) //json编码
	if err != nil {
		middleware.ResponseError(ctx, 2003, err)
		return
	}

	session := sessions.Default(ctx)                   //gin官方session
	session.Set(common.SessionKey, string(sessionBts)) //存入sessioninfo进行json编码后的字符串,sessioninfo也是根据serviceinfo建立
	if err = session.Save(); err != nil {              //保存
		middleware.ResponseError(ctx, 2004, errors.New("登录后创建session失败"))
		return
	}
	//这里也可以直接用各方发生token
	output := &dto.AdminLoginOutput{Token: admin.UserName} //返回个简单的token
	middleware.ResponseSuccess(ctx, output)
}

// LoginOut godoc
// @Summary 退出
// @Description 退出
// @Tags admin接口
// @ID /admin_login/logout
// @Accept  json
// @Produce  json
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /admin_login/logout [get]
func (a *AdminLoginController) LoginOut(ctx *gin.Context) {
	session := sessions.Default(ctx)
	session.Delete(common.SessionKey)
	if err := session.Save(); err != nil {
		middleware.ResponseError(ctx, 2000, errors.New("退出后保存删除session操作失败"))
	}
	middleware.ResponseSuccess(ctx, "")
}
