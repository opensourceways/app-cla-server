# app-cla-server 项目介绍

## 项目概述

`app-cla-server` 是一个 **CLA（Contributor License Agreement，贡献者许可协议）签署服务系统**。该系统用于管理开源项目的CLA签署流程，支持个人贡献者和企业贡献者的CLA签署，为开源社区提供完善的贡献者协议管理解决方案。

## 技术栈

### 核心技术
- **编程语言**: Go 1.24.1
- **Web框架**: Beego v2.3.6
- **数据库**: MongoDB
- **缓存**: Redis (go-redis/v8)
- **PDF生成**: gofpdf
- **邮件服务**: gomail.v2
- **OAuth2认证**: golang.org/x/oauth2
- **API文档**: Swagger

### 支持的代码平台
- GitHub
- Gitee

## 核心功能模块

### 1. CLA管理
- CLA文档的创建、配置和管理
- CLA模板管理
- CLA版本控制

### 2. 签署功能
- **个人签署（Individual Signing）**: 个人贡献者签署CLA
- **企业签署（Corporation Signing）**: 企业签署CLA，支持企业管理员管理
- **员工签署（Employee Signing）**: 企业员工签署CLA，需要企业管理员审批

### 3. 企业管理
- 企业管理员注册和管理
- 企业邮箱域名验证
- 员工管理（添加、删除、审批）
- 企业签署记录查看

### 4. 认证与授权
- OAuth2认证（支持GitHub、Gitee）
- 用户Token管理
- 访问控制和权限验证
- 密码找回功能

### 5. 邮件服务
- SMTP配置和管理
- 验证码发送
- 签署完成通知
- PDF签署文档邮件发送

### 6. PDF生成
- 签署后的CLA文档PDF生成
- 企业签署PDF文档生成
- PDF模板配置

### 7. 其他功能
- 数据库迁移
- 服务健康检查（心跳检测）
- 隐私政策管理
- 数据收集声明

## 项目结构

```
app-cla-server/
├── controllers/              # HTTP控制器层
│   ├── cla.go                # CLA管理接口
│   ├── individual_signing.go # 个人签署接口
│   ├── corp_signing.go       # 企业签署接口
│   ├── employee_signing.go   # 员工签署接口
│   ├── corp_manager.go       # 企业管理接口
│   ├── employee_manager.go   # 员工管理接口
│   ├── verification_code.go  # 验证码接口
│   ├── smtp.go               # SMTP配置接口
│   └── ...
│
├── models/                   # 数据模型层
│   ├── cla.go                # CLA模型
│   ├── corp_signing.go       # 企业签署模型
│   ├── individual_signing.go # 个人签署模型
│   ├── employee_signing.go   # 员工签署模型
│   └── ...
│
├── signing/                  # 签署业务逻辑（DDD架构）
│   ├── domain/               # 领域层
│   │   ├── dp/               # 领域原语
│   │   └── ...               # 领域服务
│   └── infrastructure/       # 基础设施层
│       ├── repositoryimpl/   # 仓储实现
│       ├── smtpimpl/         # 邮件服务实现
│       └── ...
│
├── common/                   # 通用基础设施
│   ├── infrastructure/
│   │   ├── mongodb/          # MongoDB数据库连接和操作
│   │   └── redisdb/          # Redis缓存连接和操作
│   └── ...
│
├── pdf/                      # PDF生成模块
│   ├── pdf_generator.go      # PDF生成器
│   └── corp_signing_pdf.go   # 企业签署PDF
│
├── oauth2/                   # OAuth2认证
│   └── oauth2.go
│
├── code-platform-auth/       # 代码平台认证
│   ├── platforms/
│   │   ├── github.go         # GitHub OAuth2认证
│   │   └── gitee.go          # Gitee OAuth2认证
│   └── ...
│
├── worker/                   # 后台任务
│   ├── worker.go             # Worker管理
│   └── corp_pdf_email.go     # 企业PDF邮件发送任务
│
├── config/                   # 配置管理
│   └── config.go
│
├── routers/                  # 路由配置
│   └── router.go
│
├── util/                     # 工具类
│   ├── util.go
│   ├── httpclient.go
│   └── ...
│
├── swagger/                  # Swagger API文档
│   ├── swagger.json
│   └── swagger.yml
│
├── deploy/                   # 部署配置
│   ├── app.conf
│   └── app.conf.yaml
│
├── conf/privacy/             # 隐私政策文档
│   ├── clasign_privacy_policy_zh.md
│   └── ...
│
├── main.go                   # 应用程序入口
├── go.mod                    # Go模块依赖
└── Dockerfile                # Docker构建文件
```

