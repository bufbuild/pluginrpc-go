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

package main

import (
	"context"
	"errors"

	"github.com/bufbuild/pluginrpc-go"
	examplev1 "github.com/bufbuild/pluginrpc-go/internal/example/gen/buf/pluginrpc/example/v1"
	"github.com/bufbuild/pluginrpc-go/internal/example/gen/buf/pluginrpc/example/v1/examplev1pluginrpc"
)

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
	return &examplev1.EchoRequestResponse{Message: request.GetMessage()}, nil
}

func (echoServiceHandler) EchoList(context.Context, *examplev1.EchoListRequest) (*examplev1.EchoListResponse, error) {
	return &examplev1.EchoListResponse{List: []string{"foo", "bar"}}, nil
}

func (echoServiceHandler) EchoError(_ context.Context, request *examplev1.EchoErrorRequest) (*examplev1.EchoErrorResponse, error) {
	return nil, pluginrpc.NewError(pluginrpc.Code(request.GetCode()), errors.New(request.GetMessage()))
}
