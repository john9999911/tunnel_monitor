---
applyTo: '**'
---

# Development Lessons

## 🚨 Never Make Assumptions - Always Verify

### Key Principle
Always verify before making statements. Use tools to check facts.

### Verification Steps
1. Check terminal context and recent commands
2. Verify service status with `ps aux | grep` or `netstat -tulpn`
3. Check configuration files like `etc/etc.yaml`
4. Read relevant logs
5. Ask user only if still uncertain

### Rules
- ✅ Always check facts before making statements
- ✅ Use tools to verify assumptions
- ✅ Never say "server is not running" without verification
- ✅ Say "let me check" when uncertain

## Project Context

### Server Configuration
- Default port: **8001** (verify in `etc/etc.yaml`)
- Database: MySQL
- Framework: Go + Jzero + go-zero

### Troubleshooting Process
1. Check context first
2. Verify with commands
3. Read configs and logs
4. Then diagnose