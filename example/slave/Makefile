up: dockerbuild down
	@echo ">> running docker container..."
	docker run -d --name slave_application slave_app
	@echo

local_up: build
	@echo ">> running application binary..."
	./slave_app_local
	@echo

dockerbuild: build_linux
	@echo ">> building docker contaier..."
	docker build -t slave_app .
	@echo

build_linux:
	@echo ">> building linux application binary..."
	GOOS=linux GAORCH=amd64 go build -o slave_app_linux
	@echo

build:
	@echo ">> building application binary..."
	go build -o slave_app_local
	@echo

down:
	@echo ">> stopping current service if its running..."
	(docker stop slave_application || true) && (docker rm slave_application || true)
	@echo
