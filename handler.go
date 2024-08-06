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

package pluginrpc

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/mattn/go-isatty"
)

// Handler handles requests on the server side.
//
// This is used within generated code when registering an implementation of a service.
//
// Currently, Handlers do not have any customization, however this type is exposes
// so that customization can be provided in the future.
type Handler interface {
	Handle(
		ctx context.Context,
		env Env,
		request any,
		handle func(context.Context, any) (any, error),
	) error

	isHandler()
}

// NewHandler returns a new Handler.
func NewHandler(...HandlerOption) Handler {
	return newHandler()
}

// HandlerOption is an option for a new Handler.
type HandlerOption func(*handlerOptions)

// *** PRIVATE ***

type handler struct{}

func newHandler() *handler {
	return &handler{}
}

func (h *handler) Handle(
	ctx context.Context,
	env Env,
	request any,
	handle func(context.Context, any) (any, error),
) (retErr error) {
	defer func() {
		if retErr != nil {
			retErr = h.writeError(env, retErr)
		}
	}()

	data, err := readStdin(env.Stdin)
	if err != nil {
		return err
	}
	if err := unmarshalRequest(data, request); err != nil {
		return err
	}
	response, err := handle(ctx, request)
	if err != nil {
		// TODO: This results in writeError being called, but ignores marshaling
		// the response, so we will never have a non-nil response and non-nil
		// error together, which the protocol says we can have.
		//
		// This just needs some refactoring.
		return err
	}
	data, err = marshalResponse(response, nil)
	if err != nil {
		return err
	}
	// We append a newline so that the server will behave nicely as a CLI.
	if _, err = env.Stdout.Write(append(data, []byte("\n")...)); err != nil {
		return fmt.Errorf("failed to write response to stdout: %w", err)
	}
	return err
}

func (h *handler) writeError(env Env, inputErr error) error {
	if inputErr == nil {
		return nil
	}
	data, err := marshalResponse(nil, inputErr)
	if err != nil {
		return err
	}
	if _, err := env.Stdout.Write(data); err != nil {
		return fmt.Errorf("failed to write error to stdout: %w", err)
	}
	return nil
}

func (*handler) isHandler() {}

// readStdin handles stdin specially to determine if stdin is a *os.File (likely os.Stdin)
// and is itself a terminal. If so, we don't block on io.ReadAll, as we know that there
// is no data in stdin and we can return.
//
// This allows server-side implementations of services to not require i.e.:
//
//	echo '{}' | plugin-server /pkg.Service/Method
//
// Instead allowing to just invoke the following if there is no request data:
//
//	plugin-server /pkg.Service/Method
func readStdin(stdin io.Reader) ([]byte, error) {
	file, ok := stdin.(*os.File)
	if ok {
		if isatty.IsTerminal(file.Fd()) || isatty.IsCygwinTerminal(file.Fd()) {
			// Nothing on stdin
			return nil, nil
		}
	}
	return io.ReadAll(stdin)
}

type handlerOptions struct{}
