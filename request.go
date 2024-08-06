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

func marshalRequest(request any) ([]byte, error) {
	requestMessage, err := toProtoMessage(request)
	if err != nil {
		return nil, err
	}
	body, err := anypb.New(requestMessage)
	if err != nil {
		return nil, err
	}
	protoRequest := &pluginrpcv1beta1.Request{
		Body: body,
	}
	return protojson.Marshal(protoRequest)
}

func unmarshalRequest(data []byte, request any) error {
	if len(data) == 0 {
		return nil
	}
	protoRequest := &pluginrpcv1beta1.Request{}
	if err := protojson.Unmarshal(data, protoRequest); err != nil {
		return err
	}
	if body := protoRequest.GetBody(); body != nil {
		requestMessage, err := toProtoMessage(request)
		if err != nil {
			return err
		}
		return anypb.UnmarshalTo(body, requestMessage, proto.UnmarshalOptions{})
	}
	return nil
}
