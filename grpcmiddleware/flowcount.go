package grpcmiddleware

import (
	"github.com/20gu00/gateway/common"
	"github.com/20gu00/gateway/dao"
	"google.golang.org/grpc"
	"log"
)

func GrpcFlowCountMiddleware(serviceDetail *dao.ServiceDetail) func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		totalCounter, err := common.FlowCounterHandler.GetCounter(common.FlowTotal)
		if err != nil {
			return err
		}
		totalCounter.Increase()
		serviceCounter, err := common.FlowCounterHandler.GetCounter(common.FlowServicePrefix + serviceDetail.ServiceInfo.ServiceName)
		if err != nil {
			return err
		}
		serviceCounter.Increase()

		if err := handler(srv, ss); err != nil {
			log.Printf("GrpcFlowCountMiddleware failed with error %v\n", err)
			return err
		}
		return nil
	}
}
