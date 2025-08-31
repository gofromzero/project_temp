# 监控基础设施部署指南

本目录包含多租户管理后台系统的监控基础设施配置文件。

## 架构概览

监控栈包括以下组件：
- **Prometheus**: 指标收集和存储
- **Grafana**: 可视化仪表板
- **Node Exporter**: 系统指标收集（可选）
- **MySQL & Redis**: 应用依赖服务

## 快速启动

### 1. 启动完整监控环境

```bash
# 从项目根目录执行
docker-compose -f docker-compose.monitoring.yaml up -d
```

### 2. 访问服务

- **应用后端**: http://localhost:8000
  - 健康检查: http://localhost:8000/health
  - 指标端点: http://localhost:8000/metrics
- **Prometheus**: http://localhost:9090
- **Grafana**: http://localhost:3000
  - 默认登录: admin / admin123
- **Node Exporter**: http://localhost:9100

### 3. 验证部署

```bash
# 检查所有服务状态
docker-compose -f docker-compose.monitoring.yaml ps

# 查看应用健康状态
curl http://localhost:8000/health

# 查看Prometheus指标
curl http://localhost:8000/metrics
```

## 配置文件说明

### Prometheus 配置

文件位置: `infrastructure/prometheus/prometheus.yml`

主要监控目标：
- **multi-tenant-backend**: 应用指标 (端口8000)
- **prometheus**: Prometheus自监控 (端口9090)  
- **node-exporter**: 系统指标 (端口9100)

### Grafana 配置

配置目录: `infrastructure/grafana/`

- **数据源**: 自动配置Prometheus数据源
- **仪表板**: 预配置应用监控仪表板
- **权限**: admin/admin123 (生产环境请修改)

预配置仪表板包含以下监控面板：
- HTTP请求速率和响应时间
- HTTP状态码分布
- 数据库连接数和查询性能
- Redis连接数和缓存命中率

## 生产环境配置

### 安全配置

1. **修改默认密码**:
```yaml
# docker-compose.monitoring.yaml
environment:
  - GF_SECURITY_ADMIN_PASSWORD=your_secure_password
```

2. **启用HTTPS**:
```yaml
# 在Grafana和Prometheus服务中配置SSL证书
volumes:
  - ./certs:/etc/ssl/certs
```

### 性能优化

1. **Prometheus数据保留**:
```yaml
# 在prometheus服务的command中调整保留时间
- '--storage.tsdb.retention.time=30d'  # 保留30天数据
```

2. **资源限制**:
```yaml
# 为各服务配置内存和CPU限制
deploy:
  resources:
    limits:
      memory: 1G
      cpus: '0.5'
```

## 监控指标说明

### HTTP指标
- `http_requests_total`: HTTP请求总数（按方法和路径分组）
- `http_request_duration_average`: 平均请求响应时间（毫秒）
- `http_response_status_total`: HTTP响应状态码统计

### 数据库指标
- `database_connections_active`: 当前活跃数据库连接数
- `database_queries_total`: 数据库查询总数
- `database_query_duration_average`: 平均数据库查询时间（毫秒）

### Redis指标
- `redis_connections_active`: 当前活跃Redis连接数
- `redis_cache_hits_total`: 缓存命中总数
- `redis_cache_misses_total`: 缓存未命中总数
- `redis_cache_hit_ratio`: 缓存命中率（0.0-1.0）

## 故障排除

### 常见问题

1. **Prometheus无法连接到应用**:
```bash
# 检查网络连接
docker exec -it multi-tenant-prometheus wget -qO- http://backend:8000/metrics
```

2. **Grafana无法加载仪表板**:
```bash
# 检查数据源配置
docker exec -it multi-tenant-grafana curl http://prometheus:9090/api/v1/query?query=up
```

3. **数据库连接失败**:
```bash
# 检查数据库服务状态
docker-compose -f docker-compose.monitoring.yaml logs mysql
```

### 日志查看

```bash
# 查看应用日志
docker-compose -f docker-compose.monitoring.yaml logs backend

# 查看Prometheus日志
docker-compose -f docker-compose.monitoring.yaml logs prometheus

# 查看Grafana日志
docker-compose -f docker-compose.monitoring.yaml logs grafana
```

## 扩展监控

### 添加额外导出器

1. **MySQL导出器**:
```yaml
mysql-exporter:
  image: prom/mysqld-exporter
  environment:
    - DATA_SOURCE_NAME=app:app123@tcp(mysql:3306)/
  ports:
    - "9104:9104"
```

2. **Redis导出器**:
```yaml
redis-exporter:
  image: oliver006/redis_exporter
  environment:
    - REDIS_ADDR=redis://redis:6379
  ports:
    - "9121:9121"
```

### 自定义告警规则

在`infrastructure/prometheus/`目录创建`alert_rules.yml`:

```yaml
groups:
  - name: multi-tenant-admin
    rules:
      - alert: HighResponseTime
        expr: http_request_duration_average > 1000
        for: 2m
        labels:
          severity: warning
        annotations:
          summary: "High response time detected"
```

## 备份和恢复

### Prometheus数据备份

```bash
# 创建数据备份
docker run --rm -v prometheus_data:/data -v $(pwd):/backup alpine \
  tar czf /backup/prometheus-backup.tar.gz /data
```

### Grafana配置备份

```bash
# 导出仪表板配置
curl -H "Authorization: Bearer your-api-key" \
  http://localhost:3000/api/dashboards/uid/multi-tenant-admin > dashboard-backup.json
```