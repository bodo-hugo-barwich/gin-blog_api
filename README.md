
# NAME

Gin Blog API

# DESCRIPTION

This a REST API that manages _User_ and _Article_ entities.

It also features a JWT admin authentication.

# REQUIREMENTS

To rebuild this web site the **Minimum Go Compiler Version** is _Go_ `1.20`.\
The site uses the libraries `Gin`, `Gorm` and `json-rust`.\
The _Gin_ Web Server uses the _Gorm_ framework for the database access.\
At the moment only _PostgreSQL_ is supported as database backend.\
The Server Responses are provided as `JSON` documents.

# INSTALLATION

- go

The `go` Command will install the dependencies on local user level as they
are found in the `go.mod` file.

# EXECUTION

- `go run .`

The Site can be launched using the `go run` Command.
To launch the Site call the `go run` Command within the project directory:

            go run .

# IMPLEMENTATION

- API-First Design

To be modular and extendable the **API-First** was chosen.

So, this API is meant to be combined with a web site which will give a grafical interface to the information stored in the API.

