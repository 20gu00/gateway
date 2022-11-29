package common

//负载类型
const (
	LoadTypeHTTP = iota
	LoadTypeTCP
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

	DayFlowStatKey   = "day_flow_stat"
	RedisFlowHourKey = "flow_hour_count"

	FlowTotal             = "flow_total"    //全站流量统计,大盘显示
	ServiceFlowStatPrefix = "flow_service_" //服务流量统计
	FlowTenantPrefix      = "flow_tenant_"  //租户流量统计

	JwtSignKey = "my_sign_key" //jwt的签名的key
	JwtExpires = 60 * 60       //token的过期时间,默认一小时
)

var (
	LoadTypeMap = map[int]string{
		LoadTypeHTTP: "HTTP",
		LoadTypeTCP:  "TCP",
	}
)
