#### Steps for running service locally:

Build image:

`docker build -t raas .(root of code base) ` 

Check if the image is built:

`docker images`

Run container with image: 

`docker run  -e REDIS_HOST='host.docker.internal' -it -p 3000:3000 raas`

*Note*: 


1.  Ensure docker for pc/mac is installed.  
2.  Ensure redis is running locallly with `rejson` module.  
3. `host.docker.internal` resolves to `localhost` 

Check if your running is running on docker with: 

`docker ps` 


#### Steps for deploying to beta (till jenkins is setup):


1. Tag latest to a new version: `docker tag raas:latest asia.gcr.io/scootsy-betaout-09041979/raas:v<N>` where N is the new version not available in container registry. 

2. Push built image to container registry: `docker push asia.gcr.io/scootsy-betaout-09041979/raas:v<N>`
3. Start using kubectl for beta: `gcloud container clusters get-credentials beta-scootsy --zone asia-south1-a --project scootsy-betaout-09041979`
4. Deploy latest image to beta: `kubectl set image deployment/raas-api raas=asia.gcr.io/scootsy-betaout-09041979/raas:v<N>` 

