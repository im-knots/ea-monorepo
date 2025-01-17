# Eru Labs Webpage
erulabs.ai webpage with entry to Ea Platform

## Requirements
- Docker
- curl/postman for cli testing
- A web browser for UI testing

## Build and run locally
### Build and run the image
```bash
$ docker build -t eru-labs-brand-www-frontend .
$ docker run -p 8080:8080 eru-labs-brand-www-frontend
```

### Verify the container is running
Visit `localhost:8080` in a browser or by running
```bash
$ curl localhost:8080
```

