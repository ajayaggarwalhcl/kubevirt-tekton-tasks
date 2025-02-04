#!/usr/bin/env bash

function visit() {
  pushd "${1}" > /dev/null
}

function leave() {
  popd > /dev/null
}

export IS_OKD="false"
export IS_MINIKUBE="false"

if kubectl get projects > /dev/null 2>&1; then
  export IS_OKD="true"
elif minikube status 2>&1 | grep -q Running; then
  export IS_MINIKUBE="true"
fi

export EXCLUDED_NON_IMAGE_MODULES="shared|sharedtest|tests"
export DEPENDENCY_MODULES="shared|sharedtest"

declare -A IMAGE_MODULE_NAME_TO_ENV_NAME
declare -A TASK_NAME_TO_IMAGE

export CREATE_VM_IMAGE="${CREATE_VM_IMAGE:-}"
IMAGE_MODULE_NAME_TO_ENV_NAME["create-vm"]="CREATE_VM_IMAGE"
TASK_NAME_TO_IMAGE["create-vm-from-manifest"]="${CREATE_VM_IMAGE}"
TASK_NAME_TO_IMAGE["create-vm-from-template"]="${CREATE_VM_IMAGE}"

export EXECUTE_IN_VM_IMAGE="${EXECUTE_IN_VM_IMAGE:-}"
IMAGE_MODULE_NAME_TO_ENV_NAME["execute-in-vm"]="EXECUTE_IN_VM_IMAGE"
TASK_NAME_TO_IMAGE["execute-in-vm"]="${EXECUTE_IN_VM_IMAGE}"
TASK_NAME_TO_IMAGE["cleanup-vm"]="${EXECUTE_IN_VM_IMAGE}"

export DISK_VIRT_CUSTOMIZE_IMAGE="${DISK_VIRT_CUSTOMIZE_IMAGE:-}"
IMAGE_MODULE_NAME_TO_ENV_NAME["disk-virt-customize"]="DISK_VIRT_CUSTOMIZE_IMAGE"
TASK_NAME_TO_IMAGE["disk-virt-customize"]="${DISK_VIRT_CUSTOMIZE_IMAGE}"

export GENERATE_SSH_KEYS_IMAGE="${GENERATE_SSH_KEYS_IMAGE:-}"
IMAGE_MODULE_NAME_TO_ENV_NAME["generate-ssh-keys"]="GENERATE_SSH_KEYS_IMAGE"
TASK_NAME_TO_IMAGE["generate-ssh-keys"]="${GENERATE_SSH_KEYS_IMAGE}"

export DISK_VIRT_SYSPREP_IMAGE="${DISK_VIRT_SYSPREP_IMAGE:-}"
IMAGE_MODULE_NAME_TO_ENV_NAME["disk-virt-sysprep"]="DISK_VIRT_SYSPREP_IMAGE"
TASK_NAME_TO_IMAGE["disk-virt-sysprep"]="${DISK_VIRT_SYSPREP_IMAGE}"

export WAIT_FOR_VMI_STATUS_IMAGE="${WAIT_FOR_VMI_STATUS_IMAGE:-}"
IMAGE_MODULE_NAME_TO_ENV_NAME["wait-for-vmi-status"]="WAIT_FOR_VMI_STATUS_IMAGE"
TASK_NAME_TO_IMAGE["wait-for-vmi-status"]="${WAIT_FOR_VMI_STATUS_IMAGE}"

export COPY_TEMPLATE_IMAGE="${COPY_TEMPLATE_IMAGE:-}"
IMAGE_MODULE_NAME_TO_ENV_NAME["copy-template"]="COPY_TEMPLATE_IMAGE"
TASK_NAME_TO_IMAGE["copy-template"]="${COPY_TEMPLATE_IMAGE}"

export MODIFY_VM_TEMPLATE_IMAGE="${MODIFY_VM_TEMPLATE_IMAGE:-}"
IMAGE_MODULE_NAME_TO_ENV_NAME["modify-vm-template"]="MODIFY_VM_TEMPLATE_IMAGE"
TASK_NAME_TO_IMAGE["modify-vm-template"]="${MODIFY_VM_TEMPLATE_IMAGE}"
