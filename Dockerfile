FROM golang:1.15-alpine3.12

ARG HTTPS_PREFIX

ENV GOPRIVATE "github.com/findy-network"

RUN apk update && \
    apk add git && \
    git config --global url."https://"${HTTPS_PREFIX}"github.com/".insteadOf "https://github.com/"

WORKDIR /work

COPY go.* ./
RUN go mod download

COPY . ./

RUN go build -o /go/bin/findy-agent-auth

FROM alpine:3.12

# override when running
ENV FAA_PORT "8888"
ENV FAA_AGENCY_ADDR "localhost"
ENV FAA_AGENCY_PORT "50051"
ENV FAA_DOMAIN "localhost"
ENV FAA_ORIGIN "http://localhost:8888"
ENV FAA_JWT_VERIFICATION_KEY "mySuperSecretKeyLol"

COPY --from=0 /go/bin/findy-agent-auth /findy-agent-auth

RUN echo '/findy-agent-auth \
    --port $FAA_PORT \
    --agency $FAA_AGENCY_ADDR \
    --gport $FAA_AGENCY_PORT \
    --domain $FAA_DOMAIN \
    --origin $FAA_ORIGIN \
    --jwt-secret $FAA_JWT_VERIFICATION_KEY' > /start-server.sh && chmod a+x /start-server.sh


ENTRYPOINT ["/bin/sh", "-c", "/start-server.sh"]
