#!/bin/sh

# AP 配置
AP_SSID="HlameMastar"
AP_PASSWORD="12345678"
AP_INTERFACE="wlan0"
AP_IP="192.168.4.1"

# 日志函数
log() {
    echo "$1"
}

# 调试信息函数
debug_info() {
    log "=== Debug Info ==="
    log "Wireless Interface Status:"
    ip link show $AP_INTERFACE
    log "IP Configuration:"
    ip addr show $AP_INTERFACE
    log "Process Status:"
    ps | grep hostapd
    log "hostapd Config:"
    cat /etc/hostapd.conf
    log "=== Debug Info End ==="
}

# 检查是否以 root 权限运行
check_root() {
    if [ "$(id -u)" -ne 0 ]; then
        log "Please run with root privileges"
        exit 1
    fi
}

# 停止现有的网络服务
stop_network_services() {
    log "Stopping network services..."
    
    # 停止 avahi-daemon（如果存在）
    killall avahi-daemon 2>/dev/null
    
    # 停止 wpa_supplicant
    if [ -f /etc/init.d/S50wpa_supplicant ]; then
        /etc/init.d/S50wpa_supplicant stop
    else
        killall wpa_supplicant 2>/dev/null
    fi
    
    # 停止现有的 hostapd
    killall hostapd 2>/dev/null
    
    # 等待服务完全停止
    sleep 2
}

# 配置网络接口
configure_interface() {
    log "Configuring interface $AP_INTERFACE..."
    
    # 关闭接口
    ip link set $AP_INTERFACE down || true
    sleep 1
    
    # 清除现有 IP 配置
    ip addr flush dev $AP_INTERFACE || true
    
    # 禁用 IPv6
    sysctl -w net.ipv6.conf.$AP_INTERFACE.disable_ipv6=1 >/dev/null 2>&1 || true
    
    # 设置静态 IP
    ip addr add $AP_IP/24 dev $AP_INTERFACE || true
    
    # 启用接口
    ip link set $AP_INTERFACE up || true
    sleep 1
    
    # 禁用 DHCP 客户端
    if [ -f /sbin/dhclient ]; then
        killall dhclient 2>/dev/null
    fi
    if [ -f /sbin/udhcpc ]; then
        killall udhcpc 2>/dev/null
    fi
    
    # 显示接口状态
    log "Interface configuration completed:"
    ip addr show $AP_INTERFACE
}

# 创建 hostapd 配置
create_hostapd_config() {
    log "Creating hostapd configuration..."
    
    # 使用更详细的配置
    cat > /etc/hostapd.conf << EOF
interface=$AP_INTERFACE
driver=nl80211
ssid=$AP_SSID
hw_mode=g
channel=1
auth_algs=1
wpa=2
wpa_passphrase=$AP_PASSWORD
wpa_key_mgmt=WPA-PSK
wpa_pairwise=CCMP
rsn_pairwise=CCMP
beacon_int=100
dtim_period=2
max_num_sta=5
rts_threshold=2347
fragm_threshold=2346
EOF

    log "hostapd configuration created"
}

# 启动服务
start_services() {
    log "Starting hostapd..."
    hostapd -B /etc/hostapd.conf
    sleep 2
    
    # 检查 hostapd 是否正在运行
    if ! ps | grep -v grep | grep hostapd > /dev/null; then
        log "Warning: hostapd may not have started properly"
        debug_info
    else
        log "hostapd started successfully"
    fi
}

# 主函数
main() {
    log "Starting AP hotspot creation..."
    
    check_root
    stop_network_services
    configure_interface
    create_hostapd_config
    start_services
    
    log "AP hotspot creation completed"
    log "SSID: $AP_SSID"
    log "Password: $AP_PASSWORD"
    log "IP Address: $AP_IP"
    
    # 显示调试信息
    debug_info
}

# 清理函数
cleanup() {
    log "Stopping AP hotspot..."
    killall hostapd 2>/dev/null
    ip addr flush dev $AP_INTERFACE 2>/dev/null
    log "AP hotspot stopped"
}

# 注册清理函数
trap cleanup INT TERM

# 运行主函数
main

# 保持脚本运行并定期检查状态
while true; do
    sleep 30
    debug_info
done 