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
	"os"
	"strconv"
	"strings"

	pluginrpcv1beta1 "buf.build/gen/go/bufbuild/pluginrpc/protocolbuffers/go/buf/pluginrpc/v1beta1"
	"github.com/bufbuild/pluginrpc-go"
	examplev1 "github.com/bufbuild/pluginrpc-go/internal/example/gen/buf/pluginrpc/example/v1"
	"github.com/bufbuild/pluginrpc-go/internal/example/gen/buf/pluginrpc/example/v1/examplev1pluginrpc"
)

func main() {
	if err := run(); err != nil {
		if errString := err.Error(); errString != "" {
			_, _ = os.Stderr.Write([]byte(errString + "\n"))
		}
		os.Exit(pluginrpc.WrapExitError(err).ExitCode())
	}
}

func run() error {
	client := pluginrpc.NewClient(pluginrpc.NewExecRunner("pluginrpc-example-server"))
	echoServiceClient, err := examplev1pluginrpc.NewEchoServiceClient(client)
	if err != nil {
		return err
	}
	code, err := strconv.ParseInt(os.Args[1], 10, 32)
	if err != nil {
		return err
	}
	_, err = echoServiceClient.EchoError(
		context.Background(),
		&examplev1.EchoErrorRequest{
			Code:    pluginrpcv1beta1.Code(code),
			Message: strings.Join(os.Args[2:], " "),
		},
	)
	return err
}
