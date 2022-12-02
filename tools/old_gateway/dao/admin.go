package dao

import (
	"github.com/20gu00/gateway/common"
	"github.com/20gu00/gateway/common/gorm"
	"github.com/20gu00/gateway/dto"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"time"
)

type Admin struct {
	Id        int       `json:"id" gorm:"primary_key" description:"主键"`
	UserName  string    `json:"user_name" gorm:"column:user_name" description:"admin用户名"`
	Salt      string    `json:"salt" gorm:"column:salt" description:"加密密码用的salt"`
	Password  string    `json:"password" gorm:"column:password" description:"密码"`
	UpdatedAt time.Time `json:"update_at" gorm:"column:update_at" description:"更新时间"` //加d那么gorm会自动处理
	CreatedAt time.Time `json:"create_at" gorm:"column:create_at" description:"创建时间"`
	IsDelete  int       `json:"is_delete" gorm:"column:is_delete" description:"是否删除"`
}

//获取数据表名称
func (a *Admin) TableName() string {
	return "admin"
}

//数据库查询
func (a *Admin) Find(ctx *gin.Context, tx *gorm.DB, search *Admin) (*Admin, error) {
	result := &Admin{} //查询输出的结构体
	//结构体查询
	//自定义的context,能存入mysql查询日志
	err := tx.SetCtx(common.GetGinTraceContext(ctx)).Where(search).Find(result).Error
	if err != nil {
		return nil, err
	}

	return result, nil
}

//保存数据
func (a *Admin) Save(ctx *gin.Context, tx *gorm.DB) error {
	return tx.SetCtx(common.GetGinTraceContext(ctx)).Save(a).Error
}

//登录检测
func (t *Admin) LoginCheck(ctx *gin.Context, tx *gorm.DB, in *dto.AdminLoginInput) (*Admin, error) {
	info, err := t.Find(ctx, tx, (&Admin{UserName: in.UserName, IsDelete: 0}))
	if err != nil {
		return nil, errors.New("不存在该用户的信息")
	}

	saltPassword := common.SaltPassword(info.Salt, in.Password)
	if info.Password != saltPassword {
		return nil, errors.New("密码错误")
	}
	return info, nil
}
