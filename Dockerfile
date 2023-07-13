FROM golang:1.20-alpine

WORKDIR /yandex-lavka-service

COPY go.mod .
COPY go.sum .
RUN go mod download && go mod verify

COPY . ./

WORKDIR /yandex-lavka-service/tests

RUN go test

WORKDIR /yandex-lavka-service

RUN go build -o /yandex-lavka-service/build/yandex-lavka-service /yandex-lavka-service/cmd/yandex_lavka_service/

RUN  go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

ENV POSTGRES_DSN="postgresql://postgres:password@host.docker.internal:5432"

ENV SSL_MODE="disable"

EXPOSE 8080

ENTRYPOINT [ "/bin/sh", "-c", "migrate -path internal/store/migrations -database \"${POSTGRES_DSN}?sslmode=${SSL_MODE}\" -verbose up && /yandex-lavka-service/build/yandex-lavka-service" ]