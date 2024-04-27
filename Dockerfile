FROM golang:1.22-alpine as builder
WORKDIR /
COPY . ./
RUN go mod download


RUN go build -o /tmp-websocket


FROM alpine
COPY --from=builder /tmp-websocket .


EXPOSE 80
CMD [ "/tmp-websocket" ]