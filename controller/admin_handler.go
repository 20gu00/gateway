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
	if err := c.ShouldBindJSON(p); err != nil {
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
	if err := c.ShouldBindJSON(p); err != nil {
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

	//设置token
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
	//设置session的内容(服务端保存)
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
	session.Set(common.SessionKey, string(sessionBts))
	session.Set(common.UserName, p.UserName)
	session.Set(common.IsLogin, true)
	if err := session.Save(); err != nil {
		common.Logger.Infof("创建并保存session信息失败")
		return
	}
}

func ChangePwdHandler(c *gin.Context) {
	p := new(model.Admin)
	if err := c.ShouldBindJSON(p); err != nil {
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
	session.Delete(common.SessionKey)
	session.Delete(common.UserName)
	session.Delete(common.IsLogin)

	if err := session.Save(); err != nil {
		common.Logger.Infof("退出登录过程删除session信息失败")
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "用户退出",
	})
}

func AdminInfoHandler(c *gin.Context) {
	session := sessions.Default(c)                //新建Session
	sessionInfo := session.Get(common.SessionKey) //从context中获取session的key对应的value

	//将上下文中拿到的sessioninfo解码放入adminsessioninfo中
	adminSessionInfo := &model.AdminSessionInfo{}

	//序列化结构体->json字符串 反序列化json字符串->结构体信息
	//[]byte(fmt.Sprint(sessionInfo)
	if err := json.Unmarshal([]byte(sessionInfo.(string)), adminSessionInfo); err != nil { //fmt.Sprint打印字符串,转换成[]byte,在反序列化(本质就是处理[]byte(str))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 2000,
			"msg":  "将session信息反序列化失败",
			"data": err.Error(),
		})
		return
	}

	output := &model.AdminInfoOutput{
		ID:        adminSessionInfo.ID,
		Name:      adminSessionInfo.UserName,
		LoginTime: adminSessionInfo.LoginTime,
		Avatar:    "https://camo.githubusercontent.com/2b507540e2681c1a25698f246b9dca69c30548ed66a7323075b0224cbb1bf058/68747470733a2f2f676f6c616e672e6f72672f646f632f676f706865722f6669766579656172732e6a7067",
		//Introduction: "admin超管",
		//Roles:        []string{"admin"}   //后续可以用来做权限控制
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "返回session信息",
		"data": output,
	})
}
