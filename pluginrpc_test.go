// Copyright 2024 Buf Technologies, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package pluginrpc_test

import (
	"context"
	"errors"
	"testing"

	pluginrpcv1beta1 "buf.build/gen/go/bufbuild/pluginrpc/protocolbuffers/go/buf/pluginrpc/v1beta1"
	"github.com/bufbuild/pluginrpc-go"
	examplev1 "github.com/bufbuild/pluginrpc-go/internal/example/gen/buf/pluginrpc/example/v1"
	"github.com/bufbuild/pluginrpc-go/internal/example/gen/buf/pluginrpc/example/v1/examplev1pluginrpc"
	"github.com/stretchr/testify/require"
)

func TestEchoRequest(t *testing.T) {
	t.Parallel()
	server, err := newServer()
	require.NoError(t, err)
	echoServiceClient, err := examplev1pluginrpc.NewEchoServiceClient(newClient(server))
	require.NoError(t, err)
	response, err := echoServiceClient.EchoRequest(
		context.Background(),
		&examplev1.EchoRequestRequest{
			Message: "hello",
		},
	)
	require.NoError(t, err)
	require.NotNil(t, response)
	require.Equal(t, "hello", response.GetMessage())
}

func TestEchoRequestNil(t *testing.T) {
	t.Parallel()
	server, err := newServer()
	require.NoError(t, err)
	echoServiceClient, err := examplev1pluginrpc.NewEchoServiceClient(newClient(server))
	require.NoError(t, err)
	response, err := echoServiceClient.EchoRequest(
		context.Background(),
		nil,
	)
	require.NoError(t, err)
	require.NotNil(t, response)
	require.Equal(t, "", response.GetMessage())
}

func TestEchoRequestFlagPrefix(t *testing.T) {
	t.Parallel()
	server, err := newServer(pluginrpc.ServerWithFlagPrefix("foo"))
	require.NoError(t, err)
	echoServiceClient, err := examplev1pluginrpc.NewEchoServiceClient(
		newClient(
			server,
			pluginrpc.ClientWithFlagPrefix("foo"),
		),
	)
	require.NoError(t, err)
	response, err := echoServiceClient.EchoRequest(
		context.Background(),
		&examplev1.EchoRequestRequest{
			Message: "hello",
		},
	)
	require.NoError(t, err)
	require.NotNil(t, response)
	require.Equal(t, "hello", response.GetMessage())
}

func TestEchoList(t *testing.T) {
	t.Parallel()
	server, err := newServer()
	require.NoError(t, err)
	echoServiceClient, err := examplev1pluginrpc.NewEchoServiceClient(newClient(server))
	require.NoError(t, err)
	response, err := echoServiceClient.EchoList(context.Background(), nil)
	require.NoError(t, err)
	require.NotNil(t, response)
	require.Equal(t, []string{"foo", "bar"}, response.GetList())
}

func TestEchoError(t *testing.T) {
	t.Parallel()
	server, err := newServer()
	require.NoError(t, err)
	echoServiceClient, err := examplev1pluginrpc.NewEchoServiceClient(newClient(server))
	require.NoError(t, err)
	_, err = echoServiceClient.EchoError(
		context.Background(),
		&examplev1.EchoErrorRequest{
			Code:    pluginrpcv1beta1.Code_CODE_DEADLINE_EXCEEDED,
			Message: "hello",
		},
	)
	pluginrpcError := &pluginrpc.Error{}
	require.ErrorAs(t, err, &pluginrpcError)
	require.Equal(t, pluginrpc.CodeDeadlineExceeded, pluginrpcError.Code())
	unwrappedErr := pluginrpcError.Unwrap()
	require.Error(t, unwrappedErr)
	require.Equal(t, "hello", unwrappedErr.Error())
}

func newClient(server pluginrpc.Server, clientOptions ...pluginrpc.ClientOption) pluginrpc.Client {
	return pluginrpc.NewClient(pluginrpc.NewServerRunner(server), clientOptions...)
}

func newServer(serverOptions ...pluginrpc.ServerOption) (pluginrpc.Server, error) {
	spec, err := examplev1pluginrpc.EchoServiceSpecBuilder{
		// Note that EchoList does not have a ProcedureBuilder and will default to path being the only arg.
		EchoRequest: []pluginrpc.ProcedureOption{pluginrpc.ProcedureWithArgs("echo", "request")},
		EchoError:   []pluginrpc.ProcedureOption{pluginrpc.ProcedureWithArgs("echo", "error")},
	}.Build()
	if err != nil {
		return nil, err
	}
	serverRegistrar := pluginrpc.NewServerRegistrar()
	handler := pluginrpc.NewHandler()
	echoServiceHandler := newEchoServiceHandler()
	echoServiceServer := examplev1pluginrpc.NewEchoServiceServer(handler, echoServiceHandler)
	examplev1pluginrpc.RegisterEchoServiceServer(serverRegistrar, echoServiceServer)
	return pluginrpc.NewServer(spec, serverRegistrar, serverOptions...)
}

type echoServiceHandler struct{}

func newEchoServiceHandler() *echoServiceHandler {
	return &echoServiceHandler{}
}

func (*echoServiceHandler) EchoRequest(
	_ context.Context,
	request *examplev1.EchoRequestRequest,
) (*examplev1.EchoRequestResponse, error) {
	return &examplev1.EchoRequestResponse{
		Message: request.GetMessage(),
	}, nil
}

func (*echoServiceHandler) EchoList(
	context.Context,
	*examplev1.EchoListRequest,
) (*examplev1.EchoListResponse, error) {
	return &examplev1.EchoListResponse{
		List: []string{
			"foo",
			"bar",
		},
	}, nil
}

func (*echoServiceHandler) EchoError(
	_ context.Context,
	request *examplev1.EchoErrorRequest,
) (*examplev1.EchoErrorResponse, error) {
	return nil, pluginrpc.NewError(pluginrpc.Code(request.GetCode()), errors.New(request.GetMessage()))
}
