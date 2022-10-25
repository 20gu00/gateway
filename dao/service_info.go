package dao

import (
	"github.com/20gu00/gateway/common"
	"github.com/20gu00/gateway/common/gorm"
	"github.com/20gu00/gateway/dto"
	"github.com/gin-gonic/gin"
	"time"
)

type ServiceInfo struct {
	ID          int64     `json:"id" gorm:"primary_key" description:"service基本表id"`
	LoadType    int       `json:"load_type" gorm:"column:load_type" description:"负载类型 0=http 1=tcp 2=grpc"`
	ServiceName string    `json:"service_name" gorm:"column:service_name" description:"服务名称"`
	ServiceDesc string    `json:"service_desc" gorm:"column:service_desc" description:"服务描述"`
	UpdatedAt   time.Time `json:"create_at" gorm:"column:create_at" description:"更新时间"`
	CreatedAt   time.Time `json:"update_at" gorm:"column:update_at" description:"添加时间"`
	IsDelete    int8      `json:"is_delete" gorm:"column:is_delete" description:"是否已删除；0：否；1：是"`
}

func (t *ServiceInfo) TableName() string {
	return "service_info"
}

//这里也可以直接使用调用者作为参数
//serviceinfo调用的并使用serviceinfo的id进行关联其他表,组装成一个跟详细的服务信息表
func (s *ServiceInfo) ServiceDetail(ctx *gin.Context, tx *gorm.DB, search *ServiceInfo) (*ServiceDetail, error) {
	//一般调用这个方法的前会通过id拿到相应的serviceinfo信息
	if search.ServiceName == "" {
		serviceInfo, err := s.Find(ctx, tx, search)
		if err != nil {
			return nil, err
		}
		search = serviceInfo
	}

	//根据基本表的id查询多张表

	//service_tcp_rule
	tcpSerach := &TcpRule{ServiceID: search.ID}
	tcpresult, err := tcpSerach.Find(ctx, tx, tcpSerach)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	//service_access_control
	accessControlSearch := &AccessControl{ServiceID: search.ID}
	accessControlResult, err := accessControlSearch.Find(ctx, tx, accessControlSearch)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	//service_http_rule
	httpSearch := &HttpRule{ServiceID: search.ID}
	httpResult, err := httpSearch.Find(ctx, tx, httpSearch)
	if err != nil && err != gorm.ErrRecordNotFound { //数据查询不到也是当成一种err,要处理的是数据查询操作出错而不是查不到相应的数据
		return nil, err
	}

	//service_grpc_rule
	grpcSearch := &GrpcRule{ServiceID: search.ID}
	grpcResult, err := grpcSearch.Find(ctx, tx, grpcSearch)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	loadBalanceSearch := &LoadBalance{ServiceID: search.ID}
	loadBalanceResult, err := loadBalanceSearch.Find(ctx, tx, loadBalanceSearch)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	//组装成servicedetail
	serviceDetail := &ServiceDetail{
		ServiceInfo:   search,
		HTTPRule:      httpResult,
		TCPRule:       tcpresult,
		GRPCRule:      grpcResult,
		LoadBalance:   loadBalanceResult,
		AccessControl: accessControlResult,
	}
	return serviceDetail, nil
}

func (t *ServiceInfo) GroupByLoadType(c *gin.Context, tx *gorm.DB) ([]dto.MarketServiceStatItemOutput, error) {
	list := []dto.MarketServiceStatItemOutput{}
	query := tx.SetCtx(common.GetGinTraceContext(c))
	if err := query.Table(t.TableName()).Where("is_delete=0").Select("load_type, count(*) as value").Group("load_type").Scan(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

//按页列出serviceinfo
func (t *ServiceInfo) PageList(ctx *gin.Context, tx *gorm.DB, in *dto.ServiceListInput) ([]ServiceInfo, int64, error) {
	total := int64(0) //列出的service的总条数
	list := []ServiceInfo{}
	offset := (in.PageNum - 1) * in.PageSize //偏移量

	//*DB.SetCtx,设置key value进上下文
	//可以从gin的context的上下文中获取数据库的日志
	query := tx.SetCtx(common.GetGinTraceContext(ctx)).Table(t.TableName()).Where("is_delete=0")

	if in.Info != "" {
		//根据服务名称和服务描述模糊查询
		query = query.Where("(service_name like ? or service_desc like ?)", "%"+in.Info+"%", "%"+in.Info+"%")
	}

	//limit 10  limit 1,10  asc升 desc降
	if err := query.Limit(in.PageSize).Offset(offset).Order("id desc").Find(&list).Error; err != nil && err != gorm.ErrRecordNotFound { //数据库查找不到数据也是一种错误
		return nil, 0, err
	}
	query.Limit(in.PageSize).Offset(offset).Count(&total)
	return list, total, nil
}

func (s *ServiceInfo) Find(ctx *gin.Context, tx *gorm.DB, search *ServiceInfo) (*ServiceInfo, error) {
	result := &ServiceInfo{}
	err := tx.SetCtx(common.GetGinTraceContext(ctx)).Where(search).Find(result).Error
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (t *ServiceInfo) Save(c *gin.Context, tx *gorm.DB) error {
	return tx.SetCtx(common.GetGinTraceContext(c)).Save(t).Error
}
