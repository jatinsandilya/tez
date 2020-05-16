<div align-"center">
<h1>
<img src="./stargazer.svg" alt="Logo" width="50" style="background: #feeb4e; border-radius:50%;">
<br/>
Tez is an HTTP interface for Redis, written in Go
</h1>
</div>

#### Steps for running service locally:

Build image:

`docker build -t tez .` 

Check if the image is built:

`docker images`

Run container with image: 

`docker run  -e REDIS_HOST='host.docker.internal' -it -p 3000:3000 tez`

*Note*: 

1.  Ensure docker for pc/mac is installed.  
2.  Ensure redis is running locally with `rejson` module. (https://github.com/RedisJSON/RedisJSON)  
3. `host.docker.internal` resolves to `localhost` 

Check if your running is running on docker with: 

`docker ps` 