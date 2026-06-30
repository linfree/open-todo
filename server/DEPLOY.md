# Open Todo Server 部署指南

## SQLite 部署（推荐自建）

最简单的部署方式，零依赖。

### 直接运行
```bash
./open-todo-server --driver sqlite --dsn /data/open-todo.db
```

### Docker
```bash
docker run -d -p 8080:8080 -v /data:/data open-todo-server
```

### 环境变量
| 变量 | 默认值 | 说明 |
|------|--------|------|
| PORT | 8080 | 服务端口 |
| DB_DRIVER | sqlite | 数据库驱动 |
| DB_DSN | ./open-todo-server.db | 数据库路径 |
| JWT_SECRET | 随机生成 | JWT 签名密钥（务必设置固定值） |

## PostgreSQL 部署（高并发）

### Docker Compose
```yaml
version: '3.8'
services:
  db:
    image: postgres:16-alpine
    environment:
      POSTGRES_USER: opentodo
      POSTGRES_PASSWORD: changeme
      POSTGRES_DB: opentodo
    volumes:
      - pgdata:/var/lib/postgresql/data
  server:
    image: open-todo-server
    ports:
      - "8080:8080"
    environment:
      PORT: 8080
      DB_DRIVER: postgres
      DB_DSN: postgres://opentodo:changeme@db:5432/opentodo?sslmode=disable
      JWT_SECRET: your-secret-key
    depends_on:
      - db
volumes:
  pgdata:
```

### Nginx 反向代理

```nginx
server {
    listen 80;
    server_name todo.example.com;

    client_max_body_size 10M;

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

## 桌面端连接

在桌面应用设置中填写服务端地址，登录后即可自动同步。
