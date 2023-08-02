# svc-grpc

## Config example

```yaml
grpcServers:
  - name: public-with-http
    address: ":3016"
  - name: public
    address: ":3017"
    middlewares:
      logging:
        loggedFields:
          OpenAccount:
            - server
        trimmedFields:
          OpenAccount:
            - server
    http:
      address: ":3019"
```
