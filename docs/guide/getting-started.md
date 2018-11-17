---
sidebarDepth: 2
---
# Getting Started

## CLI (vcli)


### Installation

#### Apple Macintosh and Linux

1. Run the following in a terminal to download the pre-compiled binary for your distribution:
``` bash
bash <(curl -s https://raw.githubusercontent.com/velocity-ci/velocity/master/backend/deployments/vcli/install.sh)
```

#### Microsoft Windows
We don't offer Windows support right now.


## Architect & Web UI

### Docker Compose

1. Create a `docker-compose.yml` with the following contents:
``` yaml
# docker-compose.yml
---
version: '3'

services:
  architect:
    image: civelocity/architect:latest
    environment:
      ADMIN_PASSWORD: velocity_local1234
      JWT_SECRET: jwt_local1234
      BUILDER_SECRET: builder_secret1234
    ports:
    - "80:80"
    volumes:
    - "./architect_data:/opt/velocityci"

  builder:
    image: civelocity/builder:latest
    environment:
      BUILDER_SECRET: builder_secret1234
      ARCHITECT_ADDRESS: http://architect
    volumes:
      - "/opt/velocityci:/opt/velocityci"
      - "/var/run/docker.sock:/var/run/docker.sock"

  web:
    image: civelocity/web:latest
    environment:
      ARCHITECT_ENDPOINT: "http://localhost/v1"
    ports:
    - "4200:80"
```
2. Run `docker-compose up`
3. Open [localhost:4200](http://localhost:4200) in your web browser and log in with the following credentials:
```
username: admin
password: velocity_local1234
```

--- 

Alternatively, you can do 1. and 2. in one bash line:
``` bash
curl -LO https://raw.githubusercontent.com/velocity-ci/velocity/master/backend/deployments/docker-compose/docker-compose.yml \
&& docker-compose up
```

### Kubernetes

### Amazon ECS