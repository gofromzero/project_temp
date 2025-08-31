# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 项目概览

这是一个多租户系统管理后台项目，采用现代化全栈架构：
- **前端**: React + TypeScript + Ant Design + Tailwind CSS
- **后端**: Go + GoFrame框架 + DDD领域驱动设计
- **数据**: MySQL + Redis
- **架构**: Monorepo结构，前后端分离

## 技术栈

### 后端 (Go + GoFrame)
- **语言**: Go ^1.21
- **框架**: GoFrame ^2.5 (企业级Go开发框架)
- **数据库**: MySQL ^8.0 + Redis ^7.0
- **认证**: JWT + RBAC权限模型
- **测试**: Go Test + Testify

### 前端 (React + TypeScript)
- **语言**: TypeScript ^5.0
- **框架**: React ^18.0
- **UI库**: Ant Design ^5.0 + Tailwind CSS ^3.0
- **状态管理**: Zustand ^4.0
- **构建工具**: Vite ^5.0
- **测试**: Vitest + Testing Library

## 项目结构

```
multi-tenant-admin/
├── frontends/                 # 前端应用包
│   ├── apps/admin-web/       # 管理后台前端
│   └── packages/             # 共享包
├── backend/                  # 后端应用 (DDD架构)
│   ├── api/                 # API控制器层
│   ├── domain/              # 领域层
│   ├── application/         # 应用服务层
│   ├── repository/          # 数据访问层
│   ├── infr/                # 基础设施层
│   └── pkg/                 # 共享包
├── infrastructure/          # Docker/K8s配置
└── docs/                   # 项目文档
```

## 常用开发命令

### 环境设置
```bash
# 初始化项目依赖
npm install                    # 安装前端依赖
go mod tidy                    # 整理Go依赖

# 启动开发环境
npm run dev                    # 启动所有服务
```

### 前端开发
```bash
cd frontends/apps/admin-web
npm run dev                    # 启动前端开发服务器
npm run build                  # 构建生产版本
npm run test                   # 运行前端测试
npm run lint                   # 代码格式检查
```

### 后端开发
```bash
cd backend
go run cmd/main.go serve       # 启动后端服务
go run cmd/main.go migrate     # 数据库迁移
go run cmd/main.go seed        # 初始化数据
go test ./...                  # 运行后端测试
```

### 数据库操作
```bash
# 启动依赖服务
docker-compose up -d mysql redis

# 数据库迁移和初始化
cd backend
go run cmd/main.go migrate
go run cmd/main.go seed
```

## 核心架构原则

### 后端架构 (DDD)
- **API层** (`api/`): HTTP处理器、中间件、路由
- **领域层** (`domain/`): 业务实体和领域逻辑
- **应用层** (`application/`): 业务用例编排
- **仓储层** (`repository/`): 数据访问抽象
- **基础设施层** (`infr/`): 外部服务集成

### 前端架构
- **组件化**: 基于Ant Design的可复用UI组件
- **状态管理**: 使用Zustand进行轻量级状态管理
- **样式系统**: Ant Design + Tailwind CSS原子化样式
- **类型安全**: 全面的TypeScript类型定义

### 多租户设计
- **数据隔离**: 通过租户ID实现完全数据隔离
- **权限控制**: RBAC角色权限模型，支持租户级和系统级权限
- **身份认证**: JWT Token无状态认证

## 编码规范

### Go后端
- 遵循Go标准编码规范
- 使用GoFrame框架约定
- DDD领域模型组织
- 接口抽象和依赖注入

### TypeScript前端
- 严格的TypeScript类型检查
- React Hooks和函数式组件
- Ant Design组件优先
- Tailwind CSS原子化样式

## 环境配置

### 必需环境变量
```bash
# 数据库配置
DB_HOST=localhost
DB_PORT=3306  
DB_USER=root
DB_PASSWORD=your_password
DB_NAME=multi_tenant_admin

# Redis配置
REDIS_HOST=localhost
REDIS_PORT=6379

# JWT配置
JWT_SECRET=your_jwt_secret_key
JWT_EXPIRE_HOURS=24

# 前端配置
VITE_API_URL=http://localhost:8000/v1
```

## 语言设置
- **默认语言**: 中文
- **专业术语**: 保持英文原文（如技术名词、框架名称、编程概念等）
- **代码注释**: 使用中文解释，但保留英文技术术语
- **文件名、变量名、函数名**: 保持英文命名规范