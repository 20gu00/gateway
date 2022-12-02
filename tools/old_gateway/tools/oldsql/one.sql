-- MySQL dump 10.13  Distrib 5.7.38, for Linux (x86_64)
--
-- Host: localhost    Database: gateway
-- ------------------------------------------------------
-- Server version	5.7.38

/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!40101 SET NAMES utf8 */;
/*!40103 SET @OLD_TIME_ZONE=@@TIME_ZONE */;
/*!40103 SET TIME_ZONE='+00:00' */;
/*!40014 SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;

--
-- Table structure for table `admin`
--

DROP TABLE IF EXISTS `admin`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `admin` (
  `id` bigint(30) NOT NULL AUTO_INCREMENT COMMENT '自增id',
  `user_name` varchar(255) NOT NULL DEFAULT '' COMMENT '用户名称',
  `salt` varchar(50) NOT NULL DEFAULT '' COMMENT '加密用的盐',
  `password` varchar(255) NOT NULL DEFAULT '' COMMENT '密码',
  `create_at` datetime NOT NULL DEFAULT '1971-01-01 00:00:00' COMMENT '新增的时间',
  `update_at` datetime NOT NULL DEFAULT '1971-01-01 00:00:00' COMMENT '更新的时间',
  `is_delete` tinyint(5) NOT NULL DEFAULT '0' COMMENT '删除状态',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=3 DEFAULT CHARSET=utf8 COMMENT='admin表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `admin`
--

LOCK TABLES `admin` WRITE;
/*!40000 ALTER TABLE `admin` DISABLE KEYS */;
INSERT INTO `admin` VALUES (1,'admin','admin','f6da0c07372854658b17f13b696614989029773bd1457f68225d8b4339e48cec','1971-01-01 00:00:00','1971-01-01 00:00:00',0),(2,'cjq','cjq','e6dbc41d32d5a3f47a1c7216008b4ee5a7b7d2de84284e66e78','1971-01-01 00:00:00','1971-01-01 00:00:00',0);
/*!40000 ALTER TABLE `admin` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `service_access_control`
--

DROP TABLE IF EXISTS `service_access_control`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `service_access_control` (
  `id` bigint(30) NOT NULL AUTO_INCREMENT COMMENT '自增主键',
  `service_id` bigint(20) NOT NULL DEFAULT '0' COMMENT '服务id',
  `open_auth` tinyint(4) NOT NULL DEFAULT '0' COMMENT '是否开启权限',
  `black_list` varchar(1000) NOT NULL DEFAULT '' COMMENT '黑名单',
  `white_list` varchar(1000) NOT NULL DEFAULT '' COMMENT '白名单',
  `white_host_name` varchar(1000) NOT NULL DEFAULT '' COMMENT '白名单主机',
  `clientip_flow_limit` int(10) NOT NULL DEFAULT '0' COMMENT '客户端ip限流',
  `service_flow_limit` int(30) NOT NULL DEFAULT '0' COMMENT '服务端限流',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=194 DEFAULT CHARSET=utf8 COMMENT='网关权限表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `service_access_control`
--

LOCK TABLES `service_access_control` WRITE;
/*!40000 ALTER TABLE `service_access_control` DISABLE KEYS */;
INSERT INTO `service_access_control` VALUES (190,429,0,'','','',0,0),(191,430,0,'','','',0,0),(192,431,0,'','','',0,0),(193,432,0,'','','',0,0);
/*!40000 ALTER TABLE `service_access_control` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `service_grpc_rule`
--

DROP TABLE IF EXISTS `service_grpc_rule`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `service_grpc_rule` (
  `id` bigint(30) NOT NULL AUTO_INCREMENT COMMENT '自增主键',
  `service_id` bigint(20) NOT NULL DEFAULT '0' COMMENT '服务id',
  `port` int(5) NOT NULL DEFAULT '0' COMMENT '端口',
  `header_transfor` varchar(5000) NOT NULL DEFAULT '' COMMENT 'header转换支持add del edit',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=207 DEFAULT CHARSET=utf8 COMMENT='网关grpc路由表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `service_grpc_rule`
--

LOCK TABLES `service_grpc_rule` WRITE;
/*!40000 ALTER TABLE `service_grpc_rule` DISABLE KEYS */;
INSERT INTO `service_grpc_rule` VALUES (205,430,8020,''),(206,432,8090,'');
/*!40000 ALTER TABLE `service_grpc_rule` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `service_http_rule`
--

DROP TABLE IF EXISTS `service_http_rule`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `service_http_rule` (
  `id` bigint(30) NOT NULL AUTO_INCREMENT COMMENT '自增主键',
  `service_id` bigint(20) NOT NULL COMMENT '服务id',
  `rule_type` tinyint(4) NOT NULL DEFAULT '0' COMMENT '匹配类型',
  `rule` varchar(255) NOT NULL DEFAULT '' COMMENT 'type=domain表示域名，type=url_prefix时表示url前缀',
  `need_https` tinyint(4) NOT NULL DEFAULT '0' COMMENT '支持https 1=支持',
  `need_strip_uri` tinyint(4) NOT NULL DEFAULT '0' COMMENT '启用strip_uri 1=启用',
  `need_websocket` tinyint(4) NOT NULL DEFAULT '0' COMMENT '是否支持websocket 1=支持',
  `url_rewrite` varchar(5000) NOT NULL DEFAULT '' COMMENT 'url重写功能',
  `header_transfor` varchar(5000) NOT NULL DEFAULT '' COMMENT 'header转换支持add del edit',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=317 DEFAULT CHARSET=utf8 COMMENT='网关http路由表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `service_http_rule`
--

LOCK TABLES `service_http_rule` WRITE;
/*!40000 ALTER TABLE `service_http_rule` DISABLE KEYS */;
INSERT INTO `service_http_rule` VALUES (316,429,0,'/testaddhttp',0,0,0,'','');
/*!40000 ALTER TABLE `service_http_rule` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `service_info`
--

DROP TABLE IF EXISTS `service_info`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `service_info` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '自增主键',
  `load_type` tinyint(4) NOT NULL DEFAULT '0' COMMENT '负载类型 0=http 1=tcp 2=grpc',
  `service_name` varchar(255) NOT NULL DEFAULT '' COMMENT '服务名称 6-128 数字字母下划线',
  `service_desc` varchar(255) NOT NULL DEFAULT '' COMMENT '服务描述',
  `create_at` datetime NOT NULL DEFAULT '1971-01-01 00:00:00' COMMENT '添加时间',
  `update_at` datetime NOT NULL DEFAULT '1971-01-01 00:00:00' COMMENT '更新时间',
  `is_delete` tinyint(4) DEFAULT '0' COMMENT '是否删除 1=删除',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=433 DEFAULT CHARSET=utf8 COMMENT='网关基本信息表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `service_info`
--

LOCK TABLES `service_info` WRITE;
/*!40000 ALTER TABLE `service_info` DISABLE KEYS */;
INSERT INTO `service_info` VALUES (429,0,'testrhttp','update','2022-10-25 20:31:45','2022-10-25 20:17:23',0),(430,2,'testgrpc','editgrpc','2022-10-25 20:32:31','2022-10-25 20:19:47',0),(431,1,'testtcp','update','2022-10-25 21:16:46','2022-10-25 20:21:12',0),(432,2,'testgrpc2','editgrpc10','2022-10-25 21:12:00','2022-10-25 20:56:33',0);
/*!40000 ALTER TABLE `service_info` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `service_load_balance`
--

DROP TABLE IF EXISTS `service_load_balance`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `service_load_balance` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '自增主键',
  `service_id` bigint(20) NOT NULL DEFAULT '0' COMMENT '服务id',
  `check_method` tinyint(20) NOT NULL DEFAULT '0' COMMENT '检查方法 0=tcpchk,检测端口是否握手成功',
  `check_timeout` int(10) NOT NULL DEFAULT '0' COMMENT 'check超时时间,单位s',
  `check_interval` int(11) NOT NULL DEFAULT '0' COMMENT '检查间隔, 单位s',
  `round_type` tinyint(4) NOT NULL DEFAULT '2' COMMENT '轮询方式 0=random 1=round-robin 2=weight_round-robin 3=ip_hash',
  `ip_list` varchar(2000) NOT NULL DEFAULT '' COMMENT 'ip列表',
  `weight_list` varchar(2000) NOT NULL DEFAULT '' COMMENT '权重列表',
  `forbid_list` varchar(2000) NOT NULL DEFAULT '' COMMENT '禁用ip列表',
  `upstream_connect_timeout` int(11) NOT NULL DEFAULT '0' COMMENT '建立连接超时, 单位s',
  `upstream_header_timeout` int(11) NOT NULL DEFAULT '0' COMMENT '获取header超时, 单位s',
  `upstream_idle_timeout` int(10) NOT NULL DEFAULT '0' COMMENT '链接最大空闲时间, 单位s',
  `upstream_max_idle` int(11) NOT NULL DEFAULT '0' COMMENT '最大空闲链接数',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=530 DEFAULT CHARSET=utf8 COMMENT='网关负载表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `service_load_balance`
--

LOCK TABLES `service_load_balance` WRITE;
/*!40000 ALTER TABLE `service_load_balance` DISABLE KEYS */;
INSERT INTO `service_load_balance` VALUES (526,429,0,0,0,0,'127.0.0.1:80','50','',0,0,0,0),(527,430,0,0,0,0,'127.0.0.1:1001','200','',0,0,0,0),(528,431,0,0,0,0,'127.0.0.1:1000','900','',0,0,0,0),(529,432,0,0,0,0,'127.0.0.1:1001','200','',0,0,0,0);
/*!40000 ALTER TABLE `service_load_balance` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `service_tcp_rule`
--

DROP TABLE IF EXISTS `service_tcp_rule`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `service_tcp_rule` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '自增主键',
  `service_id` bigint(20) NOT NULL COMMENT '服务id',
  `port` int(5) NOT NULL DEFAULT '0' COMMENT '端口号',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=1012 DEFAULT CHARSET=utf8 COMMENT='网关路由匹配表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `service_tcp_rule`
--

LOCK TABLES `service_tcp_rule` WRITE;
/*!40000 ALTER TABLE `service_tcp_rule` DISABLE KEYS */;
INSERT INTO `service_tcp_rule` VALUES (1011,431,8021);
/*!40000 ALTER TABLE `service_tcp_rule` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `tenant`
--

DROP TABLE IF EXISTS `tenant`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `tenant` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '自增id',
  `app_id` varchar(255) NOT NULL DEFAULT '' COMMENT '租户id',
  `name` varchar(255) NOT NULL DEFAULT '' COMMENT '租户名称',
  `secret` varchar(255) NOT NULL DEFAULT '' COMMENT '密钥',
  `white_ips` varchar(1000) NOT NULL DEFAULT '' COMMENT 'ip白名单',
  `qpd` bigint(20) NOT NULL DEFAULT '0' COMMENT '日请求量限制',
  `qps` bigint(20) NOT NULL DEFAULT '0' COMMENT '每秒请求量限制',
  `create_at` datetime NOT NULL COMMENT '添加时间',
  `update_at` datetime NOT NULL COMMENT '更新时间',
  `is_delete` tinyint(5) NOT NULL DEFAULT '0' COMMENT '是否删除',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=35 DEFAULT CHARSET=utf8 COMMENT='网关的租户表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `tenant`
--

LOCK TABLES `tenant` WRITE;
/*!40000 ALTER TABLE `tenant` DISABLE KEYS */;
INSERT INTO `tenant` VALUES (11,'tenant_id_a','租户1','449441eb5e72dca9c42a12f3924ea3a2','white_ips',100000,100,'2022-10-21 20:55:02','2022-10-22 07:23:34',0),(12,'tenant_id_b','租户2','8d7b11ec9be0e59a36b52f32366c09cb','',200,0,'2022-10-21 21:40:52','2022-10-22 07:23:27',0),(13,'tenant_id_c','租户3','','',0,0,'2022-10-21 22:02:23','2022-10-22 22:06:51',1),(14,'tenant_id_d','租户4','07d980f8a49347523ee1d5c1c41aec02','',0,0,'2022-10-21 22:06:38','2022-10-22 22:06:49',1);
/*!40000 ALTER TABLE `tenant` ENABLE KEYS */;
UNLOCK TABLES;
/*!40103 SET TIME_ZONE=@OLD_TIME_ZONE */;

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;

-- Dump completed on 2022-10-26  0:07:33
