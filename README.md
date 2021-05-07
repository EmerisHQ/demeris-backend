# Demeris Backend

## Generating API documentation

To generate the OpenAPI specification document (`swagger.yml`), run:

```shell
make generate-swagger
```

## Compiling

Each compilation target resides under a directory living in `cmd`, for example to build `tracelistener` one would execute:

```shell
make tracelistener
```

To build all the project's binaries, run:

```shell
make
```

By default `make` will produce stripped and optimized binaries.

To build a non-stripped binary with debug symbols, append `DEBUG=true` in your environment or when calling `make`:

```shell
make DEBUG=true
```

Build targets are automatically updated as soon as you create a new directory under `cmd`, no need to modify the
`Makefile` to include them.

## Cleaning

To clean the generated OpenAPI specification and build artifacts, run:

```shell
make clean
```

## Docker

To build Docker images for `cmd` binaries, run from the root of this repository:

```shell
docker build -t [yourbinary]:latest -f Dockerfile.<yourbinary> .^
```

## Local Kubernetes environment

### Requirements

* kubectl
* docker (docker desktop will probably install kubectl)
* helm
* kind

### Usage

