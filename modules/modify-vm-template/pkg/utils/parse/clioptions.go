package parse

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/kubevirt/kubevirt-tekton-tasks/modules/shared/pkg/env"
	"github.com/kubevirt/kubevirt-tekton-tasks/modules/shared/pkg/output"
	"github.com/kubevirt/kubevirt-tekton-tasks/modules/shared/pkg/zerrors"
	"github.com/kubevirt/kubevirt-tekton-tasks/modules/shared/pkg/zutils"
	"go.uber.org/zap/zapcore"
	"k8s.io/apimachinery/pkg/api/resource"
	kubevirtv1 "kubevirt.io/client-go/api/v1"
)

const (
	templateNameOptionName      = "template-name"
	templateNamespaceOptionName = "template-namespace"
	colonSeparator              = ":"
)

type CLIOptions struct {
	TemplateName        string            `arg:"--template-name,env:TEMPLATE_NAME,required" placeholder:"NAME" help:"Name of a template"`
	TemplateNamespace   string            `arg:"--template-namespace,env:TEMPLATE_NAMESPACE" placeholder:"NAMESPACE" help:"Namespace of a template"`
	CPUSockets          string            `arg:"--cpu-sockets,env:CPU_SOCKETS" placeholder:"CPU_SOCKETS" help:"Number of CPU sockets"`
	CPUCores            string            `arg:"--cpu-cores,env:CPU_CORES" placeholder:"CPU_CORES" help:"Number of CPU cores"`
	CPUThreads          string            `arg:"--cpu-threads,env:CPU_THREADS" placeholder:"CPU_THREADS" help:"Number of CPU threads"`
	Memory              string            `arg:"--memory,env:MEMORY" placeholder:"MEMORY" help:"Memory of the vm, format 1M, 1G"`
	TemplateLabels      []string          `arg:"--template-labels" placeholder:"KEY: VALUE KEY: VALUE" help:"Adds labels to template"`
	TemplateAnnotations []string          `arg:"--template-annotations" placeholder:"KEY: VALUE KEY: VALUE" help:"Adds annotations to template"`
	VMLabels            []string          `arg:"--vm-labels" placeholder:"KEY: VALUE KEY: VALUE" help:"Adds labels to VMs"`
	VMAnnotations       []string          `arg:"--vm-annotations" placeholder:"KEY: VALUE KEY: VALUE" help:"Adds annotations to VMs"`
	Output              output.OutputType `arg:"-o" placeholder:"FORMAT" help:"Output format. One of: yaml|json"`
	Disks               []string          `arg:"--disks" placeholder:'{"name": "test", "cdrom": {"bus": "sata"}}' '{"name": "disk2"}' help:"VM disks in json format, replace vm disk if same name, otherwise new disk is appended"`
	Volumes             []string          `arg:"--volumes" placeholder:'{"name": "virtiocontainerdisk", "containerDisk": {"image": "kubevirt/virtio-container-disk"}}' '{"name": "disk2"}' help:"VM volumes in json format, replace vm volume if same name, otherwise new volume is appended"`
	Debug               bool              `arg:"--debug" help:"Sets DEBUG log level"`

	templateLabels      map[string]string
	templateAnnotations map[string]string
	vmLabels            map[string]string
	vmAnnotations       map[string]string
	disks               []kubevirtv1.Disk
	volumes             []kubevirtv1.Volume
}

func (c *CLIOptions) GetDebugLevel() zapcore.Level {
	if c.Debug {
		return zapcore.DebugLevel
	}
	return zapcore.InfoLevel
}

func (c *CLIOptions) GetCPUSockets() uint32 {
	res, _ := strconv.ParseUint(c.CPUSockets, 10, 32)
	return uint32(res)
}

func (c *CLIOptions) GetCPUCores() uint32 {
	res, _ := strconv.ParseUint(c.CPUCores, 10, 32)
	return uint32(res)
}

func (c *CLIOptions) GetCPUThreads() uint32 {
	res, _ := strconv.ParseUint(c.CPUThreads, 10, 32)
	return uint32(res)
}

func (c *CLIOptions) GetDisks() []kubevirtv1.Disk {
	return c.disks
}

func (c *CLIOptions) GetVolumes() []kubevirtv1.Volume {
	return c.volumes
}

func (c *CLIOptions) GetMemory() *resource.Quantity {
	if c.Memory == "" {
		return nil
	}
	q := resource.MustParse(c.Memory)
	return &q
}

func (c *CLIOptions) GetTemplateName() string {
	return c.TemplateName
}

func (c *CLIOptions) GetTemplateNamespace() string {
	return c.TemplateNamespace
}

func (c *CLIOptions) GetTemplateLabels() map[string]string {
	return c.templateLabels
}

func (c *CLIOptions) GetTemplateAnnotations() map[string]string {
	return c.templateAnnotations
}

func (c *CLIOptions) GetVMAnnotations() map[string]string {
	return c.vmAnnotations
}

