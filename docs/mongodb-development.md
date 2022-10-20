# MongoDB Development

This is just a quick walkthrough on how to setup a local development environment with MongoDB. Useful for building features or testing PRs.

## Tools

Docker
[MongoDB Compass](https://www.mongodb.com/try/download/compass)

## Getting Started

### Starting the mongodb container

The `Makefile` has a quick way to spin up mongodb in docker using `make docker-run-mongo`

You can change the default auth (username:password) credentials which are exported in the `Makefile` if you wish.

The `make` command calls the docker-compose file located at `storage/mongodb/docker-compose.yml`. It uses (currently) mongodb version 4.4 because its compatible with apple silicon devices.

### Connecting to the container

This document wont cover using Compass to connect. The visualization is handy but not required.

Start depserver/syncer

```
./depserver-darwin-amd64 -api supersecret -storage=mongodb -storage-dsn=mongodb://root:root@127.0.0.1:27017

./depsyncer-darwin-amd64 -storage=mongodb -storage-dsn=mongodb://root:root@127.0.0.1:27017 nanomdmdev
```