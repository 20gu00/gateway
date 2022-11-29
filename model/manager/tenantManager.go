package manager

import (
	"github.com/20gu00/gateway/dao"
	"github.com/20gu00/gateway/model"
	"sync"
)

//租户的管理
var AppManagerHandler *TenantManager

func init() {
	AppManagerHandler = NewTenantManager()
}

type TenantManager struct {
	AppMap   map[string]*model.Tenant
	AppSlice []*model.Tenant
	Locker   sync.RWMutex
	init     sync.Once
	err      error
}

func NewTenantManager() *TenantManager {
	return &TenantManager{
		AppMap:   map[string]*model.Tenant{},
		AppSlice: []*model.Tenant{},
		Locker:   sync.RWMutex{},
		init:     sync.Once{},
	}
}

//获取租户列表
func (s *TenantManager) GetTenantList() []*model.Tenant {
	return s.AppSlice
}

func (s *TenantManager) LoadOnce() error {
	s.init.Do(func() {
		appInfo := &model.Tenant{}
		//c, _ := gin.CreateTestContext(httptest.NewRecorder())
		db := dao.DB
		params := &model.TenantListInput{PageNo: 1, PageSize: 99999}
		list, _, err := appInfo.TenantList(db, params)
		if err != nil {
			s.err = err
			return
		}
		s.Locker.Lock()
		defer s.Locker.Unlock()
		for _, listItem := range list {
			tmpItem := listItem
			s.AppMap[listItem.AppId] = &tmpItem
			s.AppSlice = append(s.AppSlice, &tmpItem)
		}
	})
	return s.err
}
