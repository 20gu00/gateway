package common

import (
	"github.com/natefinch/lumberjack"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"os"
	"path"
)

//初始化logrus不只是gin的请求,整个项目都可用(gin请求也可以单独通过中间件方式使用)
var Logger *logrus.Logger

func InitLogger() {
	workDir, err := os.Getwd()
	if err != nil {
		logrus.Infof("获取工作目录失败")
	}

	viper.SetConfigName("log")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(workDir + "/conf")
	if err := viper.ReadInConfig(); err != nil {
		Logger.Infof("读取log配置文件失败")
	}

	Logger = logrus.New()

	kind := viper.GetString("kind.output")
	level := viper.GetString("logConf.level")

	var logLevel logrus.Level
	switch {
	case level == "trace":
		logLevel = logrus.TraceLevel
	case level == "debug":
		logLevel = logrus.DebugLevel
	case level == "info":
		logLevel = logrus.InfoLevel
	case level == "warn":
		logLevel = logrus.WarnLevel
	case level == "error":
		logLevel = logrus.ErrorLevel
	case level == "fatal":
		logLevel = logrus.FatalLevel
	case level == "panic":
		logLevel = logrus.PanicLevel
	}

	//默认logrus.TextFormatter{}
	//Logger.SetFormatter(&logrus.JSONFormatter{})
	Logger.SetFormatter(&logrus.TextFormatter{})

	Logger.SetLevel(logLevel)

	switch {
	case kind == "console":
		Logger.SetOutput(os.Stdout)
	case kind == "file":
		logOutputFile(Logger)
	case kind == "consoleAndFile":
		go Logger.SetOutput(os.Stdout)
		go logOutputFile(Logger)
	default:
		logrus.Infof("log的配置文件配置kind不符合格式")
	}

	Logger.Infof("logger配置完成")
}

func logOutputFile(logger *logrus.Logger) {
	workDir, err := os.Getwd()
	if err != nil {
		logrus.Infof("获取工作目录失败")
	}

	viper.SetConfigName("log")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(workDir + "/conf")
	if err := viper.ReadInConfig(); err != nil {
		logger.Infof("读取log配置文件失败")
	}

	logFile := path.Join(viper.GetString("logConf.logPath"), viper.GetString("logConf.logFile"))
	logConf := &lumberjack.Logger{
		Filename:   logFile, //如果不是代码中比如os.Open等操作打开文件,记得设置好文件的权限
		MaxSize:    viper.GetInt("logConf.maxSize"),
		MaxBackups: viper.GetInt("logConf.maxBackups"),
		MaxAge:     viper.GetInt("logConf.maxAge"),
		Compress:   viper.GetBool("logConf.compress"),
	}
	//fmt.Println(logConf) //直接打印出&{},也可以使用*解引用
	logger.SetOutput(logConf)
	return
}

//有些日志是gin框架的日志,照样自主控制输出,这边控制不了
