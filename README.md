<p align="center">
    <img alt="Rabbit Logo" src="https://raw.githubusercontent.com/Clivern/Rabbit/master/assets/img/logo.png" height="80" />
    <h3 align="center">Rabbit</h3>
    <p align="center">A lightweight service that will build and store your go projects binaries.</p>
    <p align="center">
        <a href="https://godoc.org/github.com/clivern/rabbit"><img src="https://godoc.org/github.com/clivern/rabbit?status.svg"></a>
        <a href="https://travis-ci.org/Clivern/Rabbit"><img src="https://travis-ci.org/Clivern/Rabbit.svg?branch=master"></a>
        <a href="https://github.com/Clivern/Rabbit/releases"><img src="https://img.shields.io/badge/Version-0.1.2-red.svg"></a>
        <a href="https://goreportcard.com/report/github.com/Clivern/Rabbit"><img src="https://goreportcard.com/badge/github.com/Clivern/Rabbit"></a>
        <a href="https://github.com/Clivern/Rabbit/blob/master/LICENSE"><img src="https://img.shields.io/badge/LICENSE-MIT-orange.svg"></a>
    </p>
</p>

Rabbit is a lightweight service that will build and store your go projects binaries. Once a VCS system (github or bitbucket) notifies rabbit of a new release, it clones the project, builds different binaries and publish them.

<p align="center">
    <img src="https://raw.githubusercontent.com/Clivern/Rabbit/master/assets/img/diagram.png?v=0.0.2" />
</p>
<br/>
<p align="center">
    <h3 align="center">Screenshot</h3>
    <img src="https://raw.githubusercontent.com/Clivern/Rabbit/master/assets/img/screenshot.png?v=0.0.2" />
</p>
<br/>

## Documentation

### Development:

Rabbit uses [Go Modules](https://github.com/golang/go/wiki/Modules) to manage dependencies. First Create a prod config file.

```bash
$ git clone https://github.com/Clivern/Rabbit.git
$ cp config.dist.yml config.prod.yml
```

Then add your configs

```yml
# General App Configs
app:
    # Env mode (dev or prod)
    mode: dev
    # HTTP port
    port: 8080
    # App URL
    domain: http://127.0.0.1:8080
    # TLS configs
    tls:
        status: off
        pemPath: cert/server.pem
        keyPath: cert/server.key

# Redis Configs
redis:
    addr: localhost:6379
    password:
    db: 0

# Message Broker Configs
broker:
    # Broker driver (native or redis)
    driver: native
    # Native driver configs
    native:
        # Queue max capacity
        capacity: 50
        # Number of concurrent workers
        workers: 1
    # Redis configs
    redis:
        channel: rabbit

# Log configs
log:
    # Log level, it can be debug, info, warn, error, panic, fatal
    level: debug
    # output can be stdout or abs path to log file /var/logs/rabbit.log
    output: stdout
    # Format can be json or text
    format: json

# Release configs
releases:
    # Releases absolute path
    path: /app/var/releases
    name: "[.Tag]"

# Build configs
build:
    # Build absolute path
    path: /app/var/build
    # Number of parallel builds
    parallelism: 1

# Application Database
database:
    # Database driver (redis)
    driver: redis
    # Redis
    redis:
        hash_prefix: rabbit_

# Third Party API Integration
integrations:
    # Github Configs
    github:
        # Webhook URI (Full URL will be app.domain + webhook_uri)
        webhook_uri: /webhook/github
        # Webhook Secret (From Repo settings page > Webhooks)
        webhook_secret: Pz2ufk7r5BTjnkOo
        # whether to use ssh or https to clone
        clone_with: https
        # HTTPS URL format, Full name will be something like Clivern/Rabbit
        https_format: https://github.com/{$full_name}.git
        # SSH URL format, Full name will be something like Clivern/Rabbit
        ssh_format: git@github.com:{$full_name}.git
    # Bitbucket Configs
    bitbucket:
        # Webhook URI (Full URL will be app.domain + webhook_uri)
        webhook_uri: /webhook/bitbucket
        # whether to use ssh or https to clone
        clone_with: https
        # HTTPS URL format, Full name will be something like Clivern/Rabbit
        https_format: https://bitbucket.org/{$full_name}.git
        # SSH URL format, Full name will be something like Clivern/Rabbit
        ssh_format: git@bitbucket.org:{$full_name}.git
```

And then run the application.

```bash
$ go build rabbit.go
$ ./rabbit

// OR

$ go run rabbit.go

// To Provide a custom config file
$ ./rabbit -config=/custom/path/config.prod.yml
$ go run rabbit.go -config=/custom/path/config.prod.yml
```

## Deployment

Rabbit needs a decent resources to be able to work properly because the build process itself done by goreleaser and it consumes a lot. So keep `build.parallelism` equal `1` and increase if you have more resources and would like to speed the build process.

### On a Linux Server

Make sure you have `git`, `golang 1.12` and `goreleaser` installed, and make goreleaser executable from everywhere.

```bash
# To download the latest goreleaser binary for linux (https://github.com/goreleaser/goreleaser/releases)
$ curl -sL https://github.com/goreleaser/goreleaser/releases/download/v0.109.0/goreleaser_Linux_x86_64.tar.gz | tar xz
```

Also make sure you are able to clone all your repositories in a non-interactive way. Just configure ssh-key and add the remote VCS to your known hosts.

Then download [the latest Rabbit binary.](https://github.com/Clivern/Rabbit/releases)

```bash
$ curl -sL https://github.com/Clivern/Rabbit/releases/download/x.x.x/rabbit_x.x.x_OS.tar.gz | tar xz
```

Create your config file as explained before on development part and run rabbit with systemd or anything else you prefer.

```
$ ./rabbit -config=/custom/path/config.prod.yml
```

### On Docker

Running rabbit with `docker-compose` is pretty straightforward.

```bash
$ git clone https://github.com/Clivern/Rabbit.git
$ cd Rabbit
$ git checkout tags/0.1.2
$ cd deployments/docker-compose
$ docker-compose build
$ docker-compose up -d
```

Docker will mount you host server `~/.ssh` directory in order to be able to clone repositories that need ssh key. Please make sure it has the right permissions and also remote VCS added to known hosts. otherwise rabbit will stuck on git interactive clone.

## Versioning

For transparency into our release cycle and in striving to maintain backward compatibility, Rabbit is maintained under the [Semantic Versioning guidelines](https://semver.org/) and release process is predictable and business-friendly.

See the [Releases section of our GitHub project](https://github.com/clivern/rabbit/releases) for changelogs for each release version of Rabbit. It contains summaries of the most noteworthy changes made in each release.


## Bug tracker

If you have any suggestions, bug reports, or annoyances please report them to our issue tracker at https://github.com/clivern/rabbit/issues


## Security Issues

If you discover a security vulnerability within Rabbit, please send an email to [hello@clivern.com](mailto:hello@clivern.com)


## Contributing

We are an open source, community-driven project so please feel free to join us. see the [contributing guidelines](CONTRIBUTING.md) for more details.


## License

© 2019, Clivern. Released under [MIT License](https://opensource.org/licenses/mit-license.php).

**Rabbit** is authored and maintained by [@Clivern](http://github.com/clivern).
