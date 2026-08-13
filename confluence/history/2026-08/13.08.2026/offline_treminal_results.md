[denismatveev@c0der event_horizon]$ docker images | grep -E 'fulfillment|notification'
[denismatveev@c0der event_horizon]$ unset GOPROXY
[denismatveev@c0der event_horizon]$ (cd services/fulfillment && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -tags kafka -ldflags="-s -w" -o fulfillment-service ./cmd/main.go)
docker build -f Dockerfile.fulfillment.bin -t eastwesser/fulfillment:latest .
DEPRECATED: The legacy builder is deprecated and will be removed in a future release.
            Install the buildx component to build images with BuildKit:
            https://docs.docker.com/go/buildx/

Sending build context to Docker daemon  941.1MB
Step 1/5 : FROM alpine:3.20
 ---> bf8527eb54c3
Step 2/5 : RUN apk add --no-cache ca-certificates wget
 ---> Using cache
 ---> 5a4c374c6cb2
Step 3/5 : COPY services/fulfillment/fulfillment-service /fulfillment-service
 ---> 52c64446e89f
Step 4/5 : EXPOSE 9101
 ---> Running in d704c355ad1d
 ---> Removed intermediate container d704c355ad1d
 ---> 8f94fb0f5d75
Step 5/5 : CMD ["/fulfillment-service"]
 ---> Running in 0f3445795808
 ---> Removed intermediate container 0f3445795808
 ---> 671712e25c8b
Successfully built 671712e25c8b
Successfully tagged eastwesser/fulfillment:latest
[denismatveev@c0der event_horizon]$ (cd services/notification && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -tags kafka -ldflags="-s -w" -o notification-service ./cmd/main.go)
docker build -f Dockerfile.notification.bin -t eastwesser/notification:latest .
DEPRECATED: The legacy builder is deprecated and will be removed in a future release.
            Install the buildx component to build images with BuildKit:
            https://docs.docker.com/go/buildx/

Sending build context to Docker daemon  936.5MB
Step 1/5 : FROM alpine:3.20
 ---> bf8527eb54c3
Step 2/5 : RUN apk add --no-cache ca-certificates wget
 ---> Using cache
 ---> 5a4c374c6cb2
Step 3/5 : COPY services/notification/notification-service /notification-service
 ---> 381212cbaafa
Step 4/5 : EXPOSE 9102
 ---> Running in 9588f4c7d605
 ---> Removed intermediate container 9588f4c7d605
 ---> f75d240f18f3
Step 5/5 : CMD ["/notification-service"]
 ---> Running in b681df15c63e
 ---> Removed intermediate container b681df15c63e
 ---> 85828860936d
Successfully built 85828860936d
Successfully tagged eastwesser/notification:latest
[denismatveev@c0der event_horizon]$ docker push eastwesser/fulfillment:latest 2>&1 | tee ~/event_horizon/confluence/history/2026-08/13.08.2026/push-fulfillment.log
docker push eastwesser/notification:latest 2>&1 | tee ~/event_horizon/confluence/history/2026-08/13.08.2026/push-notification.log
The push refers to repository [docker.io/eastwesser/fulfillment]
8def9d8bd34b: Preparing
03b7fe2ad5a2: Preparing
08bc4e534116: Preparing
03b7fe2ad5a2: Pushed
08bc4e534116: Pushed
8def9d8bd34b: Pushed
latest: digest: sha256:912ca16f884be5df1482a923941bbb1646e5e50f45abd64bb49f4f08c9c3c193 size: 950
The push refers to repository [docker.io/eastwesser/notification]
7d2c098979b1: Preparing
03b7fe2ad5a2: Preparing
08bc4e534116: Preparing
03b7fe2ad5a2: Pushed
7d2c098979b1: Pushed
08bc4e534116: Pushed
latest: digest: sha256:c858b111d4544e166f1820bd35d35e9a7633cc2ab0cf453a01ffa23e7c818dbf size: 950
[denismatveev@c0der event_horizon]$ 