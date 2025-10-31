FROM golang:tip-alpine3.22

WORKDIR /app

COPY  ./* /app/

RUN go mod tidy

EXPOSE 8080

ENTRYPOINT [ "go", "run", "." ]