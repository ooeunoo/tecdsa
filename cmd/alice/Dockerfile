FROM golang:latest

WORKDIR /app


############################### REMOVE: production #######################################
RUN go install github.com/githubnemo/CompileDaemon@latest
############################### REMOVE: production #######################################

COPY go.mod go.sum ./
RUN go mod download

COPY . .

############################### REMOVE: production #######################################
CMD CompileDaemon --build="go build -o main ./cmd/alice" --command="./main"
############################### REMOVE: production #######################################


############################### REMOVE_COMMENT: production #######################################
# RUN go build -o alice ./cmd/alice

EXPOSE 50052

# CMD ["./alice"]
############################### REMOVE_COMMENT: production #######################################
