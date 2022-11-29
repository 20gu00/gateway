package dao

import (
	"fmt"
	"github.com/20gu00/gateway/common"
	"github.com/20gu00/gateway/model"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"os"
)

var DB *gorm.DB

func InitMysql() {
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

	DB = db
	common.Logger.Infof("初始化数据库成功")
	if err := DB.AutoMigrate(
		model.AccessControl{},
		model.LoadBalance{},
		model.Admin{},
		model.Tenant{},
		model.Service_http{},
		model.Service_tcp{},
		model.ServiceInfo{},
	); err != nil {
		common.Logger.Infof("创建表失败", err.Error())
		return
	}
}
