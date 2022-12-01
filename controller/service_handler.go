package controller

import (
	"fmt"
	"github.com/20gu00/gateway/common"
	"github.com/20gu00/gateway/dao"
	"github.com/20gu00/gateway/model"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"net/http"
	"os"
	"strings"
	"time"
)

func ServiceListHandler(c *gin.Context) {
	p := new(model.ServiceListInput)
	//ShouldBind有时候板顶后结构体参数依旧为空
	if err := c.ShouldBindJSON(p); err != nil { //会绑定相应的字段,多余字段不管,如果有不合法输入比如/一般会报错,但是输入不存在的字段不会报错,所以这里的设计并不能在后端很好地实现参数校验,应该额外设计参数校验(没有绑定的字段如果可以为空那么就默认为空)
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 2000,
			"msg":  "输入的请求参数不正确",
			"data": err.Error(),
		})
		return
	}

	//指针
	serviceInfo := new(model.ServiceInfo)
	serviceInfoList, total, err := serviceInfo.PageList(dao.DB, p)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 2001,
			"msg":  "分页查询serviceInfo失败",
			"data": err.Error(),
		})
		return
	}

	//outList := new([]model.ServiceListItemOutput)
	outList := []model.ServiceListItemOutput{}
	//根据serviceInfo获取serviceDetail
	for _, item := range serviceInfoList {
		serviceDetail, err := item.ServiceDetail(dao.DB, &item)
		if err != nil {
			common.Logger.Infof("通过serviceInfo获取serviceDetail失败" + err.Error())
			return
		}

		workDir, err := os.Getwd()
		if err != nil {
			common.Logger.Infof("获取工作目录失败")
			return
		}

		viper.SetConfigName("general")
		viper.SetConfigType("yaml")
		viper.AddConfigPath(workDir + "/conf")
		if err := viper.ReadInConfig(); err != nil {
			common.Logger.Infof("general配置文件读取失败(service_list)", err.Error())
			return
		}
		//添加服务时就需要填写这些字段(这里获取通过用户的输入来确定serviceAddr)
		//判断http的接入方式,前缀方式:clusterIP:clusterPort+Path
		serviceAddr := "unknwonAddr"
		//网关ip,集群ip,入口
		clusterIp := viper.GetString("cluster.cluster_ip")
		clusterPort := viper.GetString("cluster.cluster_port") //网关提供代理服务的端口不是后台管理系统的端口
		clusterSslPort := viper.GetString("cluster.cluster_ssl_port")

		//类型http,接入方式前缀
		//是否支持(开启)https 1支持 使用网关的sslPort
		if serviceDetail.BaseInfo.LoadType == common.LoadTypeHTTP && serviceDetail.Http.RuleType == common.HTTPRuleTypePrefixURL && serviceDetail.Http.NeedHttps == 1 {
			serviceAddr = fmt.Sprintf("%s:%s%s", clusterIp, clusterSslPort, serviceDetail.Http.Rule)
		}

		if serviceDetail.BaseInfo.LoadType == common.LoadTypeHTTP && serviceDetail.Http.RuleType == common.HTTPRuleTypePrefixURL && serviceDetail.Http.NeedHttps == 0 {
			serviceAddr = fmt.Sprintf("%s:%s%s", clusterIp, clusterPort, serviceDetail.Http.Rule) //Rule:/user
		}

		//接入方式域名(没有ssl说法)
		if serviceDetail.BaseInfo.LoadType == common.LoadTypeHTTP && serviceDetail.Http.RuleType == common.HTTPRuleTypeDomain {
			serviceAddr = serviceDetail.Http.Rule //Rule:www.test.com
		}

		//tcp(添加tcp服务时的serviceAddr)
		//if serviceDetail.BaseInfo.LoadType == common.LoadTypeTCP {
		//	serviceAddr = fmt.Sprintf("%s:%d", clusterIp, serviceDetail.Tcp.Port) //tcp端口
		//}

		//获取实际的工作负载(服务地址)(创建服务时的ip列表)
		ipList := serviceDetail.LoadBalance.GetIpList()

		//流量统计,统计这个服务的流量
		//服务流量统计前缀+服务名称 service_ + service_name
		//创建一个流量统计器
		counter, err := common.FlowCounterHandler.GetCounter(common.ServiceFlowStatPrefix + item.ServiceName)
		if err != nil {
			common.Logger.Infof("创建流量统计器失败", err.Error())
		}

		outItem := model.ServiceListItemOutput{
			ID:          int64(item.ID),
			LoadType:    item.LoadType,
			ServiceName: item.ServiceName,
			ServiceDesc: item.ServiceDesc,
			ServiceAddr: serviceAddr,
			Qps:         counter.QPS,
			Qpd:         counter.TotalCount,
			TotalNode:   len(ipList), //服务地址,工作负载数目
		}

		outList = append(outList, outItem)
	}

	out := &model.ServiceListOutput{
		Total:       int(total),
		ServiceList: outList,
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "列出服务的列表,按页查询,每一页数目统计",
		"data": out,
	})
}

