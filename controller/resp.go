package controller

//主要提供给swag显示返回值用(接口文档暂时不选择swag,另外生产环境一般也不会使用swag等接口文档)
type Response struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data string `json:"data"`
	//TraceId
}