## 架构设计

### 分层架构
项目采用清晰的分层架构：

1. **Controller层**（controllers/）: 处理HTTP请求，参数验证，调用业务逻辑
2. **Model层**（models/）: 数据模型定义和数据库操作
3. **Domain层**（signing/domain/）: 业务领域逻辑，采用DDD设计
4. **Infrastructure层**（signing/infrastructure/）: 基础设施实现，如仓储、邮件、加密等

### 领域驱动设计（DDD）
在`signing`模块中采用了DDD架构：
- **Domain层**: 包含领域原语（dp）、领域服务、领域对象
- **Infrastructure层**: 包含仓储实现、邮件服务实现、加密服务实现等

### 数据存储
- **MongoDB**: 主要数据存储，存储签署记录、用户信息、CLA配置等
- **Redis**: 用于缓存、会话管理、访问令牌存储

### API设计
- 采用RESTful API设计风格
- API版本管理（/v1）
- Swagger自动生成API文档

## 配置说明

### 主要配置项
- **PDF配置**: PDF生成相关配置
- **API配置**: API服务相关配置
- **SMTP配置**: 邮件服务器配置
- **MongoDB配置**: 数据库连接配置
- **Redis配置**: 缓存服务配置
- **代码平台配置**: GitHub/Gitee OAuth2配置
- **密码配置**: 密码加密配置
- **对称加密配置**: 敏感数据加密配置

### 配置文件
配置文件采用YAML格式，启动时通过`-config-file`参数指定配置文件路径。

## 部署与运行

### 本地运行
```bash
# 生成路由
bee generate routers

# 启动服务
go run main.go -config-file <config-file-path>
```

### Docker部署
项目提供了Dockerfile，支持容器化部署。

## API端点示例

### CLA管理
- `GET /v1/cla`: 获取CLA信息
- `POST /v1/cla`: 创建CLA

### 个人签署
- `POST /v1/individual-signing`: 个人签署CLA
- `GET /v1/individual-signing`: 查询签署记录

### 企业签署
- `POST /v1/corporation-signing`: 企业签署CLA
- `GET /v1/corporation-signing`: 查询企业签署记录

### 企业管理
- `POST /v1/corporation-manager`: 注册企业管理员
- `GET /v1/corporation-manager`: 获取企业管理员信息

### 员工签署
- `POST /v1/employee-signing`: 员工签署CLA
- `GET /v1/employee-signing`: 查询员工签署记录

### 员工管理
- `POST /v1/employee-manager`: 添加员工
- `DELETE /v1/employee-manager`: 删除员工

### 认证相关
- `GET /v1/auth/authcodeurl/{platform}/{purpose}`: 获取OAuth2认证URL
- `GET /v1/auth/{platform}/{purpose}`: OAuth2回调接口

## 开发指南

### 测试
项目包含单元测试文件（`*_test.go`），测试覆盖率目标为80%。

### API文档生成
使用Beego的Swagger工具自动生成API文档，开发模式下可通过`/swagger`访问。

### 代码规范
- 遵循Go语言编码规范
- 采用有意义的包和函数命名
- 保持代码简洁清晰

## 许可证

Apache 2.0 License

## 相关资源

- 项目地址: github.com/opensourceways/app-cla-server
- Web框架: [Beego](https://github.com/beego/beego)
- 数据库: [MongoDB](https://www.mongodb.com/)
- 缓存: [Redis](https://redis.io/)