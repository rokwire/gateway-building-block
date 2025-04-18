FROM golang:1.22-bullseye as builder

ENV CGO_ENABLED=0

#RUN apk add --no-cache --update make git

RUN mkdir /app
WORKDIR /app
# Copy the source from the current directory to the Working Directory inside the container
COPY . .
RUN make

FROM alpine:3.21.3

#we need timezone database + certificates
RUN apk add --no-cache tzdata ca-certificates

COPY --from=builder /app/bin/application /
COPY --from=builder /app/assets/assets.json /assets/assets.json
COPY --from=builder /app/driver/web/docs/gen/def.yaml /driver/web/docs/gen/def.yaml

COPY --from=builder /app/driver/web/client_permission_policy.csv /driver/web/client_permission_policy.csv
COPY --from=builder /app/driver/web/client_scope_policy.csv /driver/web/client_scope_policy.csv
COPY --from=builder /app/driver/web/admin_permission_policy.csv /driver/web/admin_permission_policy.csv
COPY --from=builder /app/driver/web/bbs_permission_policy.csv /driver/web/bbs_permission_policy.csv
COPY --from=builder /app/driver/web/tps_permission_policy.csv /driver/web/tps_permission_policy.csv
COPY --from=builder /app/driver/web/system_permission_policy.csv /driver/web/system_permission_policy.csv

COPY --from=builder /app/vendor/github.com/rokwire/core-auth-library-go/v3/authorization/authorization_model_scope.conf /app/vendor/github.com/rokwire/core-auth-library-go/v3/authorization/authorization_model_scope.conf
COPY --from=builder /app/vendor/github.com/rokwire/core-auth-library-go/v3/authorization/authorization_model_string.conf /app/vendor/github.com/rokwire/core-auth-library-go/v3/authorization/authorization_model_string.conf

COPY --from=builder /etc/passwd /etc/passwd

ENTRYPOINT ["/application"]
