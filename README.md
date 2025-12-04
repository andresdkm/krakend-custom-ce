KrakenD Custom
------------

## Descargar la version krakend-ce (community)
    // ponlo en la misma raiz que este proyecto ej: ~/proyects/
    https://github.com/krakend/krakend-ce
    //usar un release ej: v2.12.0

## Estas variables de Makefile deben coincidir con las de ../krakend-ce/Makefile
    GOLANG_VERSION := 1.25.3
    ALPINE_VERSION := 3.21
    VERSION := 2.12.0

## Backup del go.mod original para reusarlo en cada build
    cp -a go.mod go.bck

# Krakend Custome

## Agregar modulos nuevos en goBuilderDeps de este repo luego este archivo se agregara al go.mod de la version de kraken-ce para que sea compatible, Ejemplo:

    require (
        github.com/cespare/xxhash/v2 v2.3.0 // indirect
        github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
        github.com/google/uuid v1.6.0 // indirect
        github.com/redis/go-redis/v9 v9.17.1 // indirect
    )


## Ejecutar esto si requiero una nueva dependencia unicamente o es la primera vez que descargo el repo

    make docker
    make docker-builder
    make up

## Contenedor krakend-builder

    chmod +x scripts/build-plugins.sh
    ./scripts/build-plugins.sh 

## Contenedor krakend-custom 
    restart //por ahora
