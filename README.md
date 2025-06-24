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
curl -L https://github.com/egor-lukin/ssl-proxy/archive/refs/tags/v0.0.1.tar.gz -o ssl-proxy.tar.gz
tar -xzf ssl-proxy.tar.gz
sudo mv ssl-proxy /usr/local/bin/ssl-proxy 
sudo chmod +x /usr/local/bin/ssl-proxy


```

## Usage

- Prepare proxy

``` sh
ssl-proxy init --destination=turbositetest.com --proxy=proxy1.turbositetest.com --email='mail@egorlukin.me'

./ssl-proxy init --destination=turbositetest.com --proxy=proxy1.turbositetest.com --email='mail@egorlukin.me'

159.203.188.84

curl --resolve turbologotest7.egorlukin.me:443:159.203.188.84 https://turbologotest7.egorlukin.me                 

openssl s_client -connect turbologotest7.egorlukin.me:443 -tls1_2
openssl s_client -connect 159.203.188.84:443 -tls1_2
```

lego --email="mail@egorlukin.me" --domains="proxy1.turbositetest.com" --http run

[proxy1.turbositetest.com] invalid authorization: acme: error: 400 :: urn:ietf:params:acme:error:connection :: 159.203.188.84: Connection refused

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
git tag v0.0.2
git push origin v0.0.2
```
