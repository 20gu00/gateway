package dao

import (
	"github.com/20gu00/gateway/common"
	"github.com/20gu00/gateway/common/gorm"
	"github.com/20gu00/gateway/common/lib"
	"github.com/20gu00/gateway/dto"
	"github.com/gin-gonic/gin"
	"net/http/httptest"
	"sync"
	"time"
)

type Tenant struct {
	ID        int64     `json:"id" gorm:"primary_key"`
	AppID     string    `json:"app_id" gorm:"column:app_id" description:"租户id"`
	Name      string    `json:"name" gorm:"column:name" description:"租户名称"`
	Secret    string    `json:"secret" gorm:"column:secret" description:"密钥"`
	WhiteIPS  string    `json:"white_ips" gorm:"column:white_ips" description:"ip白名单，支持前缀匹配"`
	Qpd       int64     `json:"qpd" gorm:"column:qpd" description:"日请求量限制"`
	Qps       int64     `json:"qps" gorm:"column:qps" description:"每秒请求量限制"`
	CreatedAt time.Time `json:"create_at" gorm:"column:create_at" description:"添加时间"`
	UpdatedAt time.Time `json:"update_at" gorm:"column:update_at" description:"更新时间"`
	IsDelete  int8      `json:"is_delete" gorm:"column:is_delete" description:"是否已删除；0：否；1：是"`
}

func (t *Tenant) TableName() string {
	return "tenant"
}

func (t *Tenant) Find(c *gin.Context, tx *gorm.DB, search *Tenant) (*Tenant, error) {
	model := &Tenant{}
	err := tx.SetCtx(common.GetGinTraceContext(c)).Where(search).Find(model).Error
	return model, err
}

func (t *Tenant) Save(c *gin.Context, tx *gorm.DB) error {
	if err := tx.SetCtx(common.GetGinTraceContext(c)).Save(t).Error; err != nil {
		return err
	}
	return nil
}

func (t *Tenant) TenantList(c *gin.Context, tx *gorm.DB, params *dto.TenantListInput) ([]Tenant, int64, error) {
	var list []Tenant
	var count int64
	pageNo := params.PageNo
	pageSize := params.PageSize

	//limit offset,pagesize
	offset := (pageNo - 1) * pageSize
	query := tx.SetCtx(common.GetGinTraceContext(c))
	query = query.Table(t.TableName()).Select("*")
	query = query.Where("is_delete=?", 0)
	if params.Info != "" {
		query = query.Where(" (name like ? or app_id like ?)", "%"+params.Info+"%", "%"+params.Info+"%")
	}
	err := query.Limit(pageSize).Offset(offset).Order("id desc").Find(&list).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, 0, err
	}
	errCount := query.Count(&count).Error
	if errCount != nil {
		return nil, 0, err
	}
	return list, count, nil
}

var AppManagerHandler *TenantManager

func init() {
	AppManagerHandler = NewTenantManager()
}

type TenantManager struct {
	AppMap   map[string]*Tenant
	AppSlice []*Tenant
	Locker   sync.RWMutex
	init     sync.Once
	err      error
}

func NewTenantManager() *TenantManager {
	return &TenantManager{
		AppMap:   map[string]*Tenant{},
		AppSlice: []*Tenant{},
		Locker:   sync.RWMutex{},
		init:     sync.Once{},
	}
}

//获取租户列表
func (s *TenantManager) GetTenantList() []*Tenant {
	return s.AppSlice
}

func (s *TenantManager) LoadOnce() error {
	s.init.Do(func() {
		appInfo := &Tenant{}
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		tx, err := lib.GetGormPool("default")
		if err != nil {
			s.err = err
			return
		}
		params := &dto.TenantListInput{PageNo: 1, PageSize: 99999}
		list, _, err := appInfo.TenantList(c, tx, params)
		if err != nil {
			s.err = err
			return
		}
		s.Locker.Lock()
		defer s.Locker.Unlock()
		for _, listItem := range list {
			tmpItem := listItem
			s.AppMap[listItem.AppID] = &tmpItem
			s.AppSlice = append(s.AppSlice, &tmpItem)
		}
	})
	return s.err
}
