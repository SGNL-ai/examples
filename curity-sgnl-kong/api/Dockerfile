FROM golang:1.18.5-alpine AS GO_BUILD
COPY . /server
WORKDIR /server/server
RUN go mod download github.com/gorilla/mux
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /go/bin/server/server

FROM gcr.io/distroless/static:nonroot
WORKDIR /app
COPY --from=GO_BUILD /go/bin/server/server ./server
USER nonroot:nonroot
EXPOSE 8080
CMD ["./server"] 