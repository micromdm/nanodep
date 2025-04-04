# NanoDEP

[![CI/CD](https://github.com/micromdm/nanodep/actions/workflows/on-push-pr.yml/badge.svg)](https://github.com/micromdm/nanodep/actions/workflows/on-push-pr.yml) [![Go Reference](https://pkg.go.dev/badge/github.com/micromdm/nanodep.svg)](https://pkg.go.dev/github.com/micromdm/nanodep)

NanoDEP is a set of tools and a Go library powering them for communicating with Apple's Device Enrollment Program (DEP) API servers.

## Getting started & Documentation

- [Quickstart](docs/quickstart.md)  
A guide to get NanoDEP up and running quickly.

- [Operations Guide](docs/operations-guide.md)  
A brief overview of the various tools and utilities for working with NanoDEP.

## Getting the latest version

* Release `.zip` files containing the project should be attached to every [GitHub release](https://github.com/micromdm/nanodep/releases).
  * Release zips are also [published](https://github.com/micromdm/nanodep/actions) for every `main` branch commit.
* A Docker container is built and [published to the GHCR.io registry](http://ghcr.io/micromdm/nanodep) for every release.
  * `docker pull ghcr.io/micromdm/nanodep:latest` — `docker run ghcr.io/micromdm/nanodep:latest`
  * A Docker container is also published for every `main` branch commit (and tagged with `:main`)
* If you have a [Go toolchain installed](https://go.dev/doc/install) you can checkout the source and simply run `make`.

## Tools and utilities

NanoDEP contains a few tools and utilities. At a high level:

- **DEP configuration & reverse proxy server.** The primary server component, called `depserver` is used for configuring NanoDEP and talking with Apple's DEP servers. It hosts its own API for configuring MDM server instances used with Apple's servers (called DEP names) and also hosts a transparently authenticating reverse proxy for talking 'directly' to Apple's DEP API endpoints.
- **Device sync & assigner.** The `depsyncer` tool handles the device fetch/sync cursor logic to continually retrieve the assigned devices from one or more Apple DEP MDM server instance(s).
- **Scripts, tools, and helpers.**
  - A set of [tools](tools) and utilities for talking to the Apple DEP API services — mostly implemented as shell scripts that communicate to the `depserver`.
  - A stand-alone `deptokens` tool for locally working with certificate generation for DEP token decryption.

See the [Operations Guide](docs/operations-guide.md) for more details and usage documentation.

## Go library

NanoDEP is also a Go library for accessing the Apple DEP APIs. There are two components to the Go library:

* The higher-level [godep](https://pkg.go.dev/github.com/micromdm/nanodep/godep) package implements Go methods and structures for talking to the individual DEP API endpoints.
* The lower-level [client](https://pkg.go.dev/github.com/micromdm/nanodep/client) package implements primitives, helpers, and middleware for authenticating to the DEP API and managing sessions tokens.

See the [Go Reference documentation](https://pkg.go.dev/github.com/micromdm/nanodep) (or the Go source itself, of course) for details on these packages.
