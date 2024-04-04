FROM golang:1.22.0
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY *.go ./
ADD web/ ./web/

ENV TODO_PORT=7540
ENV TODO_PASSWORD=12345
ENV TODO_DBFILE="./scheduler.db"

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /my_app
CMD [ "/my_app" ]

EXPOSE ${TODO_PORT}