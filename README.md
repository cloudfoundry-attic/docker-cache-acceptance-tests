# Docker Cache Acceptance Tests (DCATs)

This test suite exercises the [Docker Cache](https://github.com/cloudfoundry-incubator/diego-docker-cache) when deployed
alongside [CF Runtime](https://github.com/cloudfoundry/cf-release) and [Diego](https://github.com/cloudfoundry-incubator/diego-release) .

## Usage

### Getting the tests

To get these tests, you can either `git clone` this repo:

```bash
git clone https://github.com/cloudfoundry-incubator/docker-cache-acceptance-tests $GOPATH/src/github.com/cloudfoundry-incubator
cd $GOPATH/src/github.com/cloudfoundry-incubator
go get -t -v ./...
```

 or `go get` it:

 ```bash
 go get -t -v github.com/cloudfoundry-incubator/docker-cache-acceptance-tests/...
 ```

Either way, we assume you have Golang setup on your workstation.

### Test setup

To run the Diego Acceptance tests, you will need:
- a running CF deployment
- a running Diego deployment
- a running Docker Cache deployment
- credentials for an Admin user
- an environment variable `CONFIG` which points to a `.json` file that contains the application domain
- the [cf CLI](https://github.com/cloudfoundry/cli)
- ginkgo testing framework
- Docker executable

#### Configuration

The following commands will setup the `CONFIG` for a [bosh-lite](https://github.com/cloudfoundry/bosh-lite)
installation. Replace credentials, URLs and the path to Docker as appropriate for your environment.

```bash
cd $GOPATH/src/github.com/cloudfoundry-incubator/docker-cache-acceptance-tests
cat > integration_config.json <<EOF
{
  "api": "api.10.244.0.34.xip.io",
  "admin_user": "admin",
  "admin_password": "admin",
  "apps_domain": "10.244.0.34.xip.io",
  "docker_registry_address": "10.244.2.6:8080",
  "docker_executable": "docker",
  "docker_private_image": "private-docker-app"
  "docker_user": "user",
  "docker_password": "password",
  "docker_email": "email@example.com",
  "skip_ssl_validation": true
}
EOF
export CONFIG=$PWD/integration_config.json
```

**Note:** The tests require that you have a copy of the public docker image `cloudfoundry/diego-docker-app:latest`, stored in a private repo in Docker Hub. Therefore you need to provide the Docker credentials (user, password and email) for access to the image in the config above.

#### Install ginkgo:

Install the testing framework with:

```
go install github.com/onsi/ginkgo/ginkgo
```

#### Setup docker (OSX)

The tests use [Docker](https://www.docker.com/) to check the Cache functionality. [Install Docker](https://docs.docker.com/installation) 

In case you use boot2docker you will need to allow access to the insecure registry by adding your registry address to `/var/lib/boot2docker/profile`. For example:

```
EXTRA_ARGS='--insecure-registry 10.244.2.6:8080'
```

#### Enable Docker Feature Flag

In order to run the docker tests you need to enable Docker support in Diego as follows:

```cf enable-feature-flag diego_docker```

After the tests complete you may disable Docker support with:

```cf diable-feature-flag diego_docker```


### Running the tests

After correctly setting the `CONFIG` environment variable, the following command will run the tests:

```
./bin/test
```

### Running as BOSH errand

To deploy the tests as BOSH errand you have to:

1. Copy the public docker image `cloudfoundry/diego-docker-app:latest`, in your private repo in Docker Hub.
1. Add the Docker credentials (user, password and email) for access to the image in `$GOPATH/src/github.com/cloudfoundry-incubator/docker-cache-acceptance-tests/templates/bosh-lite.yml`
1. Deploy the test errand

```
cd $GOPATH

bosh deployment $GOPATH/src/github.com/cloudfoundry-incubator/docker-cache-acceptance-tests/templates/bosh-lite.yml
bosh -n deploy
```

To start the tests:

```
bosh run errand docker_acceptance_tests
```
