up: dockerbuild down
	@echo ">> running docker container..."
	docker run -d --name ccserver_service -p 80:80 -p 443:443 ccserver_api
	@echo

local_up: build
	@echo ">> running application binary..."
	./ccserver_local
	@echo

dockerbuild: build_linux
	@echo ">> building docker contaier..."
	docker build -t ccserver_api .
	@echo

build_linux:
	@echo ">> building linux application binary..."
	GOOS=linux GAORCH=amd64 go build -o ccserver_linux
	@echo

build:
	@echo ">> building application binary..."
	go build -o ccserver_local
	@echo

down:
	@echo ">> stopping current service if its running..."
	(docker stop ccserver_service || true) && (docker rm ccserver_service || true)
	@echo
