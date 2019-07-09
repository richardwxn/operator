##Convert values_types.go to proto and generate documentation based on original comments

To build, ensure you have the protocol compiler (protoc) installed.

1.Add corresponding comments to values_types.go file

2.Copy the original values_types.go to new file and
  - replace map[string]interface{} occurences with map[string]string
  - replace uint8, uint16 with uint32
  - comment out type defined in k8s corev1. Ex. corev1.ServiceType

3.Setup the K8S [go-to-protobuf tool](https://github.com/kubernetes/code-generator/tree/master/cmd/go-to-protobuf)
   
  Ex:
   
   ```bash
   go build main.go
   ./main apimachinery-packages=istio.io/operator/pkg/apis/istio/v1alpha2/gotoproto [list of paths to the go struct to convert]
   ```
   
   all supported flags can be found here: [flag](https://github.com/kubernetes/code-generator/blob/master/cmd/go-to-protobuf/protobuf/cmd.go#L89)
   
4.generated.proto would be placed in the same path of original go files. The proto file is based on proto2 syntax, can optionally convert it to proto3

5.Use [protoc-gen-docs](https://github.com/istio/tools/tree/master/protoc-gen-docs) to generate documentation from generated.proto, follow the instruction from the link.

  Ex:
  ```bash
  protoc --plugin=./istio.io/tools/protoc-gen-docs/protoc-gen-docs --docs_out=warnings=false,emit_yaml=true,mode=html_page:./page ./istio.io/operator/pkg/apis/istio/v1alpha2/gotoproto/generated.proto
  ```
  The genereated html file would be placed under folder ./page here
  
TODO: Add the above steps to script