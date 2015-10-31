# registrar

**Important note: Registrar is EXPERIMENTAL.** It implements the Resource Owner
Password Credentials grant and the OpenID userinfo endpoint, and provides an API
endpoint for user registration. Large portions of the OpenID and OAuth 2.0
specifications are not yet implemented. There a no tests. You've been warned! :)

[![Build Status](https://travis-ci.org/paulrosania/registrar.svg?branch=master)](https://travis-ci.org/paulrosania/registrar)

registrar is an OpenID Connect Provider in pure Go. It is a "pure" API server,
designed to work with a separate frontend. (For example, a single-page
javascript app.) In order to make this work, registrar provides a number of API
endpoints beyond the OpenID and OAuth specs, meant for communication with the
frontend, including:

* User registration
* Permissions grants and rejection
* Authorization revocation

## Installation

    go get -u github.com/paulrosania/registrar

## Quick Start

    # Generate keys for JWT signing
    openssl genrsa -out registrar.rsa <key-size>
    openssl rsa -in registrar.rsa -pubout > registrar.rsa.pub

    # Configure registrar
    cp registrar.ini.sample registrar.ini
    $EDITOR registrar.ini

    # Run the server
    registrar -c registrar.ini

## Documentation

Full API documentation is available here:

[https://godoc.org/github.com/paulrosania/registrar](https://godoc.org/github.com/paulrosania/registrar)

## Contributing

1. Fork the project
2. Make your changes
2. Run tests (`go test`)
3. Send a pull request!

If you’re making a big change, please open an issue first, so we can discuss.

## License

MIT
