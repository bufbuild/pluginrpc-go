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
	pluginrpcv1beta1 "buf.build/gen/go/bufbuild/pluginrpc/protocolbuffers/go/buf/pluginrpc/v1beta1"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

func marshalResponse(response any, err error) ([]byte, error) {
	var body *anypb.Any
	if response != nil {
		responseMessage, err := toProtoMessage(response)
		if err != nil {
			return nil, err
		}
		body, err = anypb.New(responseMessage)
		if err != nil {
			return nil, err
		}
	}
	protoResponse := &pluginrpcv1beta1.Response{
		Body:  body,
		Error: WrapError(err).ToProto(),
	}
	return protojson.Marshal(protoResponse)
}

func unmarshalResponse(data []byte, response any) error {
	if len(data) == 0 {
		return nil
	}
	protoResponse := &pluginrpcv1beta1.Response{}
	if err := protojson.Unmarshal(data, protoResponse); err != nil {
		return err
	}
	if body := protoResponse.GetBody(); body != nil {
		responseMessage, err := toProtoMessage(response)
		if err != nil {
			return err
		}
		if err := anypb.UnmarshalTo(body, responseMessage, proto.UnmarshalOptions{}); err != nil {
			return err
		}
	}
	if protoError := protoResponse.GetError(); protoError != nil {
		return NewErrorForProto(protoError)
	}
	return nil
}
