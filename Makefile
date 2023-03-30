NAME = wechat-openai

build:
	go build -trimpath -o bin/$(OS)/$(NAME)

build-linux:
	GOOS=linux GOARCH=amd64 go build -trimpath -o bin/linux/$(NAME)

build-image:
	docker build -t $(NAME):latest .
