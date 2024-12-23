# Nats Demo

Repository I use to test nats.

## Repo structure

### infrastructure/local

Contains files for spinning up a local demo environment

### service

Contains a small go program that can be used as either a subscripber or a publisher

## Starting local demo environment

[Install flox](https://flox.dev/docs/install-flox/) and activate the environment. This will install necessary tools and setup the development environment

`flox activate`

Start the demo environment

`make local-start`