func ServiceDeleteHandler(c *gin.Context) {
	p := new(model.ServiceDeleteInput)
	if err := c.ShouldBindJSON(p); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 2000,
			"msg":  "输入的请求参数不正确",
			"data": err.Error(),
		})
		return
	}

	serviceInfo := &model.ServiceInfo{
		Model: gorm.Model{ //匿名字段,调用时可以直接调用它的成员但是赋值时是给这个结构体赋值,这时候类型名(结构提名)视为成员名
			ID: uint(p.ID),
		},
	}
	serviceInfo, err := serviceInfo.Find(c, dao.DB, serviceInfo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 2001,
			"msg":  "未知服务",
		})
		return
	}

	serviceInfo.IsDelete = 1 //软删除
	if tx := dao.DB.Model(serviceInfo).Where("id =? ", p.ID).Update("is_delete", 1); tx.Error != nil {
		c.JSON(http.StatusForbidden, gin.H{
			"msg": tx.Error,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "删除service成功",
	})
}

func ServiceDetailHandler(c *gin.Context) {
	p := new(model.ServiceDetailInput)
	if err := c.ShouldBindJSON(p); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 2000,
			"msg":  "输入的请求参数不正确",
			"data": err.Error(),
		})
		return
	}

	serviceInfo := &model.ServiceInfo{
		Model: gorm.Model{ //匿名字段,调用时可以直接调用它的成员但是赋值时是给这个结构体赋值,这时候类型名(结构提名)视为成员名
			ID: uint(p.ID),
		},
	}
	serviceInfo, err := serviceInfo.Find(c, dao.DB, serviceInfo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 2001,
			"msg":  "未知服务",
		})
		return
	}

	//获取该service的详细的信息
	serviceDetail, err := serviceInfo.ServiceDetail(dao.DB, serviceInfo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 2002,
			"msg":  "获取该服务的详细信息失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  serviceDetail,
	})
}

func ServiceStatHandler(c *gin.Context) {
	p := new(model.ServiceStatInput)
	if err := c.ShouldBindJSON(p); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 2000,
			"msg":  "输入的请求参数不正确",
			"data": err.Error(),
		})
	}

	serviceInfo := &model.ServiceInfo{
		Model: gorm.Model{
			ID: uint(p.ID),
		},
	}

	serviceDetail, err := serviceInfo.ServiceDetail(dao.DB, serviceInfo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 2001,
			"msg":  "获取该服务的详细信息失败",
			"data": err.Error(),
		})
		return
	}

	//服务流量统计器命名:服务流量统计前缀+服务名称
	counter, err := common.FlowCounterHandler.GetCounter(common.ServiceFlowStatPrefix + serviceDetail.BaseInfo.ServiceName)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 2002,
			"msg":  "创建服务的统计器失败",
			"data": err.Error(),
		})
	}

	//当天的流量统计,小时为粒度
	todayList := []int64{}
	currentTime := time.Now()

	//today
	for h := 0; h <= currentTime.Hour(); h++ {
		dateTime := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), h, 0, 0, 0, common.TimeLocation)
		hourData, _ := counter.GetHourData(dateTime) //这个小时的数据
		todayList = append(todayList, hourData)      //追加到当天的数据列表中,按小时为粒度进行统计
	}

	//昨天的数据
	yesterdayList := []int64{}
	yesterTime := currentTime.Add(-1 * time.Duration(time.Hour*24))

	for h := 0; h <= 23; h++ {
		dateTime := time.Date(yesterTime.Year(), yesterTime.Month(), yesterTime.Day(), h, 0, 0, 0, common.TimeLocation) //时区在配置加载是初始化为"Asia/Shanghai"
		hourData, _ := counter.GetHourData(dateTime)
		yesterdayList = append(yesterdayList, hourData)
	}

	//serviceList:=make(map[string][]int64)  slice chan map
	serviceStatList := map[string][]int64{
		"today":     todayList,
		"yesterday": yesterdayList,
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "service的今天和昨天的流量统计",
		"data": serviceStatList,
	})
}

func ServiceAddHttpHandler(c *gin.Context) {
	p := new(model.ServiceAddHttpInput)
	if err := c.ShouldBindJSON(p); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 2000,
			"msg":  "输入的请求参数不正确",
			"data": err.Error(),
		})
		return
	}

	//实际工作负载ip和权重
	if len(strings.Split(p.IpList, ",")) != len(strings.Split(p.WeightList, ",")) {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":        2001,
			"msg":         "IP列表与权重列表数量不一致",
			"data_ipList": strings.Split(p.IpList, ","),
			"data_Weight": strings.Split(p.WeightList, ","),
		})
		return
	}

	//这类型的添加,设计到多张表,往往组要servicedetail,修改多张表需要开启事务,这里使用本地事务
	//db := dao.DB  那么其他也是用这个连接的数据库操作就会出错,只是这里开启了事务
	//获取当前工作目录,一般是项目目录(go.mod)
	workDir, err := os.Getwd()
	if err != nil {
		common.Logger.Infof("获取工作目录失败")
		return
	}

	//读取配置文件
	viper.SetConfigName("mysql")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(workDir + "/conf") //可多个
	if err := viper.ReadInConfig(); err != nil {
		common.Logger.Infof("mysql配置文件读取失败", err.Error())
		return
	}

	//使用配置文件
	dsn := fmt.Sprintf(viper.GetString("mysql.sourceName"))
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Println(dsn)
		common.Logger.Infof("连接mysql失败")
		return
	}
	db.Begin()

	serviceInfo := &model.ServiceInfo{
		ServiceName: p.ServiceName,
	}

	//服务名称冲突校验
	//query
	if row := db.Where(serviceInfo).Where("is_delete = 0").First(serviceInfo).RowsAffected; row == 1 {
		db.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 2002,
			"msg":  "服务已存在,该服务名称已经被使用",
			"data": row,
		})
		return
	}

	httpSearch := &model.Service_http{
		RuleType: p.RuleType,
		Rule:     p.Rule,
	}
	//判断接入类型(前缀或者域名)和接入的前缀(路径)或域名是否冲突
	if row := db.Where(serviceInfo).Where("is_delete = 0").Where(httpSearch).First(httpSearch).RowsAffected; row == 1 {
		db.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 2003,
			"msg":  "服务的接入前缀或域名已存在", //网关统一代理全部后端实际的服务
			"data": row,
		})
		return
	}

	serviceInfoDao := &model.ServiceInfo{
		ServiceName: p.ServiceName,
		ServiceDesc: p.ServiceDesc,
	}
	if tx := db.Save(serviceInfoDao); tx.Error != nil {
		db.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 2004,
			"msg":  "新增服务,添加进基本服务表失败",
			"data": tx.Error,
		})
		return
	}

	httpDao := &model.Service_http{
		ServiceId:     int(serviceInfoDao.ID),
		RuleType:      p.RuleType,
		Rule:          p.Rule,
		NeedHttps:     p.NeedHttps,
		NeedStrip_uri: p.NeedStripUri,
		//NeedWebsocket:  p.NeedWebsocket,
		UrlRewrite:     p.UrlRewrite,
		HeaderTransfor: p.HeaderTransfor,
	}
	if tx := db.Save(httpDao); tx.Error != nil {
		db.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 2005,
			"msg":  "新增服务,添加进http表失败",
			"data": tx.Error,
		})
		return
	}

	accessControlDao := &model.AccessControl{
		ServiceId:         int(serviceInfoDao.ID),
		OpenAuth:          p.OpenAuth,
		BlackList:         p.BlackList,
		WhiteList:         p.WhiteList,
		ClientIPFlowLimit: p.ClientipFlowLimit,
		ServiceFlowLimit:  p.ServiceFlowLimit,
	}
	if tx := db.Save(accessControlDao); tx.Error != nil {
		db.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 2006,
			"msg":  "新增服务,添加进accessControl表失败",
			"data": tx.Error,
		})
		return
	}

	loadbalanceDao := &model.LoadBalance{
		ServiceId:  int(serviceInfoDao.ID),
		RoundType:  p.RoundType,
		IpList:     p.IpList,
		WeightList: p.WeightList,
		//网关的上游的,往往就是指服务端(信息流来看待,传出数据的地方一般成为上游)
		UpstreamConnectTimeOut: p.UpstreamConnectTimeout,
		UpstreamHeaderTimeOut:  p.UpstreamHeaderTimeout,
		UpstreamIdleTimeOut:    p.UpstreamIdleTimeout,
		UpstreamMaxIdle:        p.UpstreamMaxIdle,
	}
	if tx := db.Save(loadbalanceDao); tx.Error != nil {
		db.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 2007,
			"msg":  "新增服务,添加进loadBalance表失败",
			"data": tx.Error,
		})
		return
	}

	db.Commit()
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "新增http服务成功",
	})
}

