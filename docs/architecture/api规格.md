# API规格

## REST API规格

```yaml
openapi: 3.0.0
info:
  title: 多租户系统管理后台 API
  version: 1.0.0
  description: 支持多租户的系统管理后台RESTful API接口
servers:
  - url: https://api.example.com/v1
    description: 生产环境
  - url: http://localhost:8000/v1  
    description: 开发环境

paths:
  # 认证相关
  /auth/login:
    post:
      summary: 用户登录
      tags: [Authentication]
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              properties:
                email:
                  type: string
                password:
                  type: string
                tenantCode:
                  type: string
                  description: 租户代码（系统管理员可为空）
      responses:
        '200':
          description: 登录成功
          content:
            application/json:
              schema:
                type: object
                properties:
                  token:
                    type: string
                  user:
                    $ref: '#/components/schemas/User'
                    
  # 租户管理
  /tenants:
    get:
      summary: 获取租户列表
      tags: [Tenants]
      security:
        - BearerAuth: []
      parameters:
        - name: page
          in: query
          schema:
            type: integer
        - name: limit
          in: query
          schema:
            type: integer
        - name: status
          in: query
          schema:
            type: string
            enum: [active, suspended, disabled]
      responses:
        '200':
          description: 租户列表
          content:
            application/json:
              schema:
                type: object
                properties:
                  data:
                    type: array
                    items:
                      $ref: '#/components/schemas/Tenant'
                  pagination:
                    $ref: '#/components/schemas/Pagination'
    post:
      summary: 创建租户
      tags: [Tenants]
      security:
        - BearerAuth: []
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateTenantRequest'
      responses:
        '201':
          description: 租户创建成功
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Tenant'

  /tenants/{tenantId}:
    get:
      summary: 获取租户详情
      tags: [Tenants]
      security:
        - BearerAuth: []
      parameters:
        - name: tenantId
          in: path
          required: true
          schema:
            type: string
      responses:
        '200':
          description: 租户详情
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Tenant'
    put:
      summary: 更新租户
      tags: [Tenants]  
      security:
        - BearerAuth: []
      parameters:
        - name: tenantId
          in: path
          required: true
          schema:
            type: string
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/UpdateTenantRequest'
      responses:
        '200':
          description: 更新成功
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Tenant'

  # 用户管理  
  /users:
    get:
      summary: 获取用户列表
      tags: [Users]
      security:
        - BearerAuth: []
      parameters:
        - name: tenantId
          in: query
          schema:
            type: string
          description: 租户ID（系统管理员可查看所有）
        - name: page
          in: query
          schema:
            type: integer
        - name: limit
          in: query
          schema:
            type: integer
      responses:
        '200':
          description: 用户列表
          content:
            application/json:
              schema:
                type: object
                properties:
                  data:
                    type: array
                    items:
                      $ref: '#/components/schemas/User'
                  pagination:
                    $ref: '#/components/schemas/Pagination'
    post:
      summary: 创建用户
      tags: [Users]
      security:
        - BearerAuth: []
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateUserRequest'
      responses:
        '201':
          description: 用户创建成功
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/User'

components:
  securitySchemes:
    BearerAuth:
      type: http
      scheme: bearer
      bearerFormat: JWT
      
  schemas:
    Tenant:
      type: object
      properties:
        id:
          type: string
        name:
          type: string
        code:
          type: string
        status:
          type: string
          enum: [active, suspended, disabled]
        config:
          type: object
        createdAt:
          type: string
          format: date-time
        updatedAt:
          type: string
          format: date-time
          
    User:
      type: object
      properties:
        id:
          type: string
        tenantId:
          type: string
          nullable: true
        username:
          type: string
        email:
          type: string
        profile:
          type: object
        status:
          type: string
          enum: [active, inactive, locked]
        roles:
          type: array
          items:
            $ref: '#/components/schemas/Role'
        createdAt:
          type: string
          format: date-time
          
    Role:
      type: object
      properties:
        id:
          type: string
        name:
          type: string
        code:
          type: string
        permissions:
          type: array
          items:
            $ref: '#/components/schemas/Permission'
            
    Permission:
      type: object
      properties:
        id:
          type: string
        name:
          type: string
        code:
          type: string
        resource:
          type: string
        action:
          type: string
        scope:
          type: string
          enum: [system, tenant, self]
          
    Pagination:
      type: object
      properties:
        page:
          type: integer
        limit:
          type: integer
        total:
          type: integer
        pages:
          type: integer
          
    CreateTenantRequest:
      type: object
      required: [name, code]
      properties:
        name:
          type: string
        code:
          type: string
        config:
          type: object
        adminUser:
          type: object
          properties:
            username:
              type: string
            email:
              type: string
            password:
              type: string
              
    UpdateTenantRequest:
      type: object
      properties:
        name:
          type: string
        status:
          type: string
          enum: [active, suspended, disabled]
        config:
          type: object
          
    CreateUserRequest:
      type: object
      required: [username, email, password]
      properties:
        tenantId:
          type: string
        username:
          type: string
        email:
          type: string
        password:
          type: string
        profile:
          type: object
        roleIds:
          type: array
          items:
            type: string
```

---
