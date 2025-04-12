# Aku Web

Aku Web 是一个基于 Go 语言开发的网页音乐播放器和设备控制系统。它支持本地音乐播放、网易云音乐歌单播放，并提供了设备音量控制和 WiFi 配置等功能。

## 功能特性

### 音乐播放
- 本地音乐播放
  - 支持 MP3、WAV 格式
  - 文件管理和播放控制
  - 实时音量调节
  - 启动音效支持
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
- 系统服务管理
  - 服务状态监控
  - 服务启停控制

### 界面特性
- 响应式设计
- 动态导航页面
- 直观的播放控制
- 实时状态反馈
- 多语言支持（中文/英文）

## 技术栈

### 后端
- Go 1.21+
- 标准库
  - net/http：Web 服务器
  - encoding/json：JSON 处理
  - os/exec：系统命令执行
  - sync：并发控制
  - log：日志管理

### 前端
- HTML5
- CSS3
- JavaScript
  - Fetch API
  - ES6+ 特性
  - 响应式设计

## 系统要求
- Linux 系统（推荐 Debian/Ubuntu）
- Go 1.21 或更高版本
- mpg123（音频播放）
- hostapd（AP 热点）
- curl（网络请求）
- systemd（服务管理）

## 安装说明

### 1. 克隆项目
```bash
git clone [项目地址]
cd aku-web
```

### 2. 安装依赖
```bash
# Debian/Ubuntu
sudo apt-get update
sudo apt-get install mpg123 hostapd curl
```

### 3. 编译项目
```bash
# 使用 build.ps1 脚本（Windows）
./build.ps1

```

## 使用说明

### 启动服务
```bash
# 直接运行
./aku-web -port 80 -dir static

# 或作为系统服务
sudo systemctl start aku-web
```

### 访问界面
打开浏览器访问：`http://设备IP`

### 功能入口
- `/` - 主页导航
- `/music_url.html` - 网易云音乐播放器
- `/music_user.html` - 本地音乐播放器
- `/ap_config.html` - WiFi 配置页面
- `/system.html` - 系统服务管理
- `/service.html` - 服务状态监控

## 项目结构
```
aku-web/
├── main.go          # 主程序入口
├── internal/        # 内部模块
│   ├── api/        # API 接口定义
│   ├── service/    # 业务逻辑
│   ├── router/     # 路由处理
│   ├── server/     # 服务器配置
│   ├── netease/    # 网易云音乐相关
│   ├── player/     # 音频播放
│   ├── wifi/       # WiFi 管理
│   └── config/     # 配置管理
├── static/         # 静态资源
│   ├── css/        # 样式文件
│   ├── js/         # JavaScript 文件
│   ├── music/      # 本地音乐文件
│   ├── boot_music/ # 启动音效
│   ├── icon/       # 图标资源
│   └── *.html      # 页面文件
├── release/        # 发布文件
├── build.ps1       # Windows 构建脚本
├── create_ap.sh    # Linux AP 创建脚本
└── README.md       # 项目文档
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
- `GET /api/system/status` - 获取系统状态
- `POST /api/system/control` - 控制系统服务

## 开发指南

### 环境配置
1. 安装 Go 1.21+
2. 配置 GOPATH
3. 安装必要的系统依赖

### 代码规范
- 遵循 Go 官方代码规范
- 使用 gofmt 格式化代码
- 编写单元测试
- 添加必要的注释

### 提交规范
- 使用语义化版本号
- 编写清晰的提交信息
- 确保代码通过测试

## 注意事项

1. WiFi 配置功能需要 root 权限
2. 部分网易云音乐歌曲可能因为版权限制无法播放
3. 确保设备有足够的存储空间用于缓存音乐文件
4. 系统服务管理需要 systemd 支持

## 许可证

[添加许可证信息]

## 贡献指南

欢迎提交 Issue 和 Pull Request。在提交之前，请确保：
1. 代码符合项目规范
2. 添加必要的测试
3. 更新相关文档

## 更新日志

### v1.1.0 (最新)
- 添加系统服务管理功能
- 优化 WiFi 配置界面
- 增加启动音效支持
- 改进错误处理和日志记录

### v1.0.0
- 初始版本发布
- 基础音乐播放功能
- WiFi 配置功能
- 设备控制功能
