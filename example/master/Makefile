up: dockerbuild down
	@echo ">> running docker container..."
	docker run -d --name master_service -p 80:80 -p 443:443 master_api
	@echo

local_up: build
	@echo ">> running application binary..."
	./master_local
	@echo

dockerbuild: build_linux
	@echo ">> building docker contaier..."
	docker build -t master_api .
	@echo

build_linux:
	@echo ">> building linux application binary..."
	GOOS=linux GAORCH=amd64 go build -o master_linux
	@echo

build:
	@echo ">> building application binary..."
	go build -o master_local
	@echo

down:
	@echo ">> stopping current service if its running..."
	(docker stop master_service || true) && (docker rm master_service || true)
	@echo
