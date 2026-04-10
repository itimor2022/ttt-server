# 客户端配置指南

## Nginx 反向代理配置

```nginx
server
{
    listen 80;
    server_name api.yujj888.top;
    
    #CERT-APPLY-CHECK--START
    include /www/server/panel/vhost/nginx/well-known/api.yujj888.top.conf;
    #CERT-APPLY-CHECK--END

    # ============ 后端 API ============
    location /v1 {
        proxy_pass http://127.0.0.1:8090;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        client_max_body_size 100m;
    }

    # ============ IM WebSocket ============
    location /ws {
        proxy_pass http://127.0.0.1:5200;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_read_timeout 3600s;
    }

    # ============ 音视频 OWT 信令 ============
    location /socket.io {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_read_timeout 3600s;
    }

    # OWT 管理 API
    location /owt {
        proxy_pass http://127.0.0.1:3004;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }

    # 默认
    location / {
        return 200 '{"status":"ok"}';
        add_header Content-Type application/json;
    }

    access_log  /www/wwwlogs/api.yujj888.top.log;
    error_log  /www/wwwlogs/api.yujj888.top.error.log;
}
```

---

## 防火墙端口

```bash
# 必须开放
ufw allow 80          # HTTP
ufw allow 443         # HTTPS
ufw allow 5100        # IM TCP 直连
ufw allow 3478/tcp    # TURN
ufw allow 3478/udp    # STUN
ufw allow 5349/tcp    # TURN TLS
ufw allow 5349/udp
ufw allow 49152:49200/udp  # TURN RTP
ufw allow 59000:59050/udp  # OWT RTP
```

---

## Android 配置

### 1. API 地址

文件: `app/src/main/java/com/test/demo/TSApplication.kt`

```kotlin
// 修改 apiURL
val apiURL = "http://api.yujj888.top"
```

### 2. ICE 服务器（音视频）

文件: `app/src/main/java/com/test/demo/TSApplication.kt` 中的 `getList()` 方法

```kotlin
private fun getList(): ArrayList<PeerConnection.IceServer> {
    // STUN 服务器
    val stunServer = PeerConnection.IceServer.builder(
        "stun:103.214.172.45:3478"  // ← 修改为服务器 IP
    ).createIceServer()

    // TURN 服务器
    val turnServer = PeerConnection.IceServer.builder(
        "turn:103.214.172.45:3478?transport=udp"  // ← 修改为服务器 IP
    ).setUsername("tsdd").setPassword(
        "tsdd123456"
    ).createIceServer()

    val iceServers: ArrayList<PeerConnection.IceServer> = ArrayList()
    iceServers.add(stunServer)
    iceServers.add(turnServer)
    return iceServers
}
```

> **注意**: TURN 账号密码在 `testenv/turnserver.conf` 中配置

---

## iOS 配置

### 1. API 地址

文件: `WuKongIMSDK/Config.swift` 或 `AppDelegate.swift`

```swift
let apiURL = "http://api.yujj888.top"
```

### 2. ICE 服务器

```swift
let stunServer = RTCIceServer(urlStrings: ["stun:103.214.172.45:3478"])
let turnServer = RTCIceServer(
    urlStrings: ["turn:103.214.172.45:3478?transport=udp"],
    username: "tsdd",
    credential: "tsdd123456"
)
```

---

## Web 配置

### 1. 环境变量

文件: `apps/web/.env.production`

```env
VITE_API_URL=http://api.yujj888.top
```

### 2. ICE 服务器

文件: `packages/tsdaodaobase/src/Service/WKSDK.ts` 或相关 WebRTC 配置

```javascript
const iceServers = [
  { urls: 'stun:103.214.172.45:3478' },
  { 
    urls: 'turn:103.214.172.45:3478?transport=udp',
    username: 'tsdd',
    credential: 'tsdd123456'
  }
];
```

---

## 管理后台配置

文件: `.env.production`

```env
VITE_API_URL=http://api.yujj888.top
```

---

## 服务端口总览

| 服务 | 端口 | 协议 | 代理 |
|------|------|------|------|
| 后端 API | 8090 | HTTP | ✅ Nginx `/v1` |
| IM WebSocket | 5200 | WS | ✅ Nginx `/ws` |
| IM TCP | 5100 | TCP | ❌ 直连 IP |
| OWT 信令 | 8080 | WS | ✅ Nginx `/socket.io` |
| TURN/STUN | 3478 | UDP/TCP | ❌ 直连 IP |
| RTP 媒体 | 49152-49200 | UDP | ❌ 直连 IP |

---

## 快速检查

```bash
# 测试 API
curl http://api.yujj888.top/v1/common/appconfig

# 测试 IM
telnet 103.214.172.45 5100

# 测试 TURN
turnutils_uclient -T -u tsdd -w tsdd123456 103.214.172.45
```

