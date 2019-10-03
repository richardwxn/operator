package helmreconciler

import (
	"io"
	"strings"

	"istio.io/operator/pkg/name"
)

func InsertChildrenRecursive(componentName name.ComponentName, tree ComponentTree, children ComponentNameToListMap) {
	tree[componentName] = make(ComponentTree)
	for _, child := range children[componentName] {
		InsertChildrenRecursive(child, tree[componentName].(ComponentTree), children)
	}
}

func InstallTreeString(ct ComponentTree) string {
	var sb strings.Builder
	buildInstallTreeString(ct, name.IstioBaseComponentName, "", &sb)
	return sb.String()
}

func buildInstallTreeString(ct ComponentTree, componentName name.ComponentName, prefix string, sb io.StringWriter) {
	_, _ = sb.WriteString(prefix + string(componentName) + "\n")
	if _, ok := ct[componentName].(ComponentTree); !ok {
		return
	}
	for k := range ct[componentName].(ComponentTree) {
		buildInstallTreeString(ct, k, prefix+"  ", sb)
	}
}