func (c *CLIOptions) GetVMLabels() map[string]string {
	return c.vmLabels
}

func (c *CLIOptions) Init() error {
	c.trimSpaces()

	if err := c.convertDisks(); err != nil {
		return err
	}

	if err := c.convertVolumes(); err != nil {
		return err
	}

	if err := c.assertValidParams(); err != nil {
		return err
	}

	if err := c.assertValidTypes(); err != nil {
		return err
	}

	c.setDefaultValues()

	c.trimSpacesAnnotations()

	return nil
}

func (c *CLIOptions) trimSpacesAnnotations() {
	for key, value := range c.templateAnnotations {
		newKey := strings.TrimSpace(key)
		newValue := strings.TrimSpace(value)
		delete(c.templateAnnotations, key)
		c.templateAnnotations[newKey] = newValue
	}
	for key, value := range c.vmAnnotations {
		newKey := strings.TrimSpace(key)
		newValue := strings.TrimSpace(value)
		delete(c.vmAnnotations, key)
		c.vmAnnotations[newKey] = newValue
	}
}

func (c *CLIOptions) trimSpaces() {
	for _, strVariablePtr := range []*string{&c.TemplateName, &c.TemplateNamespace} {
		*strVariablePtr = strings.TrimSpace(*strVariablePtr)
	}

	for i, value := range c.TemplateLabels {
		value = strings.ReplaceAll(value, " ", "")
		c.TemplateLabels[i] = value
	}
	for i, value := range c.VMLabels {
		value = strings.ReplaceAll(value, " ", "")
		c.VMLabels[i] = value
	}
}

func (c *CLIOptions) setDefaultValues() error {
	if c.TemplateNamespace == "" {
		activeNamespace, err := env.GetActiveNamespace()
		if err != nil {
			return zerrors.NewMissingRequiredError("can't get active namespace: %v", err.Error())
		}

		c.TemplateNamespace = activeNamespace
	}
	return nil
}

func (c *CLIOptions) convertDisks() error {
	mError := zerrors.NewMultiError()
	for _, diskStr := range c.Disks {
		disk := &kubevirtv1.Disk{}
		err := json.Unmarshal([]byte(diskStr), disk)
		if err != nil {
			mError.AddC("wrong disk definition", err)
		}
		c.disks = append(c.disks, *disk)
	}

	if !mError.IsEmpty() {
		return mError
	}

	return nil
}

func (c *CLIOptions) convertVolumes() error {
	mError := zerrors.NewMultiError()
	for _, volumeStr := range c.Volumes {
		volume := &kubevirtv1.Volume{}
		err := json.Unmarshal([]byte(volumeStr), volume)
		if err != nil {
			mError.AddC("wrong volume definition", err)
		}
		c.volumes = append(c.volumes, *volume)
	}

	if !mError.IsEmpty() {
		return mError
	}

	return nil
}

func checkCorrectInt(value string) error {
	if value == "" {
		return nil
	}
	_, err := strconv.ParseUint(value, 0, 32)
	return err
}

func (c *CLIOptions) assertValidParams() error {
	if c.TemplateName == "" {
		return zerrors.NewMissingRequiredError("%s param has to be specified", templateNameOptionName)
	}

	mError := zerrors.NewMultiError()
	if c.Memory != "" {
		_, err := resource.ParseQuantity(c.Memory)
		if err != nil {
			mError.AddC("wrong memory value", err)
		}
	}

	err := checkCorrectInt(c.CPUCores)
	if err != nil {
		mError.AddC("wrong cpu cores value", err)
	}

	err = checkCorrectInt(c.CPUThreads)
	if err != nil {
		mError.AddC("wrong cpu threads value", err)
	}

	err = checkCorrectInt(c.CPUSockets)
	if err != nil {
		mError.AddC("wrong cpu sockets value", err)
	}

	c.templateLabels, err = zutils.ExtractKeysAndValuesByLastKnownKey(c.TemplateLabels, colonSeparator)
	if err != nil {
		return mError.AddC("wrong template labels", err)
	}

	c.templateAnnotations, err = zutils.ExtractKeysAndValuesByLastKnownKey(c.TemplateAnnotations, colonSeparator)
	if err != nil {
		mError.AddC("wrong template annotations", err)
	}

	c.vmLabels, err = zutils.ExtractKeysAndValuesByLastKnownKey(c.VMLabels, colonSeparator)
	if err != nil {
		mError.AddC("wrong vm labels", err)
	}

	c.vmAnnotations, err = zutils.ExtractKeysAndValuesByLastKnownKey(c.VMAnnotations, colonSeparator)
	if err != nil {
		mError.AddC("wrong vm annotations", err)
	}

	if !mError.IsEmpty() {
		return mError
	}

	return nil
}

func (c *CLIOptions) assertValidTypes() error {
	if !output.IsOutputType(string(c.Output)) {
		return zerrors.NewMissingRequiredError("%v is not a valid output type", c.Output)
	}
	return nil
}
