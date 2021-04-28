# Demeris Backend

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