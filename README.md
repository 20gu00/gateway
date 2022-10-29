# gateway  
go实现的网关项目，分网关和后台管理两大部分，后端项目  


## 项目介绍  
分为两大部分，**网关**和网关的**后台管理系统**  

**网关**  
功能点：   
1.反向代理(支持http和https,tcp)  
2.负载均衡(支持随机，轮询，加权轮询，一致性hash)  
3. header头转换  
4.strip_uri  
5.url重写  
6.ip白名单和黑名单控制  
7.流量统计  
8. 漏桶限流控制  
10. jwt认证  
         
**后台管理系统**  
功能点：  
1.admin的登录退出，修改密码，信息获取  
2. http,https,tcp服务的增删改查  
3. 每个服务的流量统计和近两日的流量对比  
4.网关的租户的增删改查和租户的流量统计及其近两日的流量对比  
5.主页的服务和租户的流量统计  
          
## k8s运行
我提供了脚本，运行生成镜像：  
./build-image.sh  

部署到k8s中:  
kubectl apply -f gateway.yaml  

## 本地运行  
go run main.go --config ./conf/dev/ --kind market  
go run main.go --config ./conf/dev/ --kind proxy  
或者  
./bin/gateway --config ./conf/dev --kind market  
./bin/gateway --config ./conf/dev --kind proxy  

## 数据库准备
开发测试环境建议直接本地部署mysql和redis即可，mysql创建**gaterway**数据库，导入数据库文件./tools/new.sql  
如果项目是跑在k8s上，mysql和redis照样可以使用本地环境的，配置好通信即可  

提供了k8s部署mysql和redis和数据持久化的方案：  
yaml文件在./tools/databases下，直接apply部署即可  

## 个人预计近期优化
1.结合k8s，采用operator模式，实现真正的云原生  
2.使用grpc拆分成微服务架构  

## 数据库和工具
mysql redis swagger  

## 简单使用介绍  
本地环境：  
         后台管理服务: go run main.go --config ./conf/dev --kind market  
         网关服务: go run main.go --config ./conf/dev/ --kind proxy  
