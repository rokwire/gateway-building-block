# Gateway Building Block

The Gateway Building Block provides access to external systems for the Rokwire platform.

## Documentation
The functionality provided by this application is documented in the [Wiki](https://github.com/rokwire/gateway-building-block/wiki).

The API documentation is available here: https://api.rokwire.illinois.edu/gateway/api/doc/ui/index.html

## Set Up

### Prerequisites

MongoDB v4.2.2+

Go v1.16+

### Environment variables
The following Environment variables are supported. The service will not start unless those marked as Required are supplied.

Name|Format|Required|Description
---|---|---|---
GATEWAY_PORT | < int > | yes | The port number of the listening port
GATEWAY_HOST | < url > | yes | URL where this application is being hosted
GATEWAY_MONGO_AUTH | <mongodb://USER:PASSWORD@HOST:PORT/DATABASE NAME> | yes | MongoDB authentication string. The user must have read/write privileges.
GATEWAY_MONGO_DATABASE | < string > | yes | MongoDB database name e.g dining_db
GATEWAY_MONGO_TIMEOUT | < int > | no | MongoDB timeout in milliseconds. Defaults to 500
GATEWAY_LAUNDRY_APIKEY | < string > | yes | API Key for laundry view information
GATEWAY_LAUNDRY_APIURL | < url > | yes | Base URL for Laundry view apis
GATEWAY_LAUNDRYSERVICE_APIKEY | < string > | yes | API key for calling the laundry service apis
GATEWAY_LAUNDRYSERVICE_APIURL | < url > | yes | Base URL for the laundry service API endpoints
GATEWAY_LAUNDRYSERVICE_BASICAUTH | < string > | yes | Token for calling the laundry service apis
GATEWAY_WAYFINDING_APIKEY| < string > | yes | API Key used for calling location api end points
GATEWAY_WAYFINDING_APIURL | < url > | yes | Base URL for building location API endpoints
GATEWAY_CORE_HOST | < url > | yes | Core BB host URL
GATEWAY_CONTACTINFO_APIKEY | < string > | yes | API key used to access campus student information apis
GATEWAY_CONTACTINFO_ENDPOINT | < url > | yes | Base URL to the campus student information apis
GATEWAY_BASE_URL | < url > | yes | Base URL for the gateway
GATEWAY_CORE_BB_BASE_URL | < url > | yes | Base URL for the core

### Run Application

#### Run locally without Docker

1. Clone the repo (outside GOPATH)

2. Open the terminal and go to the root folder
  
3. Make the project  
```
$ make
...
▶ building executable(s)… 1.9.0 2020-08-13T10:00:00+0300
```

4. Run the executable
```
$ ./bin/apigateway
```

#### Run locally as Docker container

1. Clone the repo (outside GOPATH)

2. Open the terminal and go to the root folder
  
3. Create Docker image  
```
docker build -t a .
```
4. Run as Docker container
```
docker-compose up
```

#### Tools

##### Run tests
```
$ make tests
```

##### Run code coverage tests
```
$ make cover
```

##### Run golint
```
$ make lint
```

##### Run gofmt to check formatting on all source files
```
$ make checkfmt
```

##### Run gofmt to fix formatting on all source files
```
$ make fixfmt
```

##### Cleanup everything
```
$ make clean
```

##### Run help
```
$ make help
```

##### Generate Swagger docs
```
$ make swagger
```

### Test Application APIs

Verify the service is running as calling the get version API.

#### Call get version API

curl -X GET -i https://api-dev.rokwire.illinois.edu/gateway/api/version

Response
```
0.1.2
```

## Contributing
If you would like to contribute to this project, please be sure to read the [Contributing Guidelines](CONTRIBUTING.md), [Code of Conduct](CODE_OF_CONDUCT.md), and [Conventions](CONVENTIONS.md) before beginning.

### Secret Detection
This repository is configured with a [pre-commit](https://pre-commit.com/) hook that runs [Yelp's Detect Secrets](https://github.com/Yelp/detect-secrets). If you intend to contribute directly to this repository, you must install pre-commit on your local machine to ensure that no secrets are pushed accidentally.

```
# Install software 
$ git pull  # Pull in pre-commit configuration & baseline 
$ pip install pre-commit 
$ pre-commit install
```