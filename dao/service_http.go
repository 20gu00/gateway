package dao

import (
	"github.com/20gu00/gateway/common"
	"github.com/20gu00/gateway/common/gorm"
	"github.com/gin-gonic/gin"
)

type HttpRule struct {
	ID             int64  `json:"id" gorm:"primary_key"`
	ServiceID      int64  `json:"service_id" gorm:"column:service_id" description:"服务id"`
	RuleType       int    `json:"rule_type" gorm:"column:rule_type" description:"匹配类型:domain(1)和url_prefix(0)"`
	Rule           string `json:"rule" gorm:"column:rule" description:"根据ruleType来定,域名或者前缀"`
	NeedHttps      int    `json:"need_https" gorm:"column:need_https" description:"1就是https"`
	NeedWebsocket  int    `json:"need_websocket" gorm:"column:need_websocket" description:"是否开启websocket,1就是启用"`
	NeedStripUri   int    `json:"need_strip_uri" gorm:"column:need_strip_uri" description:"是否开启strip_uri,就是启用"`
	UrlRewrite     string `json:"url_rewrite" gorm:"column:url_rewrite" description:"url重写功能"`
	HeaderTransfor string `json:"header_transfor" gorm:"column:header_transfor" description:"header转换(add)(del)(edit)"`
}

func (t *HttpRule) TableName() string {
	return "service_http_rule"
}

func (t *HttpRule) Find(c *gin.Context, tx *gorm.DB, search *HttpRule) (*HttpRule, error) {
	model := &HttpRule{}
	err := tx.SetCtx(common.GetGinTraceContext(c)).Where(search).Find(model).Error
	return model, err
}

func (t *HttpRule) Save(c *gin.Context, tx *gorm.DB) error {
	if err := tx.SetCtx(common.GetGinTraceContext(c)).Save(t).Error; err != nil {
		return err
	}
	return nil
}

func (t *HttpRule) ListByServiceID(c *gin.Context, tx *gorm.DB, serviceID int64) ([]HttpRule, int64, error) {
	var list []HttpRule
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
