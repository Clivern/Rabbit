<p align="center">
    <img alt="Rabbit Logo" src="https://raw.githubusercontent.com/Clivern/Rabbit/master/assets/img/logo.png" height="80" />
    <h3 align="center">Rabbit</h3>
    <p align="center">Private go binaries builder & hosting service integrated with github and bitbucket.</p>
    <p align="center">
        <a href="https://godoc.org/github.com/clivern/rabbit"><img src="https://godoc.org/github.com/clivern/rabbit?status.svg"></a>
        <a href="https://travis-ci.org/Clivern/Rabbit"><img src="https://travis-ci.org/Clivern/Rabbit.svg?branch=master"></a>
        <a href="https://github.com/Clivern/Rabbit/releases"><img src="https://img.shields.io/badge/Version-1.0.0-red.svg"></a>
        <a href="https://goreportcard.com/report/github.com/Clivern/Rabbit"><img src="https://goreportcard.com/badge/github.com/Clivern/Rabbit"></a>
        <a href="https://github.com/Clivern/Rabbit/blob/master/LICENSE"><img src="https://img.shields.io/badge/LICENSE-MIT-orange.svg"></a>
    </p>
</p>

Rabbit is a lightweight service that will build and store your go projects binaries. Once a VCS system (github or bitbucket) notifies rabbit of a new release, it clones the project, builds different binaries and publish them.

<p align="center">
    <img src="https://raw.githubusercontent.com/Clivern/Rabbit/master/assets/img/diagram.png" />
</p>

## Documentation

### Installation:

### Config & Run The Application

Rabbit uses [Go Modules](https://github.com/golang/go/wiki/Modules) to manage dependencies. First Create a prod config file.

```bash
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

# Application Database
database:
    # Database driver (redis)
    driver: redis
    # Redis
    redis:
        hash_prefix: rabbit
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

Or [download a pre-built Rabbit binary](https://github.com/Clivern/Rabbit/releases) for your operating system.

```bash
$ curl -sL https://github.com/Clivern/Rabbit/releases/download/x.x.x/rabbit_x.x.x_OS.tar.gz | tar xz
$ ./rabbit -config=config.prod.yml
```

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
