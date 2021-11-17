# Bookkeeper Operator

 [![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0) [![GoDoc](https://godoc.org/github.com/pravega/bookkeeper-operator?status.svg)](https://godoc.org/github.com/pravega/bookkeeper-operator) [![Build Status](https://travis-ci.org/pravega/bookkeeper-operator.svg?branch=master)](https://travis-ci.org/pravega/bookkeeper-operator) [![Go Report](https://goreportcard.com/badge/github.com/pravega/bookkeeper-operator)](https://goreportcard.com/report/github.com/pravega/bookkeeper-operator) [![Version](https://img.shields.io/github/release/pravega/bookkeeper-operator.svg)](https://github.com/pravega/bookkeeper-operator/releases)

## Overview

[Bookkeeper](https://bookkeeper.apache.org/) is a scalable, fault-tolerant, and low-latency storage service optimized for real-time workloads.

The Bookkeeper Operator manages Bookkeeper clusters deployed to Kubernetes and automates tasks related to operating a Bookkeeper cluster.The operator itself is built with the [Operator framework](https://github.com/operator-framework/operator-sdk).

## Project status: alpha

The project is currently alpha. While no breaking API changes are currently planned, we reserve the right to address bugs and change the API before the project is declared stable.

## Install the Operator

To understand how to deploy a Bookkeeper Operator refer to [Operator Deployment](https://github.com/pravega/charts/tree/master/charts/bookkeeper-operator#deploying-bookkeeper-operator).

## Upgrade the Operator

For upgrading the bookkeeper operator check the document on [Operator Upgrade](doc/operator-upgrade.md)

## Features

- [x] [Create and destroy a Bookkeeper cluster](https://github.com/pravega/charts/tree/master/charts/bookkeeper#deploying-bookkeeper)
- [x] [Resize cluster](https://github.com/pravega/charts/tree/master/charts/bookkeeper#updating-bookkeeper-cluster)
- [x] [Rolling upgrades/Rollback](doc/upgrade-cluster.md)
- [x] [Bookkeeper Configuration tuning](doc/configuration.md)
- [x] Input validation

## Development

Check out the [development guide](doc/development.md).

## Releases  

The latest Bookkeeper releases can be found on the [Github Release](https://github.com/pravega/bookkeeper-operator/releases) project page.

## Contributing and Community

We thrive to build a welcoming and open community for anyone who wants to use the operator or contribute to it. [Here](https://github.com/pravega/bookkeeper-operator/wiki/Contributing) we describe how to contribute to bookkeepe operator. Contact the developers and community on [slack](https://pravega-io.slack.com/) ([signup](https://pravega-slack-invite.herokuapp.com/)) if you need any help.

## Troubleshooting

Check out the [bookkeeper troubleshooting](doc/troubleshooting.md#bookkeeper-cluster-issues) for bookkeeper issues and for operator issues [operator troubleshooting](doc/troubleshooting.md#bookkeeper-operator-issues).

## License

Bookkeeper Operator is under Apache 2.0 license. See the [LICENSE](https://github.com/pravega/bookkeeper-operator/blob/master/LICENSE) for details.
