package controller

import (
	"github.com/20gu00/gateway/common"
	"github.com/20gu00/gateway/dao"
	"github.com/20gu00/gateway/model"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"time"
)

func TenantListHandler(ctx *gin.Context) {
	p := new(model.TenantListInput)
	if err := ctx.ShouldBindJSON(p); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code": 2000,
			"msg":  "输入的请求参数不正确",
			"data": err.Error(),
		})
		return
	}
	db := dao.DB
	info := &model.Tenant{}
	list, total, err := info.TenantList(db, p)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code": 2001,
			"msg":  "按页查询tenant失败",
			"data": err.Error(),
		})
		return
	}

	outputList := []model.TenantListItemOutput{}
	for _, item := range list {
		appCounter, err := common.FlowCounterHandler.GetCounter(common.FlowTenantPrefix + item.AppId)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"code": 2002,
				"msg":  "获取流量统计器失败(用于统计租户流量)",
				"data": err.Error(),
			})
			//ctx.Abort()
			return
		}

		outputList = append(outputList, model.TenantListItemOutput{
			ID:       int64(item.ID),
			AppID:    item.AppId,
			Name:     item.Name,
			Secret:   item.Secret,
			WhiteIPS: item.WhiteIps,
			Qpd:      int64(item.Qpd),
			Qps:      int64(item.Qps),
			RealQpd:  appCounter.TotalCount,
			RealQps:  appCounter.QPS,
		})
	}
	output := model.TenantListOutput{
		List:  outputList,
		Total: total,
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "按页",
		"data": output,
	})
}

func TenantDetailHandler(ctx *gin.Context) {
	params := new(model.TenantDetailInput)
	if err := ctx.ShouldBindJSON(params); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code": 2000,
			"msg":  "输入的请求参数不正确",
			"data": err.Error(),
		})
		return
	}

	db := dao.DB
	search := &model.Tenant{
		Model: gorm.Model{
			ID: uint(params.ID),
		},
	}
	//查询某条记录最好使用first,find如果没有相应的记录会将查询条件插入结构体并返回,不会报错
	err := db.Table(search.TableName()).Where(search).First(search) //row .RowAffect
	if err.Error != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code": 2001,
			"msg":  "获取tenant的相信信息失败",
			"data": err.Error,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "成功获取tenant的详细信息",
		"data": search,
	})
}

func TenantDeleteHandler(c *gin.Context) {
	params := new(model.TenantDetailInput)
	if err := c.ShouldBindJSON(params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 2000,
			"msg":  "输入的请求参数不正确",
			"data": err.Error(),
		})
		return
	}

	db := dao.DB
	search := &model.Tenant{
		Model: gorm.Model{
			ID: uint(params.ID),
		},
	}
	err := db.Where(search).Find(search)
	if err.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 2001,
			"msg":  "未找到该tenant",
			"data": err.Error,
		})
		return
	}

	search.IsDelete = 1
	if tx := db.Model(search).Where("id=?", uint(params.ID)).Updates(search); tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 2002,
			"msg":  "更新tenant为删除状态失败",
			"data": tx.Error,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "删除网关的租户成功",
	})
	return
}

func TenantAddHandler(c *gin.Context) {
	params := new(model.TenantAddHttpInput)
	if err := c.ShouldBindJSON(params); err != nil { //会判断参数的类型是否合适
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 2000,
			"msg":  "输入的请求参数不正确",
			"data": err.Error(),
		})
		return
	}

	db := dao.DB
	//验证app_id是否被占用
	search := &model.Tenant{
		AppId: params.AppID,
	}
	err := db.Where(search).Find(search)
	if err == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 2001,
			"msg":  "该租户id被占用了",
			"data": err.Error,
		})
		return
	}

	if params.Secret == "" {
		//生成secret
		params.Secret = common.MD5(params.AppID)
	}

	info := &model.Tenant{
		AppId:    params.AppID,
		Name:     params.Name,
		Secret:   params.Secret,
		WhiteIps: params.WhiteIPS,
		Qps:      int(params.Qps),
		Qpd:      int(params.Qpd),
	}
	if tx := db.Save(info); tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 2002,
			"msg":  "创建租户失败",
			"data": err.Error,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "创建租户成功",
		"data": err.Error,
	})
	return
}

func TenantUpdateHandler(c *gin.Context) {
	params := new(model.TenantUpdateInput)
	if err := c.ShouldBindJSON(params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 2000,
			"msg":  "输入的请求参数不正确",
			"data": err.Error(),
		})
		return
	}

	db := dao.DB
	search := &model.Tenant{
		Model: gorm.Model{
			ID: uint(params.ID),
		},
	}
	err := db.Where(search).Find(search)
	if err.Error != nil { //tx一般不为空,要判断的是tx.Error
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 2001,
			"msg":  "未找到该tenant",
			"data": err.Error,
		})
		return
	}

	if params.Secret == "" {
		params.Secret = common.MD5(params.AppID)
	}

	search.Name = params.Name
	search.Secret = params.Secret
	search.WhiteIps = params.WhiteIPS
	search.Qps = int(params.Qps)
	search.Qpd = int(params.Qpd)
	if tx := db.Model(search).Where("id=?", params.ID).Updates(search); tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 2002,
			"msg":  "更新tenant失败",
			"data": err.Error,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "更新tenant信息成功",
		"data": err.Error,
	})
	return
}

func TenantStatHandler(c *gin.Context) {
	params := new(model.TenantDetailInput)
	if err := c.ShouldBindJSON(params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 2000,
			"msg":  "输入的请求参数不正确",
			"data": err.Error(),
		})
		return
	}

	db := dao.DB
	search := &model.Tenant{
		Model: gorm.Model{
			ID: uint(params.ID),
		},
	}

	tx := db.Where(search).Find(search)
	if tx.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 2001,
			"msg":  "未找到该tenant",
			"data": tx.Error,
		})
		return
	}

	//今日流量全天 小时粒度访问统计
	todayStat := []int64{}
	counter, err := common.FlowCounterHandler.GetCounter(common.FlowTenantPrefix + search.AppId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 2002,
			"msg":  "创建流量统计器失败(统计租户)",
			"data": err.Error(),
		})

		//c.Abort()
		return
	}
	currentTime := time.Now()
	timeLocation, _ := time.LoadLocation("Asia/Shanghai")
	for i := 0; i <= time.Now().In(timeLocation).Hour(); i++ {
		dateTime := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), i, 0, 0, 0, timeLocation)
		hourData, _ := counter.GetHourData(dateTime)
		todayStat = append(todayStat, hourData)
	}

	//昨日流量全天小时级访问统计
	yesterdayStat := []int64{}
	yesterTime := currentTime.Add(-1 * time.Duration(time.Hour*24))
	for i := 0; i <= 23; i++ {
		dateTime := time.Date(yesterTime.Year(), yesterTime.Month(), yesterTime.Day(), i, 0, 0, 0, timeLocation)
		hourData, _ := counter.GetHourData(dateTime)
		yesterdayStat = append(yesterdayStat, hourData)
	}
	stat := model.StatOutput{
		Today:     todayStat,
		Yesterday: yesterdayStat,
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "统计租户的流量成功",
		"data": stat,
	})

	return
}
