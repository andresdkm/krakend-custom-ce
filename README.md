# KrakenD CE - Docker Build from Source

Este proyecto contiene los archivos necesarios para clonar, compilar y ejecutar KrakenD CE desde el código fuente de GitHub.

## Estructura del Proyecto

```
.
├── Dockerfile              # Archivo de construcción multi-etapa
├── docker-compose.yml      # Configuración de servicios Docker
├── config/
│   └── krakend.json       # Archivo de configuración de KrakenD
├── plugins/               # Directorio para plugins personalizados (opcional)
└── README.md              # Este archivo
```

## Requisitos

- Docker
- Docker Compose

## Construcción y Ejecución

### Opción 1: Usando Docker Compose (Recomendado)

```bash
# Construir la imagen
docker-compose build

# Iniciar el servicio
docker-compose up -d

# Ver logs
docker-compose logs -f

# Detener el servicio
docker-compose down
```

### Opción 2: Usando Docker directamente

```bash
# Construir la imagen
docker build -t krakend-ce:latest .

# Ejecutar el contenedor
docker run -d \
  --name krakend \
  -p 8080:8080 \
  -v $(pwd)/config:/etc/krakend \
  krakend-ce:latest

# Ver logs
docker logs -f krakend

# Detener el contenedor
docker stop krakend
docker rm krakend
```

## Verificación

Una vez que el contenedor esté en ejecución, puedes verificar que KrakenD funciona correctamente:

```bash
# Health check
curl http://localhost:8080/__health

# Endpoint de ejemplo (usando JSONPlaceholder)
curl http://localhost:8080/example

# Endpoint de status
curl http://localhost:8080/status
```

## Configuración Personalizada

El archivo de configuración principal se encuentra en `config/krakend.json`. Puedes modificarlo según tus necesidades:

- **Endpoints**: Define tus rutas y backends
- **Timeout**: Ajusta los tiempos de espera
- **Cache**: Configura el TTL del cache
- **Plugins**: Añade plugins personalizados en el directorio `plugins/`

### Ejemplo de modificación de configuración:

```json
{
  "endpoint": "/api/users/{id}",
  "method": "GET",
  "backend": [
    {
      "url_pattern": "/users/{id}",
      "host": ["https://api.example.com"],
      "method": "GET"
    }
  ]
}
```

## Uso de Plugins Personalizados

Si tienes plugins Go personalizados:

1. Coloca tus archivos `.so` compilados en el directorio `plugins/`
2. Referéncialos en tu `krakend.json`:

```json
{
  "extra_config": {
    "plugin/http-server": {
      "name": ["my-plugin"],
      "my-plugin": {
        "path": "/opt/krakend/plugins/my-plugin.so"
      }
    }
  }
}
```

## Características del Build

### Multi-stage Build
- **Stage 1 (builder)**: Clona el repositorio y compila KrakenD con Go 1.22
- **Stage 2 (runtime)**: Imagen ligera Alpine con solo el binario compilado

### Ventajas
- Imagen final pequeña (~30MB)
- Build desde la última versión del código fuente
- Incluye todas las dependencias necesarias
- Configuración flexible mediante volúmenes

## Variables de Entorno

Puedes configurar las siguientes variables en el `docker-compose.yml`:

- `KRAKEND_PORT`: Puerto en el que escucha KrakenD (default: 8080)
- `FC_ENABLE`: Habilita Flexible Configuration (default: 1)
- `FC_SETTINGS`: Ruta a archivos de settings
- `FC_PARTIALS`: Ruta a archivos parciales de configuración

## Reconstruir desde una versión específica

Para compilar una versión específica de KrakenD, modifica el Dockerfile:

```dockerfile
# En lugar de usar master, especifica un tag
RUN git clone --branch v2.6.0 https://github.com/krakend/krakend-ce.git .
```

## Troubleshooting

### El contenedor no inicia
```bash
# Verifica los logs
docker-compose logs krakend

# Valida la configuración
docker run --rm -v $(pwd)/config:/etc/krakend krakend-ce:latest check -c /etc/krakend/krakend.json
```

### Error de compilación
```bash
# Limpia y reconstruye sin cache
docker-compose build --no-cache
```

### Puerto ya en uso
```bash
# Cambia el puerto en docker-compose.yml
ports:
  - "8081:8080"  # Usa 8081 en el host
```

## Recursos Adicionales

- [Documentación oficial de KrakenD](https://www.krakend.io/docs/overview/introduction/)
- [Repositorio GitHub de KrakenD CE](https://github.com/krakend/krakend-ce)
- [Ejemplos de configuración](https://github.com/krakend/krakend-ce/tree/master/examples)

## Licencia

KrakenD CE está bajo licencia Apache 2.0
