package tcpController

import (
	"github.com/20gu00/gateway/common"
	"github.com/20gu00/gateway/dao"
	"github.com/20gu00/gateway/model"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

func ServiceAddTcpHandler(c *gin.Context) {
	p := new(model.ServiceAddTcpInput)
	if err := c.ShouldBind(p); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 2000,
			"msg":  "输入的请求参数不正确",
			"data": err.Error(),
		})
	}

	db := dao.DB

	infoSearch := &model.ServiceInfo{
		ServiceName: p.ServiceName,
		IsDelete:    0, //没有删除的用户
	}

	if tx := db.Where(infoSearch).First(infoSearch); tx.Error == nil {
		db.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 2002,
			"msg":  "服务已存在,该服务名称已经被使用(tcp)",
			"data": tx.Error,
		})
		return
	}

	tcpRuleSearch := &model.Service_tcp{
		Port: p.Port,
	}
	if tx := db.Where(tcpRuleSearch).First(tcpRuleSearch); tx.Error == nil {
		db.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 2003,
			"msg":  "端口被占用(tcp)",
			"data": tx.Error,
		})
		return
	}

	if len(strings.Split(p.IpList, ",")) != len(strings.Split(p.WeightList, ",")) {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":        2004,
			"msg":         "ip列表与权重设置不匹配",
			"data_ipList": strings.Split(p.IpList, ","),
			"data_weight": strings.Split(p.WeightList, ","),
		})
		return
	}

	db.Begin()
	serviceInfo := &model.ServiceInfo{
		LoadType:    common.LoadTypeTCP,
		ServiceName: p.ServiceName,
		ServiceDesc: p.ServiceDesc,
	}
	if tx := db.Save(serviceInfo); tx.Error != nil {
		db.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 2005,
			"msg":  "写入service基本表失败",
		})
		return
	}

	loadBalanceDao := &model.LoadBalance{
		ServiceId:  int(serviceInfo.ID),
		RoundType:  p.RoundType,
		IpList:     p.IpList,
		WeightList: p.WeightList,
		ForbidList: p.ForbidList,
	}
	if tx := db.Save(loadBalanceDao); tx.Error != nil {
		db.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 2006,
			"msg":  "写入loadBalance表失败",
		})
		return
	}

	tcpDao := &model.Service_tcp{
		ServiceId: int(serviceInfo.ID),
		Port:      p.Port,
	}
	if tx := db.Save(tcpDao); tx.Error != nil {
		db.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 2007,
			"msg":  "写入tcp表失败",
		})
		return
	}

	accessControlDao := &model.AccessControl{
		ServiceId:         int(serviceInfo.ID),
		OpenAuth:          p.OpenAuth,
		BlackList:         p.BlackList,
		WhiteList:         p.WhiteList,
		WhiteHostName:     p.WhiteHostName,
		ClientIPFlowLimit: p.ClientIPFlowLimit,
		ServiceFlowLimit:  p.ServiceFlowLimit,
	}
	if tx := db.Save(accessControlDao); tx.Error != nil {
		db.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 2008,
			"msg":  "写入accessControl表失败",
		})
		return
	}

	db.Commit()
	c.JSON(http.StatusBadRequest, gin.H{
		"code": 0,
		"msg":  "添加tcp服务成功",
	})
}

