---
applyTo: '**'
---

# Project Overview Guide

## Project Summary

This project consists of three core subprojects that together form a complete IP tunneling service system:

1. **tunnel_client** — IP tunnel client  
2. **tunnel_server** — IP tunnel server  
3. **tunnel_monitor** — Monitoring and management system  

---

## tunnel_client – IP Tunnel Client

A high-performance IP tunnel client written in Go and designed specifically for Linux.  
It supports the WireGuard protocol, provides secure encrypted channels, and includes intelligent traffic control and DNS resolution capabilities.

### Key Features

- **Secure Tunneling**: End-to-end encrypted communication powered by WireGuard  
- **Traffic Control**: Precise bandwidth management and traffic shaping  
- **DNS Resolution**: Intelligent DNS resolver with multi-IP concurrent testing  
- **Monitoring Integration**: Built-in support for Prometheus and Grafana  
- **High Availability**: Multi-server configuration and automatic failover  

### How It Works

The client marks packets via iptables and applies bandwidth limits using Linux Traffic Control (TC).  
It leverages the HTB (Hierarchical Token Bucket) model to deliver highly accurate traffic shaping.

---

## tunnel_server – IP Tunnel Server

A backend service built using Go and the Jzero framework.  
It handles user accounts, orders, machine configurations, IP allocation, and other core VPN/tunnel-related logic.

### Core Capabilities

- **User Management**: Account handling and subnet assignment  
- **Order Management**: Order creation and lifecycle tracking  
- **Machine Management**: POP server configuration and WireGuard key management  
- **IP Allocation**: Intelligent IP distribution mechanism  
- **DNS Service**: Remote-table-based domain resolution  

### Architecture

The server follows a layered architecture and domain-driven design principles.  
It supports stateless horizontal scaling, uses MySQL for persistent storage, and integrates with Prometheus for metrics collection.

---

## tunnel_monitor – Monitoring Management System

A standalone monitoring controller responsible for configuring and managing Prometheus and Grafana.

### Major Features

- **Component Installation**: Automated deployment of Prometheus and Grafana  
- **Service Control**: Start, stop, and status monitoring  
- **Dashboard Management**: Create and maintain monitoring dashboards  
- **Configuration Management**: Centralized handling of monitoring configs and service URLs  

### Design Highlights

Dashboard templates are modularized—large JSON templates are broken into smaller files for easier maintenance and version control.  
It supports unified dashboards for both the client and the server.

---

## System Collaboration Model

The three subprojects work together to form a cohesive IP tunneling ecosystem:

1. **tunnel_server** acts as the central controller, managing authentication, configuration distribution, and status reporting.  
2. **tunnel_client** runs on the user’s machine, pulls configurations from the server, and establishes secure tunnels.  
3. **tunnel_monitor** operates independently, providing system-wide observability and visualization.

This architecture enforces separation of concerns, ensuring scalability, robustness, and maintainability across the entire platform.