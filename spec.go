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
	"fmt"
	"slices"
	"strings"

	pluginrpcv1beta1 "buf.build/gen/go/bufbuild/pluginrpc/protocolbuffers/go/buf/pluginrpc/v1beta1"
)

// Spec specifies a set of Procedures that a plugin implements. This describes
// the shape of the plugin to clients.
//
// JSON-encoded Specs are returned on stdout when `--plugin-spec` is called, optionally
// with a custom flag prefix. For example if the flag prefix `foo` is specified, the
// flag `--foo-plugin-spec` will returned a JSON-encoded Spec.
//
// A given Spec will have no duplicate Procedures either by path or args.
type Spec interface {
	// ProcedureForPath returns the Procedure for the given path.
	//
	// If no such procedure exists, this returns nil.
	ProcedureForPath(path string) Procedure
	// Procedures returns all Procedures.
	Procedures() []Procedure

	isSpec()
}

// NewSpec returns a new validated Spec for the given Procedures.
func NewSpec(procedures []Procedure) (Spec, error) {
	return newSpec(procedures)
}

// NewSpecForProto returns a new validated Spec for the given pluginrpcv1beta1.Spec.
func NewSpecForProto(protoSpec *pluginrpcv1beta1.Spec) (Spec, error) {
	procedures := make([]Procedure, len(protoSpec.GetProcedures()))
	for i, protoProcedure := range protoSpec.GetProcedures() {
		procedure, err := NewProcedureForProto(protoProcedure)
		if err != nil {
			return nil, err
		}
		procedures[i] = procedure
	}
	return NewSpec(procedures)
}

// NewProtoSpec returns a new pluginrpcv1beta1.Spec for the given Spec.
func NewProtoSpec(spec Spec) *pluginrpcv1beta1.Spec {
	procedures := spec.Procedures()
	protoProcedures := make([]*pluginrpcv1beta1.Procedure, len(procedures))
	for i, procedure := range procedures {
		protoProcedures[i] = NewProtoProcedure(procedure)
	}
	return &pluginrpcv1beta1.Spec{
		Procedures: protoProcedures,
	}
}

// CombineSpecs returns a new validated Spec that represents the combination of the given Specs.
//
// Note that since the returns Spec is valid, this means that Procedures from the given
// Specs must not contain any duplicate Procedures by path or args.
func CombineSpecs(specs ...Spec) (Spec, error) {
	var procedures []Procedure
	for _, spec := range specs {
		procedures = append(procedures, spec.Procedures()...)
	}
	return NewSpec(procedures)
}

// *** PRIVATE ***

type spec struct {
	procedures      []Procedure
	pathToProcedure map[string]Procedure
}

func newSpec(procedures []Procedure) (*spec, error) {
	if err := validateSpecProcedures(procedures); err != nil {
		return nil, err
	}
	pathToProcedure := make(map[string]Procedure)
	for _, procedure := range procedures {
		pathToProcedure[procedure.Path()] = procedure
	}
	return &spec{
		procedures:      procedures,
		pathToProcedure: pathToProcedure,
	}, nil
}

func (s *spec) ProcedureForPath(path string) Procedure {
	return s.pathToProcedure[path]
}

func (s *spec) Procedures() []Procedure {
	return slices.Clone(s.procedures)
}

func (*spec) isSpec() {}

func validateSpecProcedures(procedures []Procedure) error {
	usedPathMap := make(map[string]struct{})
	usedArgsMap := make(map[string]struct{})
	for _, procedure := range procedures {
		path := procedure.Path()
		if _, ok := usedPathMap[path]; ok {
			return fmt.Errorf("duplicate procedure path: %q", path)
		}
		usedPathMap[path] = struct{}{}
		args := procedure.Args()
		if len(args) > 0 {
			// We can do this given that we have a valid Spec where
			// args do not contain spaces.
			joinedArgs := strings.Join(args, " ")
			if _, ok := usedArgsMap[joinedArgs]; ok {
				return fmt.Errorf("duplicate procedure args: %q", joinedArgs)
			}
			usedArgsMap[joinedArgs] = struct{}{}
		}
	}
	return nil
}
