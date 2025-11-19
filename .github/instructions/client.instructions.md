client.instructions.md
---
applyTo: 'tunnel_client/**'
---

# Tunnel Client 项目指导

## 项目概述

Tunnel Client 是一个基于 Go 语言开发的高性能 IP 隧道客户端，专为 Linux 平台设计。它支持 WireGuard 协议，提供安全的加密通信通道，并具备智能流量控制和 DNS 解析服务功能。

## 核心功能

1. **安全隧道**：基于 WireGuard 协议的端到端加密通信
2. **流量控制**：精确的带宽限制和流量整形功能
3. **DNS 解析**：智能 DNS 解析服务，支持多 IP 并发测试
4. **监控集成**：与 Prometheus 和 Grafana 集成，提供实时监控
5. **高可用性**：支持多服务器配置和自动故障转移

## 项目架构

```
tunnel_client/
├── cmd/                   # 命令行接口
│   ├── root.go            # 根命令定义
│   ├── lifecycle.go       # 生命周期管理
│   └── cleanup.go         # 清理命令
├── internal/              # 内部模块
│   ├── config/            # 配置管理
│   ├── core/              # 核心接口和工厂
│   ├── dns/               # DNS 解析服务
│   ├── hosts/             # Hosts 文件管理
│   ├── httpclient/        # 高可用HTTP客户端
│   ├── logger/            # 日志系统
│   ├── mode/              # 运行模式管理
│   ├── monitor/           # 监控系统
│   ├── platform/          # 平台相关接口
│   ├── shared/            # 共享常量和类型
│   ├── tunnel/            # 隧道核心
│   │   ├── api/           # API 接口
│   │   ├── wireguard/     # WireGuard 管理
│   │   ├── traffic/       # 流量控制
│   │   ├── nat/           # NAT 管理
│   │   ├── route/         # 路由管理
│   │   └── cleanup.go     # 隧道清理
│   └── utils/             # 通用工具函数
├── config/                # 配置文件
└── docs/                  # 文档
```

## 运行模式

客户端支持两种运行模式：

1. **User Mode（用户模式）**：普通用户使用，连接到 POP 节点
2. **POP Mode（接入点模式）**：作为接入点，为用户提供网络服务

## 关键模块说明

### Tunnel 核心模块
负责管理整个隧道生命周期，协调 WireGuard、Traffic、NAT、Route 等子模块。

### Monitor 监控模块
收集和上报客户端的各种监控指标，与 Prometheus 集成。

### DNS 模块
执行域名解析和智能 IP 选择功能，通过并发 ping 测试选择最优 IP。

### HTTP Client 模块
提供具有重试和故障转移能力的 HTTP 客户端，用于与服务端通信。

## 工作原理

1. 客户端启动时根据配置连接到服务端
2. 获取隧道配置信息，包括 WireGuard 密钥、端点等
3. 建立 WireGuard 隧道连接
4. 应用流量控制策略（如配置）
5. 启动 DNS 解析服务
6. 持续监控并上报状态信息