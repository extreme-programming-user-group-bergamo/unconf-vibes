# 1. Introduction

This document outlines the complete architecture for **UNCONF CLI**, a command-line application with text-based user interface (TUI) designed to streamline conference registration and hotel room booking for developer unconferences. It serves as the single source of truth for development, ensuring consistency across the CLI client, backend API, and infrastructure.

**Architecture Approach:** This is a **unified fullstack architecture** combining what would traditionally be separate CLI and backend documents. Since UNCONF is a Go-based monorepo where the CLI and API share code (types, constants, utilities), this integrated approach ensures coherent design decisions.

## 1.1 Starter Template or Existing Project

**N/A — Greenfield project**

The PRD specifies a greenfield Go project with standard layout. No starter template is being used; the architecture will establish patterns from scratch using:
- Go standard project layout (`/cmd`, `/internal`, `/pkg`)
- Cobra + Viper for CLI
- Bubble Tea + Lip Gloss for TUI
- Gin for REST API

## 1.2 Change Log

| Date | Version | Description | Author |
|------|---------|-------------|--------|
| Jan 31, 2026 | 1.0 | Initial architecture document | Winston (Architect) |

---
