# gateway
go开发的网关项目，分网关和网关的后台管理系统两大部分


分为两大部分，网关和网关的后台管理系统  

## 网关  
功能点：  
1.反向代理   
2.负载均衡(支持随机，轮询，加权轮询，一致性hash)  
3.header头转换  
4.strip_uri  
5.url重写  
6.ip白名单和黑名单控制  
7.流量统计  
8.漏桶限流控制  
9.jwt认证  


## 后台管理系统  
功能点：  
1.admin的登录退出，修改密码，信息获取  
2.http,https服务的增删改查  
3.每个服务的流量统计和近两日的流量对比  
4.网关的租户的增删改查和租户的流量统计及其近两日的流量对比  
5.主页的服务和租户的流量统计  
