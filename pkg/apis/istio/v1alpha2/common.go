// Copyright 2017 Istio Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v1alpha2

// TODO: create remaining enum types.

import (
	"bufio"
	"fmt"
	"github.com/gogo/protobuf/jsonpb"
	"k8s.io/apimachinery/pkg/util/intstr"
	"os"
)

// GetFileLines reads the text file at filePath and returns it as a slice of strings.
func GetFileLines(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer file.Close()

	var out []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		out = append(out, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return out, nil
}

// define new type from k8s intstr to marshal/unmarshal jsonpb
type IntOrStringForPB struct {
	intstr.IntOrString
}

// MarshalJSONPB implements the jsonpb.JSONPBMarshaler interface.
func (intstrpb *IntOrStringForPB) MarshalJSONPB(_ *jsonpb.Marshaler) ([]byte, error) {
	return intstrpb.MarshalJSON()
}

// UnmarshalJSONPB implements the jsonpb.JSONPBUnmarshaler interface.
func (intstrpb *IntOrStringForPB) UnmarshalJSONPB(_ *jsonpb.Unmarshaler, value []byte) error {
	return intstrpb.UnmarshalJSON(value)
}

// FromInt creates an IntOrStringForPB object with an int32 value.
func FromInt(val int) IntOrStringForPB {
	return IntOrStringForPB{intstr.FromInt(val)}
}

// FromString creates an IntOrStringForPB object with a string value.
func FromString(val string) IntOrStringForPB {
	return IntOrStringForPB{intstr.FromString(val)}
}