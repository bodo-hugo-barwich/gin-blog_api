[![Testing against Compiler Versions](https://github.com/bodo-hugo-barwich/gin-blog_api/actions/workflows/compiler-versions.yml/badge.svg)](https://github.com/bodo-hugo-barwich/gin-blog_api/actions/workflows/compiler-versions.yml)

# NAME

Gin Blog API

# DESCRIPTION

This a **REST API** that manages _User_ and _Article_ entities.

It also features a **JWT** admin authentication.

This API is meant to be consumed by a frontend to give it a graphical interface.

# REQUIREMENTS

To rebuild this web site the tested **Minimum Go Compiler Version** is _Go_ `1.19`.\
The site uses the libraries `Gin`, `Gorm` and `golang-jwt`.\
The _Gin_ Web Server uses the _Gorm_ framework for the database access.\
At the moment only _PostgreSQL_ is supported as database backend.\
The server responses are provided as `JSON` documents.

# INSTALLATION

- **go**

The `go` command will install the dependencies on local user level as they
are found in the `go.mod` file.

# CONFIGURATION

- `.env`

The a `.env` file contains the basic configuration for the service.\
A fallback system looks first for the `.env` file corresponding to the `GIN_MODE` like
`.env.test` or `.env.debug` and then falls back to the default `.env` file
if the dedicated file not exists.\
The `.env_sample` can be copied and configured to build a configuration file.


# EXECUTION

- `go run .`

The Site can be launched using the `go run` command.
To launch the Site call the `go run` command within the project directory:

            go run .

# IMPLEMENTATION

- **API-First Design**

To be modular and extendable the **API-First** design was chosen.

So, this API is meant to be combined with a web site which will give a grafical interface to the information stored in the API.

