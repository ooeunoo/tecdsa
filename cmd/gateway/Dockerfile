FROM golang:latest

WORKDIR /app

############################### REMOVE: production #######################################
RUN go install github.com/githubnemo/CompileDaemon@latest
############################### REMOVE: production #######################################


COPY go.mod go.sum ./
RUN go mod download

COPY . .

############################### REMOVE: production #######################################
CMD CompileDaemon --build="go build -o main ./cmd/gateway" --command="./main"
############################### REMOVE: production #######################################


############################### REMOVE_COMMENT: production #######################################
# RUN go build -o gateway ./cmd/gateway

# EXPOSE 8080

# CMD ["./gateway"]
############################### REMOVE_COMMENT: production #######################################
