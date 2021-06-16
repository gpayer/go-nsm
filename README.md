# go-nsm - NSM client framework

## Overview

This package implements the NSM protocol and enables you to easily add NSM client functionality to your music application.

See http://non.tuxfamily.org/nsm/API.html for further details.

## Installation

Execute

```
go get -u github.com/gpayer/go-nsm
```

in your project.

## Example

See directory `example`. Some session managers expect executables of clients to be in paths reachable by `$PATH`, so the `install.sh` script copies `nsm-example-client` to `/usr/local/bin`.