package controller

import (
	"fmt"
	"github.com/20gu00/gateway/common"
	"github.com/20gu00/gateway/common/lib"
	"github.com/20gu00/gateway/dao"
	"github.com/20gu00/gateway/dto"
	"github.com/20gu00/gateway/middleware"
	"github.com/gin-gonic/gin"
)

// ServiceList godoc
// @Summary service列表
// @Description service列表
// @Tags service管理
// @ID /service/list
// @Accept  json
// @Produce  json
// @Param info query string false "关键词"
// @Param page_size query int true "每页个数"
// @Param page_num query int true "当前页数"
// @Success 200 {object} middleware.Response{data=dto.ServiceListOutput} "success"
// @Router /service/list [get]
func (s *ServiceController) ServiceList(ctx *gin.Context) {
	in := &dto.ServiceListInput{}
	if err := in.BindValidParam(ctx); err != nil {
		middleware.ResponseError(ctx, 2000, err)
		return
	}

	tx, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(ctx, 2001, err)
		return
	}

	serviceInfo := &dao.ServiceInfo{}
	serviceInfoList, total, err := serviceInfo.PageList(ctx, tx, in)
	if err != nil {
		middleware.ResponseError(ctx, 2002, err)
		return
	}

	outList := []dto.ServiceListItemOutput{}
	for _, listItem := range serviceInfoList {
		serviceDetail, err := listItem.ServiceDetail(ctx, tx, &listItem)
		if err != nil {
			middleware.ResponseError(ctx, 2003, err)
			return
		}

		//1、http前缀接入 clusterIP:clusterPort+path 2、http域名接入 domain 3、tcp、grpc接入 clusterIP:servicePort
		serviceAddr := "unknowServiceAddr"
		//网关的ip,端口
		clusterIP := lib.GetStringConf("base.cluster.cluster_ip")
		clusterPort := lib.GetStringConf("base.cluster.cluster_port")
		clusterSSLPort := lib.GetStringConf("base.cluster.cluster_ssl_port")

		//http类型,接入类型,是否开启https
		if serviceDetail.ServiceInfo.LoadType == common.LoadTypeHTTP && serviceDetail.HTTPRule.RuleType == common.HTTPRuleTypePrefixURL && serviceDetail.HTTPRule.NeedHttps == 1 {
			serviceAddr = fmt.Sprintf("%s:%s%s", clusterIP, clusterSSLPort, serviceDetail.HTTPRule.Rule)
		}

		if serviceDetail.ServiceInfo.LoadType == common.LoadTypeHTTP && serviceDetail.HTTPRule.RuleType == common.HTTPRuleTypePrefixURL && serviceDetail.HTTPRule.NeedHttps == 0 {
			serviceAddr = fmt.Sprintf("%s:%s%s", clusterIP, clusterPort, serviceDetail.HTTPRule.Rule)
		}

		//域名
		if serviceDetail.ServiceInfo.LoadType == common.LoadTypeHTTP && serviceDetail.HTTPRule.RuleType == common.HTTPRuleTypeDomain {
			serviceAddr = serviceDetail.HTTPRule.Rule
		}

		//
		if serviceDetail.ServiceInfo.LoadType == common.LoadTypeTCP {
			serviceAddr = fmt.Sprintf("%s:%d", clusterIP, serviceDetail.TCPRule.Port)
		}

		if serviceDetail.ServiceInfo.LoadType == common.LoadTypeGRPC {
			serviceAddr = fmt.Sprintf("%s:%d", clusterIP, serviceDetail.GRPCRule.Port)
		}

		//实际的工作负载
		ipList := serviceDetail.LoadBalance.GetIPListByModel()
		//流量统计,统计这个服务的流量
		counter, err := common.FlowCounterHandler.GetCounter(common.FlowServicePrefix + listItem.ServiceName) //服务流量统计前缀+服务名称
		if err != nil {
			middleware.ResponseError(ctx, 2004, err)
			return
		}

		outItem := dto.ServiceListItemOutput{
			ID:          listItem.ID,
			LoadType:    listItem.LoadType,
			ServiceName: listItem.ServiceName,
			ServiceDesc: listItem.ServiceDesc,
			ServiceAddr: serviceAddr,
			Qps:         counter.QPS,
			Qpd:         counter.TotalCount,
			TotalNode:   len(ipList),
		}
		outList = append(outList, outItem)
	}
	out := &dto.ServiceListOutput{
		Total: total,
		List:  outList,
	}
	middleware.ResponseSuccess(ctx, out)
}
