package controller

import (
	"encoding/json"
	"github.com/20gu00/gateway/common"
	"github.com/20gu00/gateway/dao"
	"github.com/20gu00/gateway/model"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
	"time"
)

var (
	SessionKey = "SessionKey"
	UserName   = "userName"
	IsLogin    = "isLogin"
)

type AdminSessionInfo struct {
	ID        uint
	UserName  string
	LoginTime time.Time
}

//swag Use
type LoginSwagIn struct {
	Username string `json:"username" example:"admin" form:"username" validate:"required"`
	Password string `json:"password" example:"passwd" form:"password" validate:"required"`
}

func AdminRegisterHandler(c *gin.Context) {
	p := new(model.Admin)
	//并不会识别输入是否符合json tag,要做参数校验
	//哪怕有一个输入字段是正确的,都能绑定,可以定义一个model,手动绑定
	if err := c.ShouldBind(p); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 2000,
			"msg":  "客户端输入的请求不正确",
			"data": err.Error(),
		})
		return
	}

	//u := model.Admin{
	//	UserName: p.UserName,
	//}
	u := model.Admin{}
	if rows := dao.DB.Where("user_name = ?", p.UserName).First(&u); rows.Error == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 2002,
			"msg":  "username存在",
			"data": rows.Error, //*gorm.DB不能传递,指针
		})
		return
	}

	p.Salt = p.UserName
	p.Password = common.SaltPassword(p.Salt, p.Password)

	//这里如果没有设置参数校验,输入如果不满足数据库的字段并不会识别错误报错,所以要做好参数校验(规则校验和数据库校验)
	if tx := dao.DB.Create(p); tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 2001,
			"msg":  "注册admin用户,数据入库失败",
			"data": tx.Error,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "注册admin用户成功",
		"data": "",
	})
}

// Login godoc
// @Summary admin_login
// @Description 网关后台管理系统管理员登录
// @Tags admin接口
// @ID /admin/login
// @Accept  application/json
// @Produce  application/json
// @Param body body controller.LoginSwagIn true "body"
// @Success 200 {object} controller.Response "success"
// @Router /admin/login [post]
func AdminLoginHandler(c *gin.Context) {
	//res := new(Response)
	p := new(model.Admin) //结构体指针
	if err := c.ShouldBind(p); err != nil {
		//res = &Response{
		//	Code: 2000,
		//	Msg:  "客户端输入的请求不正确",
		//	Data: err.Error(),
		//}
		//c.JSON(http.StatusBadRequest, res)
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 2000,
			"msg":  "客户端输入的请求不正确",
			"data": err.Error(),
		})
		return
	}

	saltPassword := common.SaltPassword(p.UserName, p.Password)
	u := model.Admin{
		UserName: p.UserName,
		Password: saltPassword,
	}

	//返回tx
	if rows := dao.DB.Where(&u).First(&u); rows.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 2001,
			"msg":  "用户名或密码错误",
			"data": rows.Error,
		})
		return
	}

	token := uuid.New().String()
	//update
	if tx := dao.DB.Model(&u).Update("token", token); tx.Error != nil {
		c.JSON(http.StatusForbidden, gin.H{
			"msg": tx.Error,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":  0,
		"msg":   "登陆成功",
		"token": token,
	})

	common.Logger.Infof("登陆成功------")
	//登陆成功设置session,含有用户部分信息
	//设置session的内容
	sessionInfo := &AdminSessionInfo{
		ID:        u.ID,
		UserName:  u.UserName,
		LoginTime: time.Now(),
	}
	//编码session
	sessionBts, err := json.Marshal(sessionInfo)
	if err != nil {
		common.Logger.Infof("json编码session信息失败")
		return
	}
	//实例化一个session(通过session来保存用户状态,判断用户是否登录)
	session := sessions.Default(c)
	//session过期时间
	session.Options(sessions.Options{MaxAge: 3600 * 6})

	//结构体->[]byte->string
	//将用户信息也存放进session
	session.Set(SessionKey, string(sessionBts))
	session.Set(UserName, p.UserName)
	session.Set(IsLogin, true)
	if err := session.Save(); err != nil {
		common.Logger.Infof("创建并保存session信息失败")
		return
	}
}

func ChangePwdHandler(c *gin.Context) {
	p := new(model.Admin)
	if err := c.ShouldBind(p); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 2000,
			"msg":  "输入不正确",
			"data": err.Error(),
		})
		return
	}

	saltPassword := common.SaltPassword(p.UserName, p.Password)
	u := model.Admin{
		UserName: p.UserName,
		Password: saltPassword,
	}

	if tx := dao.DB.Model(&u).Where("user_name = ?", p.UserName).Updates(&u); tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"msg": tx.Error,
		})
		return
	}

	id, _ := c.Get("UserId")
	if tx := dao.DB.Model(&u).Where("id = ?", id).Updates(&u); tx.Error != nil {
		c.JSON(http.StatusForbidden, gin.H{
			"msg": tx.Error,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "密码修改成功",
	})
}

func LoginOutHandler(c *gin.Context) {
	session := sessions.Default(c)
	session.Delete(SessionKey)
	session.Delete(UserName)
	session.Delete(IsLogin)

	if err := session.Save(); err != nil {
		common.Logger.Infof("退出登录过程删除session信息失败")
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "用户退出",
	})
}
