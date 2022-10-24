package dto

import (
	"github.com/20gu00/gateway/common"
	"github.com/gin-gonic/gin"
	"time"
)

type AdminSessionInfo struct {
	ID        int       `json:"id"`
	UserName  string    `json:"user_name"`
	LoginTime time.Time `json:"login_time"`
}

type AdminLoginInput struct {
	UserName string `json:"username" form:"username" comment:"admin用户名" example:"admin" validate:"required,valid_username"` //admin用户名
	Password string `json:"password" form:"password" comment:"密码" example:"passwd" validate:"required"`                     //密码
}

type AdminLoginOutput struct {
	Token string `json:"token" form:"token" comment:"token" example:"token" validate:""` //token
}

//参数校验
func (param *AdminLoginInput) BindValidParam(c *gin.Context) error {
	return common.ValidDefaultParams(c, param)
}
