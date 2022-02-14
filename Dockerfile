FROM golang:alpine
WORKDIR /go/src/chat-room
COPY . .
RUN go build -o /go/bin/chat-room cmd/main.go


CMD [ "/go/bin/chat-room" ]