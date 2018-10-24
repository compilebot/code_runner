FROM golang:1.11.1-alpine3.8 as build-env
# All these steps will be cached
RUN apk update && apk add git && apk add ca-certificates
RUN mkdir /code_runner
WORKDIR /code_runner
COPY go.mod .
COPY go.sum .

# Get dependancies - will also be cached if we won't change mod/sum
RUN go mod download
# COPY the source code as the last step
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o /go/bin/code_runner

FROM scratch 
COPY --from=build-env /go/bin/code_runner /go/bin/code_runner
ENTRYPOINT ["/go/bin/code_runner"]
