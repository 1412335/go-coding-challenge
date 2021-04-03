# export GO111MODULE=on

# install
.PHONY: install
install:
	go get \
		github.com/golang/protobuf/protoc-gen-go \
		github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway \
		github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger \
		github.com/mwitkow/go-proto-validators/protoc-gen-govalidators \
		github.com/rakyll/statik

# gen cert
.PHONY: gen-cert
gen-cert:
	cd ./cert; sh gen.sh; cd ../

# gen stubs
.PHONY: gen
gen:
	@echo "====gen stubs===="
	sh ./script/gen-proto.sh

# gen openapi
.PHONY: gen-openapi
gen-openapi:
	@echo "====gen openapi===="
	sh ./script/gen-openapi.sh

# 
.PHONY: grpc
grpc: clean
	@echo "====Run grpc server with docker===="
	# docker-compose up -d mysql
	# sleep 20s
	docker-compose up -d --build


# https://github.com/ktr0731/evans
# Evans cli: calling grpc service (reflection.Register(server))
.PHONY: cli
cli:
	evans -r repl -p 8080


# gofmt
.PHONY: fmt
fmt:
	go fmt -mod=mod $(go list ./... | grep -v /pkg/api/)

# go-lint
.PHONY: lint
lint: fmt
	golangci-lint run $(go list ./... | grep -v /vendor/)

# go-lint
.PHONY: test
test: lint
	go test -v $(go list ./... | grep -v /vendor/)


# cleaning
.PHONY: clean
clean:
	@echo "====cleaning env==="
	docker-compose down -v --remove-orphans
	rm -rf ./docker/mysql/data
	# docker system prune -af --volumes
	# docker rm $(docker ps -aq -f "status=exited")