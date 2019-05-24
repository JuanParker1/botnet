up: dockerbuild down
	@echo ">> running docker container..."
	docker run -d --name bot_controller bot_image
	@echo

local_up: build
	@echo ">> running application binary..."
	./bot_local
	@echo

dockerbuild: build_linux
	@echo ">> building docker image..."
	docker build -t bot_image .
	@echo

build_linux:
	@echo ">> building linux application binary..."
	GOOS=linux GAORCH=amd64 go build -o bot_linux
	@echo

build:
	@echo ">> building application binary..."
	go build -o bot_local
	@echo

down:
	@echo ">> stopping current service if its running..."
	(docker stop bot_controller || true) && (docker rm bot_controller || true)
	@echo
