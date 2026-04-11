# HMS - Hospital Management System

A modern **microservices-based Hospital Management System** built with Golang(Gin), designed to demonstrate clean architecture, inter-service communication, and full DevOps practices.

## Project Overview

This project implements a complete backend for a hospital management system using a **microservices architecture**. It includes patient management, clinical records, laboratory requests and results, pharmacy operations, billing, and authentication.

## Services

- **API Gateway** (Reverse proxy + rate limiting)
- **Auth** (Register, Login, Authentication + Authorization)
- **Clinical** (Diagnosis, prescription, patient records)
- **Lab** (Lab Test Request, Lab Results)
- **Patient** (Patient details, Admission, CRUD + search)
- **Pharmacy** (Drug management, dispensing, stock control)
- **Payment** (Invoicing generating, billing, refunds)

## Features

- Clean Architecture (layers: models → repo → service → handler)
- Authentication & Authorization (JWT based with role management)
- gRPC inter-service communication
- Transactional stock management
- Structured logging with Zap
- Health Checks & Graceful shutdown (HTTP + gRPC)
- Full integration testing
- Docker + docker-compose support
- Kubernetes-ready manifests

## Tech Stack

- Language (Go 1.25)
- Framework (Gin Gonic)
- Database (PostgreSQL + GORM)
- Inter-service (gRPC + Protocol Buffers)
- Authentication (JWT + Custom Middleware)
- Containerization (Docker Multi-stage builds)
- Orchestration (Docker Compose, Kubernetes)
- Testing (Testcontainers, testify)
- CI/CD (GitHub Actions)

## Security & Secrets

**Important Note:**

- All secrets and sensitive values in the `k8s/` folder are **example/fake values** only.
- They are **not** used in any production environment.
- In a real production setup, secrets should be managed using:
  - Kubernetes External Secrets Operator
  - HashiCorp Vault
  - AWS Secrets Manager / Azure Key Vault / GCP Secret Manager
  - Sealed Secrets (for GitOps)

Never commit real passwords, JWT secrets, or database credentials to GitHub.

## How to Run Locally

### Prerequisites

- Docker & Docker Compose
- Go 1.25

## Quick Start Commands

```bash
# Clone the repo
git clone https://github.com/Eucastan/hms.git
cd hms

# Show all commands
make help

# Builds all services
make build-all

# Builds Docker images
make docker-build-all

# Start all services
make up

# Run tests
make test

# Deploy to Kubernetes
make k8s-apply

# Stop services
make down
```

## Current Status

**Completed:**

- All microservices
- Dockerfiles for all services
- gRPC communication between services
- Docker Compose setup
- Basic Kubernetes manifests
- CI/CD pipeline

**Note:** Due to hardware limitations(8GB RAM laptop), running all services simultaneously in a VirtualBox VM(4GB RAM) is unstable. Each service builds and runs successfully individually.

## Improvements

This project is 100% open for improvements.

- Add observability (OpenTelemetry + Prometheus + Grafana)
- Circuit breakers
- Complete end-to-end integration tests
