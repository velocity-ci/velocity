+++
date = "2012-08-15T22:32:09+01:00"
title = "Getting Started with Docker Compose"
+++

Docker Compose is the quickest way to get up and running to try Velocity CI out!

1. Create a `docker-compose.yml` with the following contents.
<script src="https://gist-it.appspot.com/https://github.com/velocity-ci/velocity/blob/master/docs/velocity/content/docs/getting-started/docker-compose.yml"></script>

2. Run ```docker-compose up```
3. Open http://localhost:4200 in your web browser and log in with the following credentials:
```
username: admin
password: velocity_local1234
```

--- 

Alternatively, you can do 1. and 2. in one bash line:
```
curl -LO https://raw.githubusercontent.com/velocity-ci/velocity/master/docs/velocity/content/docs/getting-started/docker-compose.yml && docker-compose up
```
