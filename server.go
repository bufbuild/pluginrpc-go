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
	"slices"
	"strconv"
)

// Server is the server for plugin implementations.
//
// The easiest way to run a server for a plugin is to call ServerMain.
type Server interface {
	// Serve serves the plugin.
	Serve(ctx context.Context, env Env) error

	isServer()
}

// NewServer returns a new Server for a given Spec and ServerRegistrar.
//
// The Spec will be validated against the ServerRegistar to make sure there is a
// 1-1 mapping between Procedures and registered paths.
//
// Once passed to this constructor, the ServerRegistrar can no longer have new
// paths registered to it.
func NewServer(spec Spec, serverRegistrar ServerRegistrar, options ...ServerOption) (Server, error) {
	return newServer(spec, serverRegistrar, options...)
}

// ServerOption is an option for a new Server.
type ServerOption func(*serverOptions)

// ServerWithFlagPrefix adds a prefix to the `--plugin-protocol` and `--plugin-spec` flags.
//
// For example, if the prefix `foo` is given, the flags `--foo-plugin-protocol` and
// `--foo-plugin-spec` will be used instead. Plugin authors can choose to specify such prefixes.
func ServerWithFlagPrefix(flagPrefix string) ServerOption {
	return func(serverOptions *serverOptions) {
		serverOptions.flagPrefix = flagPrefix
	}
}

// *** PRIVATE ***

type server struct {
	spec            Spec
	flagPrefix      string
	pathToServeFunc map[string]func(context.Context, Env) error
}

func newServer(spec Spec, serverRegistrar ServerRegistrar, options ...ServerOption) (*server, error) {
	serverOptions := newServerOptions()
	for _, option := range options {
		option(serverOptions)
	}

	pathToServeFunc, err := serverRegistrar.pathToServeFunc()
	if err != nil {
		return nil, err
	}
	for path := range pathToServeFunc {
		if spec.ProcedureForPath(path) == nil {
			return nil, fmt.Errorf("path %q not contained within spec", path)
		}
	}
	for _, procedure := range spec.Procedures() {
		if _, ok := pathToServeFunc[procedure.Path()]; !ok {
			return nil, fmt.Errorf("path %q not registered", procedure.Path())
		}
	}
	return &server{
		spec:            spec,
		flagPrefix:      serverOptions.flagPrefix,
		pathToServeFunc: pathToServeFunc,
	}, nil
}

func (s *server) Serve(ctx context.Context, env Env) error {
	if len(env.Args) == 1 {
		if env.Args[0] == fullFlag(s.flagPrefix, flagProtocolSuffix) {
			_, err := env.Stdout.Write([]byte(strconv.Itoa(protocolVersion) + "\n"))
			return err
		}
		if env.Args[0] == fullFlag(s.flagPrefix, flagSpecSuffix) {
			data, err := marshalFlag(NewProtoSpec(s.spec))
			if err != nil {
				return err
			}
			_, err = env.Stdout.Write(append(data, []byte("\n")...))
			return err
		}
	}
	for _, procedure := range s.spec.Procedures() {
		if slices.Equal(env.Args, []string{procedure.Path()}) {
			serveFunc := s.pathToServeFunc[procedure.Path()]
			return serveFunc(ctx, env)
		}
		// TODO: Make sure args do not overlap in procedures
		if slices.Equal(env.Args, procedure.Args()) {
			serveFunc := s.pathToServeFunc[procedure.Path()]
			return serveFunc(ctx, env)
		}
	}
	return fmt.Errorf("args not recognized: %v", env.Args)
}

func (*server) isServer() {}

type serverOptions struct {
	flagPrefix string
}

func newServerOptions() *serverOptions {
	return &serverOptions{}
}
