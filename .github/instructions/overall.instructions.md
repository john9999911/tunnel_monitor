---
applyTo: '**'
---

# Project Overview Guide

**Updated**: 2026-03-05 (UTC)

## Project Summary

This workspace has three collaborating subprojects:

1. **tunnel_server**: central control plane (auth, config distribution, IP/bandwidth/order management)
2. **tunnel_client**: WireGuard tunnel executor (routing, traffic shaping, DNS handling, reporting)
3. **tunnel_monitor**: observability stack manager (Prometheus/Grafana lifecycle and dashboards)

## Collaboration Model

- Server distributes configs and policies
- Client applies configs and maintains encrypted tunnels
- Monitor collects and visualizes metrics from both sides

## Practical Orientation

- Keep subproject boundaries clear; avoid mixing responsibilities across folders
- Prefer subproject-specific instruction files for implementation details
- Use this file only as context, not as implementation-level rules