package test

import (
	"github.com/kubevirt/kubevirt-tekton-tasks/modules/sharedtest/testobjects"
	. "github.com/kubevirt/kubevirt-tekton-tasks/modules/tests/test/constants"
	"github.com/kubevirt/kubevirt-tekton-tasks/modules/tests/test/framework"
	"github.com/kubevirt/kubevirt-tekton-tasks/modules/tests/test/runner"
	"github.com/kubevirt/kubevirt-tekton-tasks/modules/tests/test/testconfigs"
	"github.com/kubevirt/kubevirt-tekton-tasks/modules/tests/test/vm"
	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	kubevirtv1 "kubevirt.io/client-go/api/v1"
)

var _ = Describe("Create VM from manifest", func() {
	f := framework.NewFramework().
		OnBeforeTestSetup(func(config framework.TestConfig) {
			if createVMConfig, ok := config.(*testconfigs.CreateVMTestConfig); ok {
				createVMConfig.TaskData.CreateMode = CreateVMVMManifestMode
			}
		})

	table.DescribeTable("taskrun fails and no VM is created", func(config *testconfigs.CreateVMTestConfig) {
		f.TestSetup(config)

		expectedVM := config.TaskData.GetExpectedVMStubMeta()
		f.ManageVMs(expectedVM) // in case it succeeds

		runner.NewTaskRunRunner(f, config.GetTaskRun()).
			CreateTaskRun().
			ExpectFailure().
			ExpectLogs(config.GetAllExpectedLogs()...).
			ExpectResults(nil)

		_, err := vm.WaitForVM(f.KubevirtClient, f.CdiClient, expectedVM.Namespace, expectedVM.Name,
			"", config.GetTaskRunTimeout(), false)
		Expect(err).Should(HaveOccurred())
	},
		table.Entry("no vm manifest", &testconfigs.CreateVMTestConfig{
			TaskRunTestConfig: testconfigs.TaskRunTestConfig{
				ExpectedLogs: "one of vm-manifest, template-name should be specified",
			},
			TaskData: testconfigs.CreateVMTaskData{},
		}),
		table.Entry("no service account", &testconfigs.CreateVMTestConfig{
			TaskRunTestConfig: testconfigs.TaskRunTestConfig{
				ExpectedLogs: "cannot create resource \"virtualmachines\" in API group \"kubevirt.io\"",
			},
			TaskData: testconfigs.CreateVMTaskData{
				VM: testobjects.NewTestAlpineVM("no-sa").Build(),
			},
		}),
		table.Entry("invalid manifest", &testconfigs.CreateVMTestConfig{
			TaskRunTestConfig: testconfigs.TaskRunTestConfig{
				ServiceAccount: CreateVMFromManifestServiceAccountName,
				ExpectedLogs:   "could not read VM manifest: error unmarshaling",
			},
			TaskData: testconfigs.CreateVMTaskData{
				VMManifest: "invalid manifest",
			},
		}),
		table.Entry("non existent dv", &testconfigs.CreateVMTestConfig{
			TaskRunTestConfig: testconfigs.TaskRunTestConfig{
				ServiceAccount: CreateVMFromManifestServiceAccountName,
				ExpectedLogs:   "datavolumes.cdi.kubevirt.io \"non-existent-dv\" not found",
			},
			TaskData: testconfigs.CreateVMTaskData{
				VM:          testobjects.NewTestAlpineVM("vm-with-non-existent-dv").Build(),
				DataVolumes: []string{"non-existent-dv"},
			},
		}),
		table.Entry("non existent owned dv", &testconfigs.CreateVMTestConfig{
			TaskRunTestConfig: testconfigs.TaskRunTestConfig{
				ServiceAccount: CreateVMFromManifestServiceAccountName,
				ExpectedLogs:   "datavolumes.cdi.kubevirt.io \"non-existent-own-dv\" not found",
			},
			TaskData: testconfigs.CreateVMTaskData{
				VM:             testobjects.NewTestAlpineVM("vm-with-non-existent-owned-dv").Build(),
				OwnDataVolumes: []string{"non-existent-own-dv"},
			},
		}),
		table.Entry("non existent pvc", &testconfigs.CreateVMTestConfig{
			TaskRunTestConfig: testconfigs.TaskRunTestConfig{
				ServiceAccount: CreateVMFromManifestServiceAccountName,
				ExpectedLogs:   "persistentvolumeclaims \"non-existent-pvc\" not found",
			},
			TaskData: testconfigs.CreateVMTaskData{
				VM:                     testobjects.NewTestAlpineVM("vm-with-non-existent-pvc").Build(),
				PersistentVolumeClaims: []string{"non-existent-pvc"},
			},
		}),
		table.Entry("non existent owned pvcs", &testconfigs.CreateVMTestConfig{
			TaskRunTestConfig: testconfigs.TaskRunTestConfig{
				ServiceAccount: CreateVMFromManifestServiceAccountName,
				ExpectedLogs:   "persistentvolumeclaims \"non-existent-own-pvc\" not found\npersistentvolumeclaims \"non-existent-own-pvc-2\" not found",
			},
			TaskData: testconfigs.CreateVMTaskData{
				VM:                        testobjects.NewTestAlpineVM("vm-with-non-existent-owned-pvcs").Build(),
				OwnPersistentVolumeClaims: []string{"non-existent-own-pvc", "non-existent-own-pvc-2"},
			},
		}),
		table.Entry("create vm with non matching disk fails", &testconfigs.CreateVMTestConfig{
			TaskRunTestConfig: testconfigs.TaskRunTestConfig{
				ServiceAccount: CreateVMFromManifestServiceAccountName,
				ExpectedLogs:   "admission webhook \"virtualmachine-validator.kubevirt.io\" denied the request: spec.template.spec.domain.devices.disks[0].Name",
			},
			TaskData: testconfigs.CreateVMTaskData{
				VM: testobjects.NewTestAlpineVM("vm-with-non-existent-pvc").WithNonMatchingDisk().Build(),
			},
		}),
		table.Entry("[NAMESPACE SCOPED] cannot create a VM in different namespace", &testconfigs.CreateVMTestConfig{
			TaskRunTestConfig: testconfigs.TaskRunTestConfig{
				ServiceAccount: CreateVMFromManifestServiceAccountName,
				ExpectedLogs:   "cannot create resource \"virtualmachines\" in API group \"kubevirt.io\"",
				LimitTestScope: NamespaceTestScope,
			},
			TaskData: testconfigs.CreateVMTaskData{
				VM:                testobjects.NewTestAlpineVM("different-ns-namespace-scope").Build(),
				VMTargetNamespace: SystemTargetNS,
			},
		}),
		table.Entry("[NAMESPACE SCOPED] cannot create a VM in different namespace in manifest", &testconfigs.CreateVMTestConfig{
			TaskRunTestConfig: testconfigs.TaskRunTestConfig{
				ServiceAccount: CreateVMFromManifestServiceAccountName,
				ExpectedLogs:   "cannot create resource \"virtualmachines\" in API group \"kubevirt.io\"",
				LimitTestScope: NamespaceTestScope,
			},
			TaskData: testconfigs.CreateVMTaskData{
				VM:                        testobjects.NewTestAlpineVM("different-ns-namespace-scope-in-manifest").Build(),
				VMManifestTargetNamespace: SystemTargetNS,
			},
		}),
	)

	table.DescribeTable("VM is created successfully", func(config *testconfigs.CreateVMTestConfig) {
		f.TestSetup(config)

		expectedVM := config.TaskData.GetExpectedVMStubMeta()
		f.ManageVMs(expectedVM)

		runner.NewTaskRunRunner(f, config.GetTaskRun()).
			CreateTaskRun().
			ExpectSuccess().
			ExpectLogs(config.GetAllExpectedLogs()...).
			ExpectResults(map[string]string{
				CreateVMResults.Name:      expectedVM.Name,
				CreateVMResults.Namespace: expectedVM.Namespace,
			})

		_, err := vm.WaitForVM(f.KubevirtClient, f.CdiClient, expectedVM.Namespace, expectedVM.Name,
			"", config.GetTaskRunTimeout(), false)
		Expect(err).ShouldNot(HaveOccurred())
	},
		table.Entry("simple vm", &testconfigs.CreateVMTestConfig{
			TaskRunTestConfig: testconfigs.TaskRunTestConfig{
				ServiceAccount: CreateVMFromManifestServiceAccountName,
				ExpectedLogs:   ExpectedSuccessfulVMCreation,
			},
			TaskData: testconfigs.CreateVMTaskData{
				VM: testobjects.NewTestAlpineVM("simple-vm").Build(),
			},
		}),
		table.Entry("vm to deploy namespace by default", &testconfigs.CreateVMTestConfig{
			TaskRunTestConfig: testconfigs.TaskRunTestConfig{
				ServiceAccount: CreateVMFromManifestServiceAccountName,
				ExpectedLogs:   ExpectedSuccessfulVMCreation,
			},
			TaskData: testconfigs.CreateVMTaskData{
				VM:                                 testobjects.NewTestAlpineVM("vm-to-deploy-by-default").Build(),
				VMTargetNamespace:                  DeployTargetNS,
				UseDefaultVMNamespacesInTaskParams: true,
			},
		}),
		table.Entry("vm with manifest namespace", &testconfigs.CreateVMTestConfig{
			TaskRunTestConfig: testconfigs.TaskRunTestConfig{
				ServiceAccount: CreateVMFromManifestServiceAccountName,
				ExpectedLogs:   ExpectedSuccessfulVMCreation,
			},
			TaskData: testconfigs.CreateVMTaskData{
				VM:                                 testobjects.NewTestAlpineVM("vm-with-manifest-namespace").Build(),
				VMManifestTargetNamespace:          DeployTargetNS,
				UseDefaultVMNamespacesInTaskParams: true,
			},
		}),

		table.Entry("vm with overridden manifest namespace", &testconfigs.CreateVMTestConfig{
			TaskRunTestConfig: testconfigs.TaskRunTestConfig{
				ServiceAccount: CreateVMFromManifestServiceAccountName,
				ExpectedLogs:   ExpectedSuccessfulVMCreation,
			},
			TaskData: testconfigs.CreateVMTaskData{
				VM:                        testobjects.NewTestAlpineVM("vm-with-overridden-manifest-namespace").Build(),
				VMManifestTargetNamespace: DeployTargetNS,
			},
		}),
	)

	It("VM is created from manifest properly ", func() {
		config := &testconfigs.CreateVMTestConfig{
			TaskRunTestConfig: testconfigs.TaskRunTestConfig{
				ServiceAccount: CreateVMFromManifestServiceAccountName,
				ExpectedLogs:   ExpectedSuccessfulVMCreation,
			},
			TaskData: testconfigs.CreateVMTaskData{
				VM: testobjects.NewTestAlpineVM("vm-from-manifest-data").
					WithLabel("app", "my-custom-app").
					WithVMILabel("name", "test").
					WithVMILabel("ra", "rara").
					Build(),
			},
		}
		f.TestSetup(config)

		expectedVMStub := config.TaskData.GetExpectedVMStubMeta()
		f.ManageVMs(expectedVMStub)

		runner.NewTaskRunRunner(f, config.GetTaskRun()).
			CreateTaskRun().
			ExpectSuccess().
			ExpectLogs(config.GetAllExpectedLogs()...).
			ExpectResults(map[string]string{
				CreateVMResults.Name:      expectedVMStub.Name,
				CreateVMResults.Namespace: expectedVMStub.Namespace,
			})

		vm, err := vm.WaitForVM(f.KubevirtClient, f.CdiClient, expectedVMStub.Namespace, expectedVMStub.Name,
			"", config.GetTaskRunTimeout(), false)
		Expect(err).ShouldNot(HaveOccurred())

		vmName := expectedVMStub.Name
		expectedVM := config.TaskData.VM
		// fill VM accordingly
		expectedVM.Spec.Template.Spec.Domain.Machine = vm.Spec.Template.Spec.Domain.Machine // ignore Machine

		Expect(vm.Spec.Template.Spec).Should(Equal(expectedVM.Spec.Template.Spec))
		// check VM labels
		Expect(vm.Labels).Should(Equal(expectedVM.Labels))
		// check VMI labels
		Expect(vm.Spec.Template.ObjectMeta.Labels).Should(Equal(map[string]string{
			"name":                "test",
			"ra":                  "rara",
			"vm.kubevirt.io/name": vmName,
		}))
	})

	Context("with StartVM", func() {
		table.DescribeTable("VM is created successfully", func(config *testconfigs.CreateVMTestConfig, phase kubevirtv1.VirtualMachineInstancePhase, running bool) {
			f.TestSetup(config)

			expectedVMStub := config.TaskData.GetExpectedVMStubMeta()
			f.ManageVMs(expectedVMStub)

			runner.NewTaskRunRunner(f, config.GetTaskRun()).
				CreateTaskRun().
				ExpectSuccess().
				ExpectLogs(config.GetAllExpectedLogs()...).
				ExpectResults(map[string]string{
					CreateVMResults.Name:      expectedVMStub.Name,
					CreateVMResults.Namespace: expectedVMStub.Namespace,
				})

			vm, err := vm.WaitForVM(f.KubevirtClient, f.CdiClient, expectedVMStub.Namespace, expectedVMStub.Name,
				phase, config.GetTaskRunTimeout(), false)
			Expect(err).ShouldNot(HaveOccurred())

			Expect(*vm.Spec.Running).To(Equal(running), "vm should be in correct running phase")
		},
			table.Entry("with false StartVM value", &testconfigs.CreateVMTestConfig{
				TaskRunTestConfig: testconfigs.TaskRunTestConfig{
					ServiceAccount: CreateVMFromManifestServiceAccountName,
					ExpectedLogs:   ExpectedSuccessfulVMCreation,
				},
				TaskData: testconfigs.CreateVMTaskData{
					VM: testobjects.NewTestAlpineVM("vm-from-manifest-data").
						WithLabel("app", "my-custom-app").
						WithVMILabel("name", "test").
						WithVMILabel("ra", "rara").
						Build(),
					StartVM: "false",
				},
			}, kubevirtv1.VirtualMachineInstancePhase(""), false),
			table.Entry("with invalid StartVM value", &testconfigs.CreateVMTestConfig{
				TaskRunTestConfig: testconfigs.TaskRunTestConfig{
					ServiceAccount: CreateVMFromManifestServiceAccountName,
					ExpectedLogs:   ExpectedSuccessfulVMCreation,
				},
				TaskData: testconfigs.CreateVMTaskData{
					VM: testobjects.NewTestAlpineVM("vm-from-manifest-data").
						WithLabel("app", "my-custom-app").
						WithVMILabel("name", "test").
						WithVMILabel("ra", "rara").
						Build(),
					StartVM: "invalid_value",
				},
			}, kubevirtv1.VirtualMachineInstancePhase(""), false),
			table.Entry("with true StartVM value", &testconfigs.CreateVMTestConfig{
				TaskRunTestConfig: testconfigs.TaskRunTestConfig{
					ServiceAccount: CreateVMFromManifestServiceAccountName,
					ExpectedLogs:   ExpectedSuccessfulVMCreation,
				},
				TaskData: testconfigs.CreateVMTaskData{
					VM: testobjects.NewTestAlpineVM("vm-from-manifest-data").
						WithLabel("app", "my-custom-app").
						WithVMILabel("name", "test").
						WithVMILabel("ra", "rara").
						Build(),
					StartVM: "true",
				},
			}, kubevirtv1.Running, true),
		)
	})
})
