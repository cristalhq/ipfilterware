# ipfilterware

[![build-img]][build-url]
[![pkg-img]][pkg-url]
[![reportcard-img]][reportcard-url]
[![coverage-img]][coverage-url]

Go HTTP middleware to filter clients by IP

## Rationale

To protect your application open to the internet you might want to allow only verified or well-known IPs. This can be easily done via firewall but sometimes you do not have access to such tools (cloud providers, proxies, serverless, etc). To make this real you can check a connection IP and check it with your config. This library does this.

## Features

* Simple API.
* Clean and tested code.
* Thread-safe updates.
* Dependency-free.
* Fetches for popular providers.

## Install

Go version 1.17+

```
go get github.com/cristaloleg/ipfilterware
```

## Example

```go
TODO
```

## Documentation

See [these docs][pkg-url].

## License

[MIT License](LICENSE).

[build-img]: https://github.com/cristaloleg/ipfilterware/workflows/build/badge.svg
[build-url]: https://github.com/cristaloleg/ipfilterware/actions
[pkg-img]: https://pkg.go.dev/badge/cristaloleg/ipfilterware
[pkg-url]: https://pkg.go.dev/github.com/cristaloleg/ipfilterware
[reportcard-img]: https://goreportcard.com/badge/cristaloleg/ipfilterware
[reportcard-url]: https://goreportcard.com/report/cristaloleg/ipfilterware
[coverage-img]: https://codecov.io/gh/cristaloleg/ipfilterware/branch/master/graph/badge.svg
[coverage-url]: https://codecov.io/gh/cristaloleg/ipfilterware