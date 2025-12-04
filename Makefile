.PHONY: build up down logs restart clean validate help

# Directories
KRAKEND_CE_DIR := ../krakend-ce

# Variables
COMPOSE_FILE = docker-compose.yml
SERVICE_NAME = krakend

# Estas variables deben coincidir con las de ../krakend-ce/Makefile
GOLANG_VERSION := 1.25.3
ALPINE_VERSION := 3.21
VERSION := 2.12.0

help: ## Muestra esta ayuda
	@echo "Comandos disponibles:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

buildg: ## Construye la imagen de KrakenD CE
	echo go.mod > ../krakend-ce/go.mode
	cd ../krakend-ce/ && make docker
# 	make docker-builder
# 	cd ../krakend-custome-ce/ 
# 	docker-compose -f $(COMPOSE_FILE) build

# Build the container using the Dockerfile (alpine)
docker:
	cp -a ../krakend-ce/go.bck ../krakend-ce/go.mod
	cat goBuilderDeps >> ../krakend-ce/go.mod
# 	docker build --no-cache --pull --build-arg GOLANG_VERSION=$(GOLANG_VERSION) --build-arg ALPINE_VERSION=$(ALPINE_VERSION) -t krakend-ce:local -f $(KRAKEND_CE_DIR)/Dockerfile $(KRAKEND_CE_DIR)
	docker build --build-arg GOLANG_VERSION=$(GOLANG_VERSION) --build-arg ALPINE_VERSION=$(ALPINE_VERSION) -t krakend-ce:local -f $(KRAKEND_CE_DIR)/Dockerfile $(KRAKEND_CE_DIR)

docker-builder:
	cp -a ../krakend-ce/go.bck ../krakend-ce/go.mod
	cat goBuilderDeps >> ../krakend-ce/go.mod
# 	docker build --no-cache --pull --target builder --build-arg GOLANG_VERSION=$(GOLANG_VERSION) --build-arg ALPINE_VERSION=$(ALPINE_VERSION) -t krakend-ce-builder:local -f $(KRAKEND_CE_DIR)/Dockerfile $(KRAKEND_CE_DIR)
	docker build --target builder --build-arg GOLANG_VERSION=$(GOLANG_VERSION) --build-arg ALPINE_VERSION=$(ALPINE_VERSION) -t krakend-ce-builder:local -f $(KRAKEND_CE_DIR)/Dockerfile $(KRAKEND_CE_DIR)

docker-builder-linux:
	cp -a ../krakend-ce/go.bck ../krakend-ce/go.mod
	cat goBuilderDeps >> ../krakend-ce/go.mod
# 	docker build --no-cache --pull --target builder --build-arg GOLANG_VERSION=${GOLANG_VERSION} -t krakend-ce-builder:local-linux-generic -f $(KRAKEND_CE_DIR)/Dockerfile $(KRAKEND_CE_DIR)


up: ## Inicia el servicio de KrakenD
	docker-compose -f $(COMPOSE_FILE) up -d
	@echo "KrakenD está ejecutándose en http://localhost:8080"

down: ## Detiene y elimina los contenedores
	docker-compose -f $(COMPOSE_FILE) down

logs: ## Muestra los logs del servicio
	docker-compose -f $(COMPOSE_FILE) logs -f $(SERVICE_NAME)

restart: down up ## Reinicia el servicio

validate: ## Valida la configuración de KrakenD
	docker run --rm -v $(PWD)/config:/etc/krakend krakend-ce:latest check -c /etc/krakend/krakend.json

shell: ## Abre una shell en el contenedor
	docker-compose -f $(COMPOSE_FILE) exec $(SERVICE_NAME) sh

clean: ## Limpia contenedores, imágenes y volúmenes
	docker-compose -f $(COMPOSE_FILE) down -v
	docker rmi krakend-ce:latest 2>/dev/null || true

rebuild: clean build up ## Limpia, reconstruye e inicia

status: ## Muestra el estado del servicio
	docker-compose -f $(COMPOSE_FILE) ps

test: ## Prueba el endpoint de health
	@echo "Probando health check..."
	@curl -s http://localhost:8080/__health || echo "El servicio no está disponible"
	@echo "\nProbando endpoint de ejemplo..."
	@curl -s http://localhost:8080/example | head -n 20 || echo "El endpoint no está disponible"
