# Hello API

Hello API exposes a simple service over gRPC and OpenAPI to record visitors and visits.


## Prerequisites

* gcloud CLI
* Docker
* Go 1.8

## Development

Run `make help` or `make`


## API

The API can be explored by pointing your browser to [https://localhost:9999/#/Hello](https://localhost:9999/#/Hello).

When using `curl`, an Accept header specifying `application/json` is required in order to hit the HTTP API. Otherwise, it is going to
attempt to serve Swagger UI.

```shell
curl --insecure -H "Accept: application/json" https://localhost:9999/hello/camilo | jq .
curl --insecure -H "Accept: application/json" https://localhost:9999/counts | jq .
curl -v --insecure -X DELETE -H "Accept: application/json" https://localhost:9999/counts | jq .
```

## Technical details

* I used gRPC and Open API to generate the service from `hello.proto` and implemented `hello_service.go` and `hello_service_test.go`
* I added Swagger UI to make it more convenient to explore the HTTP API.
* Third-party dependencies are kept to the minimum possible.
* The API also supports HTTP2 and TLS 1.2
* A very minimal `Dockerfile` was also added.
* Releasing a binary of the service was automated with Github Releases: https://github.com/c4milo/hello/releases
* Dependencies are vendored using `govendor`
* The total effort, including tests, docs and deployment was roughly 4 hours. Most of the time was spent dealing with GCP.

## Deployment

A GKE cluster was created using GCP's web console, only 1 replica is launched as the service is managing state locally.
In order to deploy changes follow these steps:

1. Increase version number in [Makefile](https://github.com/c4milo/hello/tree/master/Makefile#L2)
2. Increase version number in [deployment/hello.yaml](https://github.com/c4milo/hello/tree/master/deployment/hello.yaml#L14) accordingly.
3. Run `make image-push`
4. Run `kubectl apply deployment/hello.yaml`
5. Verify by visiting [https://35.185.24.220/#/Hello](https://35.185.24.220/#/Hello)

## Things that can be improved further

* Manage state externally so the service can be scaled horizontally.
* Add compression
* Add rate limiting
* Add LetsEncrypt certificate and schedule K8S to periodically renew the certificate.
* Use domain name instead of IP address to access the service.

## Issues I ran across

* I spent a good deal of time trying to fix an authentication issue with gcloud as I had a previous project configured before.
* K8s does not seem to be able to register service health checks in GCP's network load balancer
