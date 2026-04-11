.PHONY: help test-up test-down test test-integration \
        build-all docker-build-all docker-push-all \
        up down logs k8s-apply k8s-delete clean

# COLORS
GREEN  := $(shell tput -Txterm setaf 2)
YELLOW := $(shell tput -Txterm setaf 3)
RESET  := $(shell tput -Txterm sgr0)

# HELP
help:
	@echo "${GREEN}HMS Microservices - Available Commands${RESET}"
	@echo ""
	@echo "${YELLOW}Testing:${RESET}"
	@echo "  make test-up             → Start test PostgreSQL"
	@echo "  make test                → Run unit + fast tests"
	@echo "  make test-integration    → Run full integration tests"
	@echo ""
	@echo "${YELLOW}Build:${RESET}"
	@echo "  make build-all           → Build all services binaries"
	@echo "  make docker-build-all    → Build all Docker images"
	@echo "  make docker-push-all     → Push all images to Docker Hub"
	@echo ""
	@echo "${YELLOW}Development:${RESET}"
	@echo "  make up                  → Start all services"
	@echo "  make down                → Stop all services"
	@echo "  make logs-clinical       → Tail clinical logs"
	@echo ""
	@echo "${YELLOW}Kubernetes:${RESET}"
	@echo "  make k8s-apply           → Deploy to Kubernetes"
	@echo "  make k8s-delete          → Delete all resources"

# TESTING
test-up:
	docker compose -f docker-compose.test.yaml up -d --build postgres

test-down:
	docker compose -f docker-compose.test.yaml down -v

test:
	go test ./... -v -short

test-integration:
	go test ./... -v

# BUILD
# ====================== BUILD ======================
build-all:
	@echo "${GREEN}Building all services...${RESET}"
	@for svc in api-gateway auth clinical lab patient payment pharmacy; do \
		echo "→ Building $$svc..."; \
		if [ "$$svc" = "api-gateway" ]; then \
			cd $$svc && go build -o bin/$$svc ./cmd/gateway/main.go && cd ..; \
		else \
			cd $$svc && go build -o bin/$$svc ./cmd/http/main.go && cd ..; \
		fi; \
	done
	@echo "${GREEN}All services built successfully!${RESET}"

# DOCKER
docker-build-all:
	@echo "${GREEN}Building Docker images from root context...${RESET}"
	@for svc in api-gateway auth clinical lab patient payment pharmacy; do \
		echo "→ Building $$svc..."; \
		docker build -t eucastan001/$$svc:latest -f $$svc/Dockerfile .; \
	done

docker-push-all:
	@echo "${GREEN}Pushing all images to Docker Hub...${RESET}"
	@for svc in api-gateway auth clinical lab patient payment pharmacy; do \
		echo "→ Pushing $$svc..."; \
		docker push eucastan001/$$svc:latest; \
	done

# LOCAL DEVELOPMENT
up:
	docker compose up -d --build

down:
	docker compose down -v

logs-clinical:
	docker compose logs -f clinical

logs-pharmacy:
	docker compose logs -f pharmacy

logs-payment:
	docker compose logs -f payment

# KUBERNETES
k8s-apply:
	@echo "${GREEN}Applying Kubernetes manifests...${RESET}"
	kubectl apply -f k8s/namespace.yaml
	kubectl apply -f k8s/api-gateway/
	kubectl apply -f k8s/auth/
	kubectl apply -f k8s/clinical/
	kubectl apply -f k8s/lab/
	kubectl apply -f k8s/patient/
	kubectl apply -f k8s/payment/
	kubectl apply -f k8s/pharmacy/

k8s-delete:
	@echo "${YELLOW}Deleting HMS resources...${RESET}"
	kubectl delete -f k8s/ --ignore-not-found=true

# UTILS
clean:
	rm -rf */bin

restart-clinical:
	docker compose restart clinical

.PHONY: help test-up test-down test test-integration build-all docker-build-all docker-push-all up down logs k8s-apply k8s-delete clean