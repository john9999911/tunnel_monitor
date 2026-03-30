---
applyTo: '**'
---

# Development Lessons

**Updated**: 2026-03-05 (UTC)

## 🚨 Never Make Assumptions - Always Verify

### Key Principle
Always verify before making statements. Use tools to check facts.

### Verification Steps
1. Check terminal context and recent commands
2. Verify process/network state with commands (`ps`, `ss`/`netstat`)
3. Check relevant configuration files before concluding
4. Read related logs and recent changes
5. Ask user only if still uncertain

### Rules
- ✅ Always check facts before making statements
- ✅ Use tools to verify assumptions
- ✅ Say "let me check" when uncertain

## Troubleshooting Sequence

1. Establish current state (cwd, branch, running processes)
2. Reproduce or locate the failing path
3. Validate config and dependency assumptions
4. Propose fix only after evidence is complete