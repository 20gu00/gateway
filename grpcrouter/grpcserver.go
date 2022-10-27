package grpcrouter

import (
	"fmt"
	"github.com/20gu00/gateway/common/grpc-proxy/proxy"
	"github.com/20gu00/gateway/dao"
	"github.com/20gu00/gateway/grpcmiddleware"
	"github.com/20gu00/gateway/reverse_proxy"
	"google.golang.org/grpc"
	"log"
	"net"
)

var grpcServerList = []*warpGrpcServer{}

type warpGrpcServer struct {
	Addr string
	*grpc.Server
}

//port方式接入 客户端主动探测的服务发现
func GrpcServerRun() {
	serviceList := dao.ServiceManagerHandler.GetGrpcServiceList()
	for _, serviceItem := range serviceList {
		tempItem := serviceItem
		go func(serviceDetail *dao.ServiceDetail) {
			addr := fmt.Sprintf(":%d", serviceDetail.GRPCRule.Port)
			//负载均衡器
			rb, err := dao.LoadBalancerHandler.GetLoadBalancer(serviceDetail)
			if err != nil {
				log.Fatalf(" [INFO] GetTcpLoadBalancer %v err:%v\n", addr, err)
				return
			}
			//连接监听器
			lis, err := net.Listen("tcp", addr)
			if err != nil {
				log.Fatalf(" [INFO] GrpcListen %v err:%v\n", addr, err)
			}
			grpcHandler := reverse_proxy.NewGrpcLoadBalanceHandler(rb)
			s := grpc.NewServer(
				grpc.ChainStreamInterceptor(
					grpcmiddleware.GrpcFlowCountMiddleware(serviceDetail),
					grpcmiddleware.GrpcFlowLimitMiddleware(serviceDetail),
					grpcmiddleware.GrpcJwtAuthTokenMiddleware(serviceDetail),
					grpcmiddleware.GrpcJwtFlowCountMiddleware(serviceDetail),
					grpcmiddleware.GrpcJwtFlowLimitMiddleware(serviceDetail),
					grpcmiddleware.GrpcWhiteListMiddleware(serviceDetail),
					grpcmiddleware.GrpcBlackListMiddleware(serviceDetail),
					grpcmiddleware.GrpcHeaderTransferMiddleware(serviceDetail),
				),
				grpc.CustomCodec(proxy.Codec()),
				grpc.UnknownServiceHandler(grpcHandler)) //grpc没有设置任何回调,直接回调这个handler

			grpcServerList = append(grpcServerList, &warpGrpcServer{
				Addr:   addr,
				Server: s,
			})
			log.Printf(" [INFO] grpc_proxy_run %v\n", addr)
			if err := s.Serve(lis); err != nil {
				log.Fatalf(" [INFO] grpc_proxy_run %v err:%v\n", addr, err)
			}
		}(tempItem)
	}
}

func GrpcServerStop() {
	for _, grpcServer := range grpcServerList {
		grpcServer.GracefulStop()
		log.Printf(" [INFO] grpc_proxy_stop %v stopped\n", grpcServer.Addr)
	}
}
