# integreated_exporter 服务运行手册

## 1. 配置文件准备
```bash
sudo nano /etc/systemd/system/integrated-exporter.service
```
## 2. 写入内容
```ini
[Unit]
Description=Integrated Exporter Service
After=network.target

[Service]
Type=simple
User=ubuntu
WorkingDirectory=/home/ubuntu/src/tunnel_server
ExecStart=/home/ubuntu/src/tunnel_server/integrated_exporter server --port=6070 --config monitor.yaml
Restart=always
RestartSec=5

# 日志直接进 journalctl
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
```

## 3. 启动服务
```bash
sudo systemctl daemon-reload
sudo systemctl start integrated-exporter
```

## 4. 查看服务状态
```bash
sudo systemctl status integrated-exporter
journalctl -u integrated-exporter -f
```