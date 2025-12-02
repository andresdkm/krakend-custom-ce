module expedia-provider

go 1.25.3

require hotels-common v0.0.0

require redis v0.0.0

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/redis/go-redis/v9 v9.17.1 // indirect
)

replace hotels-common => ../hotels-common

replace redis => ../redis
