# mDNS资产测绘CLI工具

一个用于扫描和识别局域网中mDNS服务资产的命令行工具。

## 功能特性

- 支持CIDR和IP范围格式的网段扫描
- 支持单端口、端口范围和端口列表
- 深度识别mDNS服务信息
- 支持QNAP、Synology、Apple等设备的详细banner解析
- YAML格式输出，易于解析和查看

## 安装

```bash
go build -o mdnsscan.exe .
```

## 使用方法

### 基本用法

```bash
# 扫描指定网段和端口范围
mdnsscan --cidr 192.168.1.0/24 --ports 1-1000

# 使用IP范围格式
mdnsscan --cidr 192.168.1.1-192.168.1.255 --ports 80,443,5000

# 指定并发数和超时
mdnsscan --cidr 10.0.0.0/24 --ports 1-65535 --workers 100 --timeout 10

# 详细输出模式
mdnsscan --cidr 192.168.1.0/24 --ports 5000 -v

# 输出到文件
mdnsscan --cidr 192.168.1.0/24 --ports 5000 -o result.yaml

# 演示输出格式
mdnsscan --demo
```

### 参数说明

| 参数 | 简写 | 说明 | 默认值 |
|------|------|------|--------|
| --cidr | -c | IP网段（CIDR或范围格式） | 必填 |
| --ports | -p | 端口范围 | 必填 |
| --timeout | -t | 扫描超时时间(秒) | 5 |
| --workers | -w | 并发工作线程数 | 50 |
| --verbose | -v | 详细输出模式 | false |
| --output | -o | 输出文件路径 | stdout |
| --demo | -d | 演示输出格式 | false |

### 端口格式支持

- 单端口: `80`
- 端口范围: `1-1000`
- 端口列表: `80,443,5000`
- 混合格式: `80,443,5000-6000,8080`

## 输出示例

```yaml
services:
  9/tcp workstation:
    Name=slw-nas [24:5e:be:69:a3:13]
    IPv4=192.168.1.100
    IPv6=fe80::265e:beff:fe69:a313
    Hostname=slw-nas.local
    TTL=10
  5000/tcp http:
    Name=slw-nas
    IPv4=192.168.1.100
    IPv6=fe80::265e:beff:fe69:a313
    Hostname=slw-nas.local
    TTL=10
    path=/
  5000/tcp qdiscover:
    Name=slw-nas
    IPv4=192.168.1.100
    IPv6=fe80::265e:beff:fe69:a313
    Hostname=slw-nas.local
    TTL=10
    accessType=https
    accessPort=86
    model=TS-X64
    displayModel=TS-464C
    fwVer=5.2.9
    fwBuildNum=20260214

answers:
  PTR:
    _workstation._tcp.local
    _http._tcp.local
    _qdiscover._tcp.local
```

## 支持的服务类型

- `_workstation._tcp` - 工作站服务
- `_http._tcp` - HTTP服务
- `_https._tcp` - HTTPS服务
- `_smb._tcp` - SMB文件共享
- `_afpovertcp._tcp` - AFP文件共享(Apple)
- `_ssh._tcp` - SSH服务
- `_ftp._tcp` - FTP服务
- `_printer._tcp` - 打印机服务
- `_airplay._tcp` - AirPlay服务
- `_qdiscover._tcp` - QNAP发现服务
- `_device-info._tcp` - 设备信息服务
- 以及更多...

## 技术架构

```
├── cmd/                    # 命令行入口
│   └── root.go            # Cobra CLI框架
├── internal/
│   ├── config/            # 配置管理
│   ├── parser/            # IP和端口解析
│   ├── scanner/
│   │   ├── mdns/         # mDNS扫描核心
│   │   └── port/         # 端口扫描
│   └── output/           # 输出格式化
├── pkg/
│   ├── models/           # 数据模型
│   └── utils/            # 工具函数
├── main.go               # 程序入口
└── go.mod                # 依赖管理
```

## 依赖库

- `github.com/spf13/cobra` - CLI框架
- `github.com/miekg/dns` - DNS/mDNS协议解析
- `gopkg.in/yaml.v3` - YAML格式输出

## 许可证

MIT License
