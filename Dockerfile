FROM golang:1.20.1-buster as builder

ENV CGO_ENABLED=0

RUN mkdir /gateway-app
WORKDIR /gateway-app
# Copy the source from the current directory to the Working Directory inside the container
COPY . .
RUN make

FROM alpine:3.16.2

#we need timezone database
RUN apk --no-cache add tzdata

COPY --from=builder /gateway-app/bin/apigateway /
COPY --from=builder /gateway-app/docs/swagger.yaml /docs/swagger.yaml

COPY --from=builder /gateway-app/assets/assets.json /assets/assets.json

COPY --from=builder /gateway-app/driver/web/authorization_model.conf /driver/web/authorization_model.conf
COPY --from=builder /gateway-app/driver/web/authorization_policy.csv /driver/web/authorization_policy.csv

COPY --from=builder /etc/passwd /etc/passwd

#we need timezone database
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo 

ENTRYPOINT ["/apigateway"]
