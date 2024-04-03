FROM golang:1.22

WORKDIR /bot

# Download all dependencies.
COPY go.mod go.sum ./
RUN go mod download
RUN go mod verify

COPY . .

RUN GOOS=linux go build -o bin

CMD [ "./bin" ]