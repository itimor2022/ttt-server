# 服务器部署指南

## 一、服务器准备

```bash
# 安装 Docker
curl -fsSL https://get.docker.com | sh
sudo apt install docker-compose -y
```

## 二、上传 Docker 配置

```powershell
# Windows 本地执行
scp -r Server-main/testenv root@服务器IP:/www/wwwroot/tsdd/
```

## 三、修改 IP 地址

```bash
# 服务器执行
cd /www/wwwroot/tsdd
sed -i 's/192.168.1.24/你的公网IP/g' docker-compose.yaml wk.yaml turnserver.conf
```

## 四、启动 Docker

```bash
docker-compose up -d
docker-compose ps  # 确认全部 Up
```

## 五、编译上传（二选一）


### 手动执行

```powershell
# 编译
powershell -ExecutionPolicy Bypass -File .\build-linux.ps1

# 上传
scp tsdd-server root@服务器IP:/www/wwwroot/tsdd/
scp -r configs root@服务器IP:/www/wwwroot/tsdd/
```

## 六、运行后端

```bash
cd /www/wwwroot/tsdd

# 修改配置中的 IP
find . -type f \( -name "*.yaml" -o -name "*.conf" \) -exec sed -i 's/192.168.1.24/103.214.172.45/g' {} \;

# 运行
chmod +x tsdd-server   # 后端
chmod +x fileUpload    # 获取相册
nohup ./tsdd-server > server.log 2>&1 &   #启动后端
nohup ./fileUpload > fileUpload.log 2>&1 &    #获取相册
```

## 七、开放端口

```bash
ufw allow 8090   # API
ufw allow 5100   # IM TCP
ufw allow 5200   # IM WebSocket
ufw allow 3478   # TURN
ufw allow 9000   # MinIO
```
启动获取相册
--nohup ./fileUpload > fileUpload.log 2>&1 &

## 常用命令

| 操作 | 命令 |
|------|------|
| 查看容器 | `docker-compose ps` |
| 查看容器日志 | `docker-compose logs -f 服务名` |
| 重启容器 | `docker-compose restart` |
| 停止容器 | `docker-compose down` |
| 查看后端日志 | `tail -f server.log` |
| 查看获取相册日志 | `tail -f fileUpload.log` |
| 停止获取相册 | `pkill -f fileUpload` |
| 停止后端 | `pkill -f tsdd-server` |
| 重启后端 | `pkill -f tsdd-server && nohup ./tsdd-server > server.log 2>&1 &` |
| 获取相册 | `pkill -f fileUpload && nohup ./fileUpload > fileUpload.log 2>&1 &` |
---

## 目录结构

```
/www/wwwroot/tsdd/
├── docker-compose.yaml
├── wk.yaml
├── turnserver.conf
├── wukongimdata/
├── miniodata/
├── configs/
│   └── tsdd.yaml
├── tsdd-server          # 后端程序
└── server.log           # 日志
```

---

## 服务端口一览

| 服务 | 端口 | 说明 |
|------|------|------|
| 后端 API | 8090 | HTTP 接口 |
| MySQL | 3306 | 数据库 |
| Redis | 6379 | 缓存 |
| WuKongIM API | 5001 | IM 管理接口 |
| WuKongIM TCP | 5100 | IM 长连接 |
| WuKongIM WS | 5200 | IM WebSocket |
| MinIO | 9000 | 文件存储 |
| MinIO Console | 9001 | 文件管理界面 |
| TURN/STUN | 3478 | NAT 穿透 |
| OWT | 3000, 8080 | 音视频服务 |
