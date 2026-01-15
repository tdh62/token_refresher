# JWT Token Refresher

一个功能强大的JWT token自动刷新工具，使用Go语言编写，支持自定义刷新方法和token提取规则。

## 功能特性

- **灵活的刷新配置**: 通过配置文件模板定义HTTP请求，支持变量替换
- **JSONPath提取**: 使用JSONPath表达式从响应中提取token
- **自动刷新**: 智能调度器自动在token过期前刷新
- **SQLite存储**: 轻量级数据库存储项目配置和刷新日志
- **Web管理界面**: 简洁美观的Web界面，方便管理和查看
- **多项目支持**: 同时管理多个不同的token刷新项目
- **刷新日志**: 详细记录每次刷新的结果和错误信息

## 快速开始

### 安装依赖

```bash
go mod tidy
```

### 配置

在运行程序前，需要配置用户名和密码。有两种方式：

**方式1：使用配置文件（推荐）**

```bash
cp config.yaml.example config.yaml
# 编辑 config.yaml，设置用户名和密码
```

**方式2：使用环境变量**

```bash
export USERNAME=admin
export PASSWORD=your-secure-password
```

### 运行程序

```bash
go run main.go
```

程序将在 `http://localhost:3007` 启动Web服务。

### 构建二进制文件

```bash
go build -o jwt_refresher
```

Windows:
```bash
go build -o jwt_refresher.exe
```

## 使用说明

### 1. 访问Web界面

打开浏览器访问 `http://localhost:3007`

浏览器会提示输入用户名和密码（HTTP Basic Auth）。

### 2. 创建项目

点击"新建项目"按钮，填写以下信息:

#### 基本信息
- **项目名称**: 项目的唯一标识
- **描述**: 项目说明（可选）

#### 刷新配置
- **刷新URL**: Token刷新接口地址
- **请求方法**: HTTP方法（POST/GET/PUT）
- **请求头**: JSON格式的HTTP headers
- **请求体模板**: 支持变量替换的请求体模板

#### Token提取规则
- **Access Token路径**: JSONPath表达式，用于提取access token
- **Refresh Token路径**: JSONPath表达式，用于提取refresh token
- **过期时间路径**: JSONPath表达式，用于提取过期时间（秒）

#### 凭证信息
- **Client ID**: 客户端ID
- **Client Secret**: 客户端密钥
- **初始Refresh Token**: 用于首次刷新的refresh token

#### 刷新策略
- **提前刷新时间**: 在token过期前多少秒开始刷新（默认300秒）

### 3. 模板变量

请求体模板支持以下变量:

- `{{.ClientId}}` - 替换为项目的Client ID
- `{{.ClientSecret}}` - 替换为项目的Client Secret
- `{{.RefreshToken}}` - 替换为当前的Refresh Token

### 4. AWS OIDC示例

以下是AWS OIDC的完整配置示例:

**刷新URL**:
```
https://oidc.us-east-1.amazonaws.com/token
```

**请求方法**: `POST`

**请求头**:
```json
{
  "Content-Type": "application/json"
}
```

**请求体模板**:
```json
{
  "clientId": "{{.ClientId}}",
  "clientSecret": "{{.ClientSecret}}",
  "grantType": "refresh_token",
  "refreshToken": "{{.RefreshToken}}"
}
```

**Token提取规则**:
- Access Token路径: `accessToken`
- Refresh Token路径: `refreshToken`
- 过期时间路径: `expiresIn`

**凭证信息**:
- Client ID: 你的AWS客户端ID
- Client Secret: 你的AWS客户端密钥
- 初始Refresh Token: 你的初始refresh token

## API接口

**注意：所有API接口都需要HTTP Basic Auth认证。**

### 项目管理

- `GET /api/projects` - 获取所有项目
- `GET /api/projects/:id` - 获取项目详情
- `POST /api/projects` - 创建项目
- `PUT /api/projects/:id` - 更新项目
- `DELETE /api/projects/:id` - 删除项目
- `POST /api/projects/:id/toggle` - 启用/禁用项目
- `POST /api/projects/:id/refresh` - 手动触发刷新

