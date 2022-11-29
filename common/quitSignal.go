package common

import (
	"os"
	"os/signal"
	"syscall"
)

func QuitSignal() chan os.Signal {
	quit := make(chan os.Signal)
	//defer close(quit)  这里直接用于程序关闭了
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	return quit
}
