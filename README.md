# Aku Web

Aku Web 是一个基于 Go 语言开发的网页音乐播放器和设备控制系统。它支持本地音乐播放、网易云音乐歌单播放，并提供了设备音量控制和 WiFi 配置等功能。

## 功能特性

### 音乐播放
- 本地音乐播放
  - 支持 MP3、WAV 格式
  - 文件管理和播放控制
  - 实时音量调节
- 网易云音乐歌单
  - 歌单导入和播放
  - 支持顺序和随机播放
  - VIP 歌曲标识
  - 歌手和歌曲信息显示

### 设备控制
- WiFi 配置
  - AP 热点自动创建
  - WiFi 连接配置
  - 网络状态监控
- 音量控制
  - 实时音量调节
  - 音量状态保存

### 界面特性
- 响应式设计
- 动态导航页面
- 直观的播放控制
- 实时状态反馈

## 技术栈

### 后端
- Go 1.21+
- 标准库
  - net/http：Web 服务器
  - encoding/json：JSON 处理
  - os/exec：系统命令执行
  - sync：并发控制

### 前端
- HTML5
- CSS3
- JavaScript
  - Fetch API
  - ES6+ 特性

## 安装说明

### 系统要求
- Linux 系统（推荐 Debian/Ubuntu）
- Go 1.21 或更高版本
- mpg123（音频播放）
- hostapd（AP 热点）
- curl（网络请求）

### 安装步骤

1. 克隆项目
```bash
git clone [项目地址]
cd aku-web
```

2. 安装依赖
```bash
# Debian/Ubuntu
sudo apt-get update
sudo apt-get install mpg123 hostapd curl
```

3. 编译项目
```bash
go build
```

## 使用说明

### 启动服务
```bash
./aku-web -port 80 -dir static
```

### 访问界面
打开浏览器访问：`http://设备IP`

### 功能入口
- `/` - 主页导航
- `/music_url.html` - 网易云音乐播放器
- `/music_user.html` - 本地音乐播放器
- `/ap_config.html` - WiFi 配置页面

## 目录结构
```
aku-web/
├── main.go          # 主程序
├── netease/         # 网易云音乐相关
│   └── api.go       # API 实现
├── static/          # 静态文件
│   ├── css/         # 样式文件
│   ├── js/          # JavaScript 文件
│   ├── music/       # 本地音乐文件
│   ├── index.html   # 导航页面
│   ├── music_url.html    # 网易云播放器
│   ├── music_user.html   # 本地音乐播放器
│   └── ap_config.html    # WiFi 配置页面
└── README.md        # 项目文档
```

## API 接口

### 音乐相关
- `GET /api/music/list` - 获取本地音乐列表
- `POST /api/music/play` - 播放指定音乐
- `POST /api/music/stop` - 停止播放
- `GET /api/playlist/detail` - 获取歌单详情
- `POST /api/playlist/play` - 播放歌单

### 设备控制
- `GET /api/volume/get` - 获取当前音量
- `POST /api/volume/set` - 设置音量
- `POST /api/ap/config` - 配置 WiFi 连接

## 注意事项

1. WiFi 配置功能需要 root 权限
2. 部分网易云音乐歌曲可能因为版权限制无法播放
3. 确保设备有足够的存储空间用于缓存音乐文件

## 许可证

[添加许可证信息]

## 贡献指南

欢迎提交 Issue 和 Pull Request。

## 更新日志

### v1.0.0
- 初始版本发布
- 基础音乐播放功能
- WiFi 配置功能
- 设备控制功能
