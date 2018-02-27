// Copyright 2018 Twitch Interactive, Inc.  All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may not
// use this file except in compliance with the License. A copy of the License is
// located at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// or in the "license" file accompanying this file. This file is distributed on
// an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
// express or implied. See the License for the specific language governing
// permissions and limitations under the License.

package protogen

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	plugin "github.com/golang/protobuf/protoc-gen-go/plugin"
)

// ProtocPlugin describes the interface for protoc-based code generators.
type ProtocPlugin interface {
	Generate(in *plugin.CodeGeneratorRequest) (*plugin.CodeGeneratorResponse, error)
}

// RunProtocPlugin reads a protobuf generator request from standard in, runs the
// supplied generator, and writes its output to stdout. If an error occurs, the
// error will be printed to stderr, and the program will immediately exit with
// status code 1, so you should only call this at the end of your main function.
// This is the way protoc plugins interact with the protoc command.
func RunProtocPlugin(g ProtocPlugin) {
	req, err := readGenRequest(os.Stdin)
	if err != nil {
		exitErr(err)
	}
	resp, err := g.Generate(req)
	if err != nil {
		exitErr(err)
	}
	err = writeResponse(os.Stdout, resp)
	if err != nil {
		exitErr(err)
	}
}

func exitErr(err error) {
	fmt.Fprintln(os.Stderr, "fatal error: "+err.Error())
	os.Exit(1)
}

// FilesToGenerate is a helper to retrieve the set of FileDescriptorProtos
// targeted for generation, as opposed to the ones only included via imports.
func FilesToGenerate(req *plugin.CodeGeneratorRequest) ([]*descriptor.FileDescriptorProto, error) {
	genFiles := make([]*descriptor.FileDescriptorProto, 0)
Outer:
	for _, name := range req.FileToGenerate {
		for _, f := range req.ProtoFile {
			if f.GetName() == name {
				genFiles = append(genFiles, f)
				continue Outer
			}
		}
		return nil, fmt.Errorf("could not find file named %s", name)
	}

	return genFiles, nil
}

func readGenRequest(r io.Reader) (*plugin.CodeGeneratorRequest, error) {
	data, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return nil, err
	}

	req := new(plugin.CodeGeneratorRequest)
	if err = proto.Unmarshal(data, req); err != nil {
		return nil, err
	}

	if len(req.FileToGenerate) == 0 {
		return nil, errors.New("no files to generate")
	}

	return req, nil
}

func writeResponse(w io.Writer, resp *plugin.CodeGeneratorResponse) error {
	data, err := proto.Marshal(resp)
	if err != nil {
		return err
	}
	_, err = w.Write(data)
	if err != nil {
		return err
	}
	return nil
}
