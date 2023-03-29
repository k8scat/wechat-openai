NAME = wechat-openai

OS = $(shell uname | tr '[:upper:]' '[:lower:]')
ARCH = $(shell uname -m)

build:
	GOOS=$(OS) GOARCH=$(ARCH) go build -trimpath -o bin/$(OS)/$(NAME)

build-linux:
	GOOS=linux GOARCH=amd64 go build -trimpath -o bin/linux/$(NAME)

build-image:
	docker build -t $(NAME):latest .
