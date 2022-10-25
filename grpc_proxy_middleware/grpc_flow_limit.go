package grpc_proxy_middleware

import (
	"fmt"
	"github.com/20gu00/gateway/common"
	"github.com/20gu00/gateway/dao"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
	"log"
	"strings"
)

func GrpcFlowLimitMiddleware(serviceDetail *dao.ServiceDetail) func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if serviceDetail.AccessControl.ServiceFlowLimit != 0 {
			serviceLimiter, err := common.FlowLimiterHandler.GetLimiter(
				common.FlowServicePrefix+serviceDetail.ServiceInfo.ServiceName,
				float64(serviceDetail.AccessControl.ServiceFlowLimit))
			if err != nil {
				return err
			}
			//Allow()判断当前是否能拿到token,拿不到那就是被限流了  Wait()阻塞等待拿到token  Reserve()返回个预估的等待时间,再去取token  (拿到token也就是说该请求可以进入可以被处理,没有被限速限流)
			if !serviceLimiter.Allow() {
				return errors.New(fmt.Sprintf("service flow limit %v", serviceDetail.AccessControl.ServiceFlowLimit))
			}
		}
		peerCtx, ok := peer.FromContext(ss.Context())
		if !ok {
			return errors.New("peer not found with context")
		}
		peerAddr := peerCtx.Addr.String()
		addrPos := strings.LastIndex(peerAddr, ":")
		clientIP := peerAddr[0:addrPos]
		if serviceDetail.AccessControl.ClientIPFlowLimit > 0 {
			clientLimiter, err := common.FlowLimiterHandler.GetLimiter(
				common.FlowServicePrefix+serviceDetail.ServiceInfo.ServiceName+"_"+clientIP,
				float64(serviceDetail.AccessControl.ClientIPFlowLimit))
			if err != nil {
				return err
			}
			if !clientLimiter.Allow() {
				return errors.New(fmt.Sprintf("%v flow limit %v", clientIP, serviceDetail.AccessControl.ClientIPFlowLimit))
			}
		}
		if err := handler(srv, ss); err != nil {
			log.Printf("GrpcFlowLimitMiddleware failed with error %v\n", err)
			return err
		}
		return nil
	}
}
