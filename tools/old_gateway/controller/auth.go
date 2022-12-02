package controller

import (
	"encoding/base64"
	"github.com/20gu00/gateway/common"
	"github.com/20gu00/gateway/common/lib"
	"github.com/20gu00/gateway/dao"
	"github.com/20gu00/gateway/dto"
	"github.com/20gu00/gateway/middleware"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"strings"
	"time"
)

type AuthController struct{}

func AuthRegister(group *gin.RouterGroup) {
	auth := &AuthController{}
	group.POST("/token", auth.GetToken)
}

// GetToken godoc
// @Summary 获取token
// @Description 获取token
// @Tags AUTH
// @ID /oauth/token
// @Accept  json
// @Produce  json
// @Param body body dto.TokensInput true "body"
// @Success 200 {object} middleware.Response{data=dto.TokensOutput} "success"
// @Router /auth/tokens [post]

//获取token属于代理逻辑,不是后台管理系统的逻辑
func (oauth *AuthController) GetToken(c *gin.Context) {
	params := &dto.TokensInput{}
	if err := params.BindValidParam(c); err != nil {
		middleware.ResponseError(c, 2000, err)
		return
	}

	//Authorization: Beare (usernamepassword)
	splits := strings.Split(c.GetHeader("Authorization"), " ")
	if len(splits) != 2 {
		middleware.ResponseError(c, 2001, errors.New("Authorization格式错误"))
		return
	}

	appSecret, err := base64.StdEncoding.DecodeString(splits[1]) //base64解码
	if err != nil {
		middleware.ResponseError(c, 2002, err)
		return
	}

	parts := strings.Split(string(appSecret), ":")
	if len(parts) != 2 {
		middleware.ResponseError(c, 2003, errors.New("用户名或密码格式错误"))
		return
	}

	appList := dao.AppManagerHandler.GetTenantList()
	for _, appInfo := range appList {
		if appInfo.AppID == parts[0] && appInfo.Secret == parts[1] {
			claims := jwt.StandardClaims{ //标准jwt的字段
				Issuer:    appInfo.AppID,
				ExpiresAt: time.Now().Add(common.JwtExpires * time.Second).In(lib.TimeLocation).Unix(),
			}
			token, err := common.JwtEncode(claims)
			if err != nil {
				middleware.ResponseError(c, 2004, err)
				return
			}
			output := &dto.TokensOutput{
				ExpiresIn:   common.JwtExpires,
				TokenType:   "Bearer",
				AccessToken: token,
				Scope:       "read_write",
			}
			middleware.ResponseSuccess(c, output)
			return
		}
	}
	middleware.ResponseError(c, 2005, errors.New("未匹配正确APP信息"))
}

// Login godoc
// @Summary 管理员退出
// @Description 管理员退出
// @Tags admin接口
// @ID /admin_login/logout
// @Accept  json
// @Produce  json
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /admin_login/logout [get]
//func (adminlogin *AuthController) AdminLoginOut(c *gin.Context) {
//	sess := sessions.Default(c)
//	sess.Delete(common.SessionKey)
//	sess.Save()
//	middleware.ResponseSuccess(c, "")
//}
