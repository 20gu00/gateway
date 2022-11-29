package initdo

import (
	"github.com/20gu00/gateway/common"
	"github.com/20gu00/gateway/docs"
	"github.com/spf13/viper"
	"os"
)

func InitSwag() {
	//docs do
	workDir, err := os.Getwd()
	if err != nil {
		common.Logger.Infof("获取工作目录失败")
		return
	}

	//读取配置文件
	viper.SetConfigName("general")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(workDir + "/conf") //可多个
	if err := viper.ReadInConfig(); err != nil {
		common.Logger.Infof("general配置文件读取失败(swagger)", err.Error())
		return
	}

	docs.SwaggerInfo.Title = viper.GetString("swagger.title")
	docs.SwaggerInfo.Description = viper.GetString("swagger.desc")
	docs.SwaggerInfo.Version = "1.0" //文档版本
	docs.SwaggerInfo.Host = viper.GetString("swagger.host")
	docs.SwaggerInfo.BasePath = viper.GetString("swagger.base_path")
	docs.SwaggerInfo.Schemes = []string{"http", "https"} //有没有tls

}
