# HMS Shared Files

This folder contains shared files that are used through out the app. It contains reusable files like middlewares for _HTTP_ and _GRPC_, _proto definition_ files, _utility_ files, and gitignore. This folder house important files needed and reused throughout the application.

# Files And Folders

- auth > All _HTTP_ Middlewares (authmiddleware, ratelimiter, RBAC)
- grpcserver > All _GRPC_ Middleware (interceptors, retry policy JSON)
- proto > Proto Definitions Folders (billing, lab, patient, pharmacy) .proto
- utils > Contains JWT for now

# Dependencies

- Go, Gin, GRPC, JWT etc.
