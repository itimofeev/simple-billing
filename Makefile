

run-env: stop-env
	docker network create billing

	docker run -d --network=billing  -p 5432:5432 -e POSTGRES_PASSWORD=password --name=db postgres:13
	docker run -d --network=billing -p 4222:4222 -p 8222:8222 -p 6222:6222 --name queue nats-streaming:0.21
	docker run -d --network=billing -p 8282:8282 --name queue-ui kuali/nats-streaming-console

stop-env:
	docker rm -f db queue queue-ui || true
	docker network rm billing || true

lint:
	#go get golang.org/x/tools/cmd/goimports
	goimports -w -local github.com/itimofeev/simple-billing internal cmd

	#curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.39.0
	GO111MODULE=on GL_DEBUG=debug L_DEBUG=linters_output GOPACKAGESPRINTGOLISTERRORS=1 golangci-lint -v run

test:
	go test --tags=load ./...