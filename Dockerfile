FROM golang:1.22.1

RUN mkdir /server

WORKDIR /server

COPY ./ ./

RUN go env -w GO111MODULE=on

RUN go mod download

RUN go build -o go-message.exe ./cmd/api/

CMD [ "./go-message.exe" ]
