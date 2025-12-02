.PHONY: build up down logs restart clean validate help

# Variables
COMPOSE_FILE = docker-compose.yml
SERVICE_NAME = krakend

help: ## Muestra esta ayuda
	@echo "Comandos disponibles:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

build: ## Construye la imagen de KrakenD CE
	docker-compose -f $(COMPOSE_FILE) build

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
