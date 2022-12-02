package common

import (
	"os"
	"os/signal"
	"syscall"
)

//func QuitSignal() chan os.Signal {
//	quit := make(chan os.Signal, 2) //缓冲2
//	//defer close(quit)  这里直接用于程序关闭了
//	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
//	return quit
//}
func SignalHandler() chan struct{} { //<-chan struct{}
	stop := make(chan struct{})
	c := make(chan os.Signal, 2)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-c
		close(stop)
		<-c
		os.Exit(1)
	}()
	return stop //返回一个通道,该通道可以被关闭
}
