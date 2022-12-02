package dao

import (
	"github.com/20gu00/gateway/common"
	"github.com/20gu00/gateway/common/gorm"
	"github.com/gin-gonic/gin"
)

type TcpRule struct {
	ID        int64 `json:"id" gorm:"primary_key"`
	ServiceID int64 `json:"service_id" gorm:"column:service_id" description:"服务id	"`
	Port      int   `json:"port" gorm:"column:port" description:"端口	"`
}

func (t *TcpRule) TableName() string {
	return "service_tcp_rule"
}

func (t *TcpRule) Find(c *gin.Context, tx *gorm.DB, search *TcpRule) (*TcpRule, error) {
	model := &TcpRule{}
	err := tx.SetCtx(common.GetGinTraceContext(c)).Where(search).Find(model).Error
	return model, err
}

func (t *TcpRule) Save(c *gin.Context, tx *gorm.DB) error {
	if err := tx.SetCtx(common.GetGinTraceContext(c)).Save(t).Error; err != nil {
		return err
	}
	return nil
}

func (t *TcpRule) ListByServiceID(c *gin.Context, tx *gorm.DB, serviceID int64) ([]TcpRule, int64, error) {
	var list []TcpRule
	var count int64
	query := tx.SetCtx(common.GetGinTraceContext(c))
	query = query.Table(t.TableName()).Select("*")
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
