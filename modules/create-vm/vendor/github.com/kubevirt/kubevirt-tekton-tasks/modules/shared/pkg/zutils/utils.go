package zutils

import (
	"strings"

	"github.com/kubevirt/kubevirt-tekton-tasks/modules/shared/pkg/zconstants"
	"github.com/kubevirt/kubevirt-tekton-tasks/modules/shared/pkg/zerrors"
	templatev1 "github.com/openshift/api/template/v1"
	"k8s.io/apimachinery/pkg/runtime"
	kubevirtv1 "kubevirt.io/client-go/api/v1"
)

func IsTrue(value string) bool {
	return strings.ToLower(value) == zconstants.True
}

func DecodeVM(template *templatev1.Template) (*kubevirtv1.VirtualMachine, int, error) {
	var vm *kubevirtv1.VirtualMachine
	vmIndex := -1
	for i, obj := range template.Objects {
		decoder := kubevirtv1.Codecs.UniversalDecoder(kubevirtv1.GroupVersion)
		decoded, err := runtime.Decode(decoder, obj.Raw)
		if err != nil {
			return nil, vmIndex, err
		}
		done, ok := decoded.(*kubevirtv1.VirtualMachine)
		if ok {
			vm = done
			vmIndex = i
			break
		}
	}
	if vm == nil {
		return nil, vmIndex, zerrors.NewMissingRequiredError("no VM object found in the template")
	}
	return vm, vmIndex, nil
}
