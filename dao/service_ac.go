package dao

import (
	"github.com/20gu00/gateway/common"
	"github.com/20gu00/gateway/common/gorm"
	"github.com/gin-gonic/gin"
)

type AccessControl struct {
	ID                int64  `json:"id" gorm:"primary_key"`
	ServiceID         int64  `json:"service_id" gorm:"column:service_id" description:"服务id"`
	OpenAuth          int    `json:"open_auth" gorm:"column:open_auth" description:"是否开启权限 1就是开启"`
	BlackList         string `json:"black_list" gorm:"column:black_list" description:"黑名单ip列表"`
	WhiteList         string `json:"white_list" gorm:"column:white_list" description:"白名单ip列表"`
	WhiteHostName     string `json:"white_host_name" gorm:"column:white_host_name" description:"白名单host"`
	ClientIPFlowLimit int    `json:"clientip_flow_limit" gorm:"column:clientip_flow_limit" description:"客户端限流"`
	ServiceFlowLimit  int    `json:"service_flow_limit" gorm:"column:service_flow_limit" description:"服务端限流"`
}

func (a *AccessControl) TableName() string {
	return "service_access_control"
}

func (a *AccessControl) Find(ctx *gin.Context, tx *gorm.DB, search *AccessControl) (*AccessControl, error) {
	result := &AccessControl{} //指针
	err := tx.SetCtx(common.GetGinTraceContext(ctx)).Where(search).Find(result).Error
	return result, err
}

func (a *AccessControl) Save(ctx *gin.Context, tx *gorm.DB) error {
	if err := tx.SetCtx(common.GetGinTraceContext(ctx)).Save(a).Error; err != nil { //在数据库中保存更新值，如果该值没有主键，将插入它。
		return err
	}
	return nil
}

func (a *AccessControl) ListBYServiceID(ctx *gin.Context, tx *gorm.DB, serviceID int64) ([]AccessControl, int64, error) {
	var list []AccessControl
	var count int64
	query := tx.SetCtx(common.GetGinTraceContext(ctx))
	query = query.Table(a.TableName()).Select("*")
	query = query.Where("service_id=?", serviceID)
	err := query.Order("id desc").Find(&list).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, 0, err
	}
	errCount := query.Count(&count).Error
	if errCount != nil {
		return nil, 0, err
	}
	return list, count, nil
}
