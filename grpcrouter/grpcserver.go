package grpcrouter

import (
	"fmt"
	"github.com/20gu00/gateway/dao"
	"github.com/20gu00/gateway/grpcmiddleware"
	"github.com/20gu00/gateway/reverse_proxy"
	"github.com/e421083458/grpc-proxy/proxy"
	"google.golang.org/grpc"
	"log"
	"net"
)

var grpcServerList = []*warpGrpcServer{}

type warpGrpcServer struct {
	Addr string
	*grpc.Server
}

func GrpcServerRun() {
	serviceList := dao.ServiceManagerHandler.GetGrpcServiceList()
	for _, serviceItem := range serviceList {
		tempItem := serviceItem
		go func(serviceDetail *dao.ServiceDetail) {
			addr := fmt.Sprintf(":%d", serviceDetail.GRPCRule.Port)
			rb, err := dao.LoadBalancerHandler.GetLoadBalancer(serviceDetail)
			if err != nil {
				log.Fatalf(" [INFO] GetTcpLoadBalancer %v err:%v\n", addr, err)
				return
			}
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
				grpc.UnknownServiceHandler(grpcHandler))

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
