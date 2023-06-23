FROM golang:1.20.5

WORKDIR /app

COPY  . ./
RUN go mod download

RUN go build -o ex1App main/main.go

EXPOSE 8080

CMD [ "./ex1App" ]

