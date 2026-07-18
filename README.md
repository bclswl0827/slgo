# slgo

SeedLink protocol basic implementation in pure Go.

slgo implements a lightweight SeedLink server/provider/consumer stack suitable for testing and simple deployments. It includes example provider and consumer implementations that demonstrate how to publish seismic-like sample data into the SeedLink pipeline and serve it to SeedLink clients (for example the Swarm client).

## Preview

![Swarm Screenshot](https://raw.githubusercontent.com/bclswl0827/slgo/master/preview/swarm.png)

## Features

- **SeedLink server**: minimal SeedLink server implementation.
- **Provider interface**: implement `GetStations()`, `GetStreams()`, `QueryHistory()` to supply station/stream metadata and historical data.
- **Consumer interface**: subscribe/unsubscribe to live data published by your data source.
- **Example**: runnable example in the `example/` directory showing a simple provider, consumer and a message bus generating sample waveforms.

## Installation

Requires Go 1.20 or later.

To fetch the module for use in other projects:

```shell
$ go get github.com/bclswl0827/slgo
```

To run the included example server locally:

```shell
$ go run ./example
```

The example launches a SeedLink server on `0.0.0.0:18000` by default and publishes synthetic data into the message bus. Test it with a SeedLink client such as Swarm (https://volcanoes.usgs.gov/software/swarm/download.shtml).

## Quick Usage

Create a server by providing implementations of the provider, consumer and hooks interfaces and call `Start`:

```go
server := slgo.New(myProvider, myConsumer, myHooks)
err := server.Start(ctx, "0.0.0.0", 18000, true)
```

- `myProvider` should implement methods like `GetStations()`, `GetStreams()`, `QueryHistory(start,end,channels)`.
- `myConsumer` should implement `Subscribe(clientId, channels, handler)` and `Unsubscribe(clientId)`.

See the reference implementations in the `example` directory: [example/provider.go](example/provider.go#L1-L200) and [example/consumer.go](example/consumer.go#L1-L200).

## Project layout

- `example/` — runnable example server, provider and consumer.
- `handlers/` — SeedLink protocol handlers and data types.
- `*.go` — core implementation and tests.

## Contributing

Contributions, issues and feature requests are welcome. Please open an issue or pull request on the upstream repository.

## License

This project is licensed under the MIT License.
