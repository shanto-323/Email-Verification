FROM golang:alpine AS build
WORKDIR /email-auth
COPY email-auth/go.mod email-auth/go.sum ./
COPY email-auth ./
RUN go build -o app .

FROM alpine:3.21
WORKDIR /usr/bin
COPY --from=build /email-auth/app .
EXPOSE 8080
CMD ["app"]