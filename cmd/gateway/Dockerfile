FROM golang:latest

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

EXPOSE 50050

# go run 명령어로 실행
CMD ["go", "run", "./gateway/main.go"]