### Token查询

- `GET /api/projects/:id/token` - 获取当前有效token
- `GET /api/projects/:id/logs` - 获取刷新日志

### 示例

获取token:
```bash
curl -u admin:password http://localhost:3007/api/projects/1/token
```

手动刷新:
```bash
curl -u admin:password -X POST http://localhost:3007/api/projects/1/refresh
```

或使用Authorization头:
```bash
curl -H "Authorization: Basic $(echo -n 'admin:password' | base64)" \
  http://localhost:3007/api/projects/1/token
```

## 配置

JWT Refresher支持两种配置方式：配置文件和环境变量。

### 配置文件（推荐）

创建 `config.yaml` 文件：

```yaml
# 服务端口（默认: 3007）
port: 3007

# 数据目录，用于存储数据库和日志（默认: ./data）
data_dir: ./data

# Basic Auth认证凭据（必需）
username: admin
password: your-secure-password

# 日志文件名（相对于data_dir，默认: app.log）
log_file: app.log
```

### 环境变量

环境变量会覆盖配置文件中的设置：

- `PORT` - Web服务端口（默认: 3007）
- `DATA_DIR` - 数据目录路径（默认: ./data）
- `USERNAME` - 认证用户名（必需）
- `PASSWORD` - 认证密码（必需）
- `LOG_FILE` - 日志文件名（默认: app.log）

### 配置优先级

1. 环境变量（最高优先级）
2. 配置文件（config.yaml）
3. 默认值（最低优先级）

### 示例

使用环境变量运行：

```bash
export USERNAME=admin
export PASSWORD=secret123
export PORT=3007
./jwt_refresher
```

使用配置文件运行：

```bash
# 创建配置文件
cp config.yaml.example config.yaml
# 编辑配置文件后运行
./jwt_refresher
```

### 文件结构

```
./
├── jwt_refresher          # 应用程序二进制文件
├── config.yaml            # 配置文件（可选）
└── data/                  # 数据目录（自动创建）
    ├── jwt_refresher.db   # SQLite数据库
    └── app.log            # 应用日志
```

所有持久化数据都存储在 `./data` 目录中，便于备份和Docker卷挂载。

## 项目结构

```
jwt_refresher/
├── main.go                 # 程序入口
├── go.mod                  # Go模块定义
├── config/
│   └── config.go          # 配置加载
├── models/
│   ├── project.go         # 项目数据模型
│   └── refresh_log.go     # 刷新日志模型
├── database/
│   └── db.go              # 数据库操作
├── refresher/
│   ├── engine.go          # 刷新引擎核心逻辑
│   ├── template.go        # 请求模板解析
│   └── extractor.go       # JSONPath token提取
├── scheduler/
│   └── scheduler.go       # 定时调度器
├── api/
│   ├── router.go          # API路由
│   ├── project.go         # 项目管理API
│   └── token.go           # Token查询API
└── web/
    └── static/
        ├── index.html     # 管理界面
        └── app.js         # 前端逻辑
```

## 技术栈

- **语言**: Go 1.21+
- **数据库**: SQLite3
- **Web框架**: Gin
- **JSONPath**: gjson
- **前端**: HTML + Tailwind CSS + Vanilla JavaScript

## 工作原理

1. **调度器**: 每分钟检查一次所有启用的项目
2. **刷新判断**: 如果token即将过期（在`refresh_before_seconds`秒内），触发刷新
3. **HTTP请求**: 使用配置的URL、方法、headers和body模板构建请求
4. **变量替换**: 将模板中的变量替换为实际值
5. **发送请求**: 发送HTTP请求到刷新接口
6. **提取Token**: 使用JSONPath从响应中提取新的token
7. **更新数据库**: 保存新的token和过期时间
8. **记录日志**: 记录刷新结果（成功/失败）

## 安全建议

- **认证保护**: 所有API和Web界面都需要认证，请设置强密码
- **配置文件权限**: 如果使用配置文件存储密码，建议设置文件权限为600（仅所有者可读写）
- **HTTPS**: 在生产环境中使用HTTPS，避免密码在网络传输中被窃取
- **定期备份**: 定期备份 `./data` 目录中的数据库文件
- **环境变量**: 在生产环境中，推荐使用环境变量而非配置文件存储敏感信息
- **版本控制**: 不要将包含真实密码的 `config.yaml` 提交到版本控制系统

## Docker部署

### 使用Docker Compose（推荐）

项目包含 `docker-compose.yml` 文件，可以快速部署：

```bash
# 1. 创建配置文件
cp config.yaml.example config.yaml
# 编辑 config.yaml，设置用户名和密码

# 2. 启动服务
docker-compose up -d

# 3. 查看日志
docker-compose logs -f

# 4. 停止服务
docker-compose down
```

### Docker Compose配置说明

```yaml
version: '3.8'
services:
  jwt_refresher:
    build: .
    ports:
      - "3007:3007"
    volumes:
      # 配置文件单文件挂载（只读）
      - ./config.yaml:/app/config.yaml:ro
      # 数据目录卷挂载（持久化）
      - jwt_data:/app/data
    environment:
      # 可选：使用环境变量覆盖配置
      - USERNAME=${USERNAME:-admin}
      - PASSWORD=${PASSWORD:-changeme}
    restart: unless-stopped

volumes:
  jwt_data:
    driver: local
```

### 卷挂载说明

- **配置文件**: 单文件挂载 `./config.yaml:/app/config.yaml:ro`（只读）
- **数据目录**: 卷挂载 `jwt_data:/app/data`（持久化存储数据库和日志）

### 环境变量部署

也可以完全使用环境变量，不需要配置文件：

```bash
docker run -d \
  -p 3007:3007 \
  -e USERNAME=admin \
  -e PASSWORD=your-secure-password \
  -v jwt_data:/app/data \
  --name jwt_refresher \
  jwt_refresher:latest
```

## 从旧版本迁移

如果你已经在使用旧版本的JWT Refresher，升级到新版本时需要注意以下变化：

### 重大变更

1. **端口变更**: 默认端口从 `8080` 改为 `3007`
2. **数据库位置**: 数据库从根目录移动到 `./data/jwt_refresher.db`
3. **认证要求**: 现在所有接口都需要HTTP Basic Auth认证
4. **配置方式**: 新增YAML配置文件支持

### 自动迁移

程序会自动迁移现有数据库：

1. 首次运行新版本时，程序会检查根目录是否存在 `jwt_refresher.db`
2. 如果存在，会自动移动到 `./data/jwt_refresher.db`
3. 所有数据都会保留，无需手动操作

### 手动迁移（可选）

如果需要手动迁移：

```bash
# 创建数据目录
mkdir -p ./data

# 移动数据库文件
mv jwt_refresher.db ./data/

# 创建配置文件
cp config.yaml.example config.yaml
# 编辑 config.yaml，设置用户名和密码

# 运行新版本
./jwt_refresher
```

### 更新API调用

如果你的脚本或程序调用了API接口，需要添加认证：

**旧版本（无认证）:**
```bash
curl http://localhost:8080/api/projects/1/token
```

**新版本（需要认证）:**
```bash
curl -u admin:password http://localhost:3007/api/projects/1/token
```

## 常见问题

### Q: 如何查看刷新日志?
A: 在Web界面中点击项目的"查看"按钮，可以看到详细的刷新日志。

### Q: Token刷新失败怎么办?
A: 检查刷新日志中的错误信息，常见问题包括:
- URL配置错误
- 请求体模板格式错误
- JSONPath表达式不匹配响应结构
- Client ID/Secret或Refresh Token无效

### Q: 如何修改刷新频率?
A: 调度器每分钟检查一次，但只有在token即将过期时才会刷新。你可以通过修改"提前刷新时间"来控制刷新时机。

### Q: 支持哪些JSONPath表达式?
A: 使用gjson库，支持标准JSONPath语法。例如:
- `accessToken` - 顶层字段
- `data.token` - 嵌套字段
- `tokens.0.value` - 数组元素

## 许可证

MIT License

## 贡献

欢迎提交Issue和Pull Request!
