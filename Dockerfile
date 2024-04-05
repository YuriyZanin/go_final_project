FROM golang:1.22.0
WORKDIR /app
ARG DATA_DIR
VOLUME [ ${DATA_DIR} ]
COPY go.mod go.sum ./
RUN go mod download
COPY *.go ./
ADD web/ ./web/
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /my_app
CMD [ "/my_app" ]