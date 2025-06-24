# Ssl Proxy Server

## TODO

- [X] generate token for auth
- [X] add endpoint PUT /servers
- [X] add endpoint POST /servers
- [X] add sqlite for keeping zones
- [X] use token for auth
- [X] add reverse proxy
- [X] create ssl for proxy server
- [X] add github action
- [X] add proxy health_check
- [-] zero downtime deploy

## Install

```sh
curl -L https://github.com/egor-lukin/ssl-proxy/releases/latest/download/ssl-proxy -o /usr/local/bin/ssl-proxy && chmod +x /usr/local/bin/ssl-proxy
```

## Usage

- Prepare proxy

``` sh
ssl-proxy init --destination=turbologo.com --proxy=proxy1.turbologo.com --email='mail@egorlukin.me'
```

- Run proxy server

``` sh
ssl-proxy run
```

## Internal Api

- POST /__internal/servers

- PUT /__internal/servers

- GET /__internal/health_check

## Release

To create a new release and trigger the GitHub Actions workflow to build and upload the binary, create and push a git tag with the desired version. For example:

```sh
git tag v0.0.1
git push origin v0.0.1
```
