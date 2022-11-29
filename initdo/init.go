package initdo

import (
	"github.com/20gu00/gateway/common"
	"github.com/20gu00/gateway/dao"
)

func InitDo() {
	common.InitLogger()
	dao.InitMysql()
	dao.InitSessionRedis() //session 统计器通用一个
}
