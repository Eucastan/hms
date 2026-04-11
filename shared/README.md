# HMS Shared Files

This folder contains shared files that are used through out the app. It contains reusable files like middlewares for **HTTP** and interceptors for **GRPC**, **proto definition** files, **utility** files, and gitignore.

## Files And Folders

- **auth** All **HTTP** Middlewares (authmiddleware, ratelimiter, RBAC)
- **grpcserver** All **GRPC** Middleware (interceptors, retry policy JSON)
- **healthcheck** (health, readiness, liveness for **kubernetes**)
- **logger** All logs, **zap**
- **proto** Proto Definitions Folders (billing, lab, patient, pharmacy) .proto
- **utils** Contains JWT and error files for now

## Dependencies

- Go, Gin, GRPC, JWT etc.
