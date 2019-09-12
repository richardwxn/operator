package values

import (
	"github.com/gogo/protobuf/jsonpb"
	"k8s.io/apimachinery/pkg/util/intstr"
)

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
