package common

//负载类型
const (
	LoadTypeHTTP = iota
	LoadTypeTCP
	LoadTypeGRPC
)

//接入方式
const (
	HTTPRuleTypePrefixURL = iota
	HTTPRuleTypeDomain
)
const (
	ValidatorKey  = "ValidatorKey"
	TranslatorKey = "TranslatorKey"
	SessionKey    = "SessionKey"

	RedisFlowDayKey  = "flow_day_count"
	RedisFlowHourKey = "flow_hour_count"

	FlowTotal         = "flow_total"
	FlowServicePrefix = "flow_service_"
	FlowAppPrefix     = "flow_app_"

	JwtSignKey = "my_sign_key" //jwt的签名的key
	JwtExpires = 60 * 60       //token的过期时间,默认一小时
)

var (
	LoadTypeMap = map[int]string{
		LoadTypeHTTP: "HTTP",
		LoadTypeTCP:  "TCP",
		LoadTypeGRPC: "GRPC",
	}
)
