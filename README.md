# pluginrpc-go

[![Build](https://github.com/bufbuild/pluginrpc-go/actions/workflows/ci.yaml/badge.svg?branch=main)](https://github.com/bufbuild/pluginrpc-go/actions/workflows/ci.yaml)
[![Report Card](https://goreportcard.com/badge/github.com/bufbuild/pluginrpc-go)](https://goreportcard.com/report/github.com/bufbuild/pluginrpc-go)
[![GoDoc](https://pkg.go.dev/badge/github.com/bufbuild/pluginrpc-go.svg)](https://pkg.go.dev/github.com/bufbuild/pluginrpc-go)
[![Slack](https://img.shields.io/badge/slack-buf-%23e01563)](https://buf.build/links/slack)

The Golang library for [PluginRPC](https://github.com/bufbuild/pluginrpc-proto).

The [github.com/bufbuild/pluginrpc-go](https://pkg.go.dev/github.com/bufbuild/pluginrpc-go) library
provides all the primitives necessary to operate with the PluginRPC ecosystem. The
`protoc-gen-pluginrpc-go` plugin generates stubs for Protobuf services to work with PluginRPC. It
makes authoring and consuming plugins based on Protobuf services incredibly simple.

For more on the motivation behind PluginRPC, see the
[github.com/bufbuild/pluginrpc-proto](https://github.com/bufbuild/pluginrpc-proto) documentation.

For a full example, see the [internal/example](internal/example) directory. This contains:

- [proto/buf/pluginrpc/example/v1](internal/example/proto/buf/pluginrpc/example/v1): An Protobuf
  package that contains an example Protobuf service `EchoService`.
- [gen/buf/pluginrpc/example/v1](internal/example/gen/buf/pluginrpc/example/v1): The generated code
  from `protoc-gen-go` and `protoc-gen-pluginrpc-go` for the example Protobuf Package.
- [pluginrpc-example-server](internal/example/cmd/pluginrpc-example-server): An implementation of a
  PluginRPC plugin for `EchoService`.
- [pluginrpc-example-client-request](internal/example/cmd/pluginrpc-example-client-request): A
  simple client that calls the `EchoRequest` RPC via invoking `pluginrpc-example-server`.
- [pluginrpc-example-client-list](internal/example/cmd/pluginrpc-example-client-request): A simple
  client that calls the `EchoList` RPC via invoking `pluginrpc-example-server`.
- [pluginrpc-example-client-error](internal/example/cmd/pluginrpc-example-client-error): A simple
  client that calls the `EchoError` RPC via invoking `pluginrpc-example-server`.

## Usage

Install the `protoc-gen-go` and `protoc-gen-pluginrpc-go` plugins:

```bash
$ go install \
    google.golang.org/protobuf/cmd/protoc-gen-go@latest \
    github.com/bufbuild/pluginrpc-go/cmd/protoc-gen-pluginrpc-go@latest
```

Generate stubs. The easiest way to do so is by using [buf](github.com/bufbuild/buf). See Buf's
[generation tutorial] for more details on setting up generation. You'll likely need a `buf.gen.yaml`
file that looks approximately like the following:

```yaml
version: v2
inputs:
  # Or wherever your .proto files live
  - directory: proto
managed:
  enabled: true
  override:
    - file_option: go_package_prefix
      # Replace github.com/acme/foo with the name of your Golang module
      value: github.com/acme/foo/gen
plugins:
  - local: protoc-gen-go
    out: gen
    opt: paths=source_relative
  - local: protoc-gen-pluginrpc-go
    out: gen
    opt: paths=source_relative
```

Build your plugin. See [pluginrpc-example-server](internal/example/cmd/pluginrpc-example-server) for
a full example. Assuming you intend to expose the `EchoService` as a plugin, your code will look
something like this:

```go
func main() {
	pluginrpc.Main(newServer)
}

func newServer() (pluginrpc.Server, error) {
	spec, err := examplev1pluginrpc.EchoServiceSpecBuilder{
		// Note that EchoList does not have optional args and will default to path being the only arg.
		EchoRequest: []pluginrpc.ProcedureOption{pluginrpc.ProcedureWithArgs("echo", "request")},
		EchoError:   []pluginrpc.ProcedureOption{pluginrpc.ProcedureWithArgs("echo", "error")},
	}.Build()
	if err != nil {
		return nil, err
	}
	serverRegistrar := pluginrpc.NewServerRegistrar()
	echoServiceServer := examplev1pluginrpc.NewEchoServiceServer(pluginrpc.NewHandler(), echoServiceHandler{})
	examplev1pluginrpc.RegisterEchoServiceServer(serverRegistrar, echoServiceServer)
	return pluginrpc.NewServer(spec, serverRegistrar)
}

type echoServiceHandler struct{}

func (echoServiceHandler) EchoRequest(_ context.Context, request *examplev1.EchoRequestRequest) (*examplev1.EchoRequestResponse, error) {
    ...
}

func (echoServiceHandler) EchoList(context.Context, *examplev1.EchoListRequest) (*examplev1.EchoListResponse, error) {
    ...
}

func (echoServiceHandler) EchoError(_ context.Context, request *examplev1.EchoErrorRequest) (*examplev1.EchoErrorResponse, error) {
    ...
}
```

Invoke your plugin. You'll create a client that points to your plugin. See
[pluginrpc-example-client-request](internal/example/cmd/pluginrpc-example-client-request) for a full
example. Invocation will look something like this:

```go
client := pluginrpc.NewClient(pluginrpc.NewExecRunner("pluginrpc-example-server"))
echoServiceClient, err := examplev1pluginrpc.NewEchoServiceClient(client)
if err != nil {
    return err
}
response, err := echoServiceClient.EchoRequest(
    context.Background(),
    &examplev1.EchoRequestRequest{
        ...
    },
)
```

See [pluginrpc_test.go](pluginrpc_test.go) for an example of how to test plugins.

## Status: Alpha

This framework is in active development, and should not be considered stable. We're publishing it
publicly to get early feedback as we approach stability.

## Legal

Offered under the [Apache 2 license](https://github.com/bufbuild/pluginrpc-go/blob/main/LICENSE).
