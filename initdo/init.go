package initdo

import (
	"github.com/20gu00/gateway/common"
	"github.com/20gu00/gateway/dao"
	"github.com/20gu00/gateway/model/manager"
)

func InitDo() {
	common.InitLogger()
	dao.InitMysql()
	dao.InitSessionRedis()                                           //session 统计器通用一个
	if err := manager.ServiceManagerHandler.LoadOnce(); err != nil { //nil不存在
		common.Logger.Infof("一次性加载数据到内存中失败", err.Error())
	}
}
