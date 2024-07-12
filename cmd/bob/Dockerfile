FROM golang:latest

WORKDIR /app

############################### REMOVE: production #######################################
RUN go install github.com/githubnemo/CompileDaemon@latest
############################### REMOVE: production #######################################

COPY go.mod go.sum ./
RUN go mod download

COPY . .

############################### REMOVE: production #######################################
CMD CompileDaemon --build="go build -o bob ./cmd/bob" --command="./bob"
############################### REMOVE: production #######################################


############################### REMOVE_COMMENT: production #######################################
# RUN go build -o bob ./cmd/bob

# EXPOSE 50051

# CMD ["./bob"]
############################### REMOVE_COMMENT: production #######################################
