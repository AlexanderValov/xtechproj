FROM golang:1.18.0-alpine

RUN go version
ENV GOPATH=/

COPY ./ ./

RUN go mod download
RUN go build -o server ./cmd/app/main.go

EXPOSE 8000
CMD ["./XTechProject"]
