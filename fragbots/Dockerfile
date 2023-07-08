FROM golang:alpine

WORKDIR /app

COPY go.* ./
RUN go mod download

COPY . ./
RUN ls

RUN go build -o /fragbots

CMD ["/fragbots"]
