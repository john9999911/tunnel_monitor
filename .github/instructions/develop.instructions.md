---
applyTo: '**'
---

# Development Guidelines

**Updated**: 2026-03-05 (UTC)

## 1. Development Workflow

- Understand the requirement and affected scope before changing code
- Prefer minimal, surgical changes; avoid unrelated refactors
- Keep implementation and documentation synchronized
- Verify behavior with focused checks before broad checks

## 2. Code Development Standards

### 2.1 General Principles
- Keep naming and behavior consistent with existing domain conventions
- Include concrete identifiers in error messages (codes, IPs, IDs)
- Do not edit generated files; regenerate from source contracts/schemas

### 2.2 Execution and Verification
- Prefer project scripts for build/start/restart flows
- Validate only impacted scope first, then expand if needed

### 2.3 Documentation
- Keep writing concise and clear
- Avoid duplicating content across multiple documents