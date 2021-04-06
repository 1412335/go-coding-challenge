# 2021 Moneyforward - Go Coding Challenge

## Overrall

- API [user-account-transaction service](./api/proto/user-service.proto) with CRUD functionality
- Database: Postgres
- [Unit test](./service/user/user-service_test.go)
- Config service running with [config.yml](./service/user/config.yml)
- Set up service with [docker-compse.yml](./docker-compose.yml)
- CI with github action
- gRPC server running at http://127.0.0.1:9090
- Running gRPC-gateway at [http://127.0.0.1:8000](http://127.0.0.1:8000/openapi-ui)
- [Makefile](./Makefile)

## Install

```sh
make install
```

## Testing

```sh
make run
```

## More commands

```sh
# Gen proto
make gen

# test
make test

# Cli with evans
make cli
```

## Testing w OpenAPI

- [http://127.0.0.1:8000/openapi-ui](http://127.0.0.1:8000/openapi-ui)

## TODO

- Split user-service into multi services: User, Account, Transaction
- Manage user with role
- Add gRPC client & its unittest
- Cache with redis
- Tracing
- Audit
- Enable secure TLS