func ServiceUpdateHttpHandler(c *gin.Context) {
	p := new(model.ServiceUpdateHttpInput)
	if err := c.ShouldBindJSON(p); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 2000,
			"msg":  "输入的请求参数不正确",
			"data": err.Error(),
		})
	}

	if len(strings.Split(p.IpList, ",")) != len(strings.Split(p.WeightList, ",")) {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":        2001,
			"msg":         "IP列表与权重列表数量不一致",
			"data_ipList": strings.Split(p.IpList, ","),
			"data_Weight": strings.Split(p.WeightList, ","),
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
			"code": 2002,
			"msg":  "服务不存在",
			"data": tx.Error,
		})
		return
	}

	serviceDetail, err := serviceInfo.ServiceDetail(db, serviceInfo)
	if err != nil {
		db.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 2003,
			"msg":  "获取该服务的详情失败",
			"data": db.Error,
		})
		return
	}

	serviceInfoDao := serviceDetail.BaseInfo
	serviceInfoDao.ServiceDesc = p.ServiceDesc
	if tx := db.Save(serviceInfoDao); tx.Error != nil {
		db.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 2004,
			"msg":  "写入serviceInfo表失败",
			"data": tx.Error,
		})
		return
	}

	httpRuleDao := serviceDetail.Http
	httpRuleDao.NeedHttps = p.NeedHttps
	httpRuleDao.NeedStrip_uri = p.NeedStripUri
	//httpRuleDao.NeedWebsocket = p.NeedWebsocket
	httpRuleDao.UrlRewrite = p.UrlRewrite
	httpRuleDao.HeaderTransfor = p.HeaderTransfor
	if tx := db.Save(httpRuleDao); tx.Error != nil {
		db.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 2005,
			"msg":  "写入http表失败",
			"data": tx.Error,
		})
		return
	}

	accessControlDao := serviceDetail.AccessControl
	accessControlDao.OpenAuth = p.OpenAuth
	accessControlDao.BlackList = p.BlackList
	accessControlDao.WhiteList = p.WhiteList
	accessControlDao.ClientIPFlowLimit = p.ClientipFlowLimit
	accessControlDao.ServiceFlowLimit = p.ServiceFlowLimit
	if tx := db.Save(httpRuleDao); tx.Error != nil {
		db.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 2006,
			"msg":  "写入accessControl表失败",
			"data": tx.Error,
		})
		return
	}

	loadbalanceDao := serviceDetail.LoadBalance
	loadbalanceDao.RoundType = p.RoundType
	loadbalanceDao.IpList = p.IpList
	loadbalanceDao.WeightList = p.WeightList
	loadbalanceDao.UpstreamHeaderTimeOut = p.UpstreamHeaderTimeout
	loadbalanceDao.UpstreamIdleTimeOut = p.UpstreamIdleTimeout
	loadbalanceDao.UpstreamMaxIdle = p.UpstreamMaxIdle
	if tx := db.Save(httpRuleDao); tx.Error != nil {
		db.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 2007,
			"msg":  "写入loadBalance表失败",
			"data": tx.Error,
		})
		return
	}

	db.Commit()
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "修改http服务成功",
	})
}