func ServiceUpdateTcpHandler(c *gin.Context) {
	p := new(model.ServiceUpdateTcpInput)
	if err := c.ShouldBind(p); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 2000,
			"msg":  "输入的请求参数不正确",
			"data": err.Error(),
		})
	}

	if len(strings.Split(p.IpList, ",")) != len(strings.Split(p.WeightList, ",")) {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":        2004,
			"msg":         "ip列表与权重设置不匹配",
			"data_ipList": strings.Split(p.IpList, ","),
			"data_weight": strings.Split(p.WeightList, ","),
		})
		return
	}

	db := dao.DB
	db.Begin()

	serviceInfo := &model.ServiceInfo{
		ServiceName: p.ServiceName,
	}
	if tx := db.Where(serviceInfo).First(serviceInfo); tx.Error != nil {
		db.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 2000,
			"msg":  "服务不存在",
			"data": tx.Error,
		})
		return
	}

	serviceDetail, err := serviceInfo.ServiceDetail(db, serviceInfo)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 2001,
			"msg":  "获取该服务的详情失败",
			"data": db.Error,
		})
		return
	}

	seriverInfoDao := serviceDetail.BaseInfo
	seriverInfoDao.ServiceDesc = p.ServiceDesc
	if tx := db.Save(seriverInfoDao); tx.Error != nil {
		db.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 2002,
			"msg":  "写入service基本表失败",
			"data": tx.Error,
		})
		return
	}

	loadBalanceDao := &model.LoadBalance{}
	if serviceDetail.LoadBalance != nil {
		loadBalanceDao = serviceDetail.LoadBalance
	}

	loadBalanceDao.ServiceId = int(serviceInfo.ID)
	loadBalanceDao.RoundType = p.RoundType
	loadBalanceDao.IpList = p.IpList
	loadBalanceDao.WeightList = p.WeightList
	loadBalanceDao.ForbidList = p.ForbidList

	if tx := db.Save(loadBalanceDao); tx.Error != nil {
		db.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 2003,
			"msg":  "写入loadbalance表失败",
			"data": tx.Error,
		})
		return
	}

	tcpRuleDao := &model.Service_tcp{}
	if serviceDetail.Tcp != nil {
		tcpRuleDao = serviceDetail.Tcp
	}
	if tx := db.Save(tcpRuleDao); tx.Error != nil {
		db.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 2004,
			"msg":  "写入tcp表失败",
			"data": tx.Error,
		})
		return
	}

	accessControlDao := &model.AccessControl{}
	if serviceDetail.AccessControl != nil {
		accessControlDao = serviceDetail.AccessControl
	}
	accessControlDao.ServiceId = int(serviceInfo.ID)
	accessControlDao.OpenAuth = p.OpenAuth
	accessControlDao.BlackList = p.BlackList
	accessControlDao.WhiteList = p.WhiteList
	accessControlDao.WhiteHostName = p.WhiteHostName
	accessControlDao.ClientIPFlowLimit = p.ClientIPFlowLimit
	accessControlDao.ServiceFlowLimit = p.ServiceFlowLimit
	if tx := db.Save(accessControlDao); tx.Error != nil {
		db.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 2005,
			"msg":  "写入accessControl表失败",
			"data": tx.Error,
		})
		return
	}

	db.Commit()
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "修改tcp表成功",
	})
}

type ServiceAddTcpInput struct {
	ServiceName       string `json:"service_name"`
	ServiceDesc       string `json:"service_desc"`
	Port              int    `json:"port"`
	HeaderTransfor    string `json:"header_transfor"`
	OpenAuth          int    `json:"open_auth"`
	BlackList         string `json:"black_list"`
	WhiteList         string `json:"white_list"`
	WhiteHostName     string `json:"white_host_name"`
	ClientIPFlowLimit int    `json:"clientip_flow_limit"`
	ServiceFlowLimit  int    `json:"service_flow_limit"`
	RoundType         int    `json:"round_type"`
	IpList            string `json:"ip_list"`
	WeightList        string `json:"weight_list"`
	ForbidList        string `json:"forbid_list"`
}

type ServiceUpdateTcpInput struct {
	ID                int64  `json:"id"`
	ServiceName       string `json:"service_name"`
	ServiceDesc       string `json:"service_desc"`
	Port              int    `json:"port"`
	OpenAuth          int    `json:"open_auth"`
	BlackList         string `json:"black_list"`
	WhiteList         string `json:"white_list"`
	WhiteHostName     string `json:"white_host_name"`
	ClientIPFlowLimit int    `json:"clientip_flow_limit"`
	ServiceFlowLimit  int    `json:"service_flow_limit"`
	RoundType         int    `json:"round_type"`
	IpList            string `json:"ip_list"`
	WeightList        string `json:"weight_list"`
	ForbidList        string `json:"forbid_list"`
}

type Service_tcp struct {
	ID        int `gorm:"primary_key" json:"id"`
	ServiceId int `json:"service_id"`
	Port      int `json:"port"`
}

func (*Service_tcp) TableName() string {
	return "service_tcp"
}
