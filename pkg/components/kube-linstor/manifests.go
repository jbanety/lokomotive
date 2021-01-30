// Copyright 2020 The Lokomotive Authors
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

package kubelinstor

const DbSecrets = `
apiVersion: v1
kind: Secret
metadata:
  name: {{ .NameOverride }}-db-tls
  namespace: {{ .Namespace }}
  annotations:
    "helm.sh/resource-policy": "keep"
    "helm.sh/hook": "pre-install"
    "helm.sh/hook-delete-policy": "before-hook-creation"
    "directives.qbec.io/update-policy": "never"
type: kubernetes.io/tls
data:
{{- if .Controller.Db.ClientCertificateRaw }}
  tls.crt: {{ .Controller.Db.ClientCertificateRaw }}
{{- end }}
{{- if .Controller.Db.CaCertificate }}
  ca.crt: {{ .Controller.Db.CaCertificateRaw }}
{{- end }}
`

const PodSecurityPolicy = `
apiVersion: policy/v1beta1
kind: PodSecurityPolicy
metadata:
  name: {{ .NameOverride }}
  annotations:
    seccomp.security.alpha.kubernetes.io/allowedProfileNames: '*'
spec:
  allowPrivilegeEscalation: true
  allowedCapabilities:
    - '*'
  fsGroup:
    rule: RunAsAny
  hostIPC: true
  hostNetwork: true
  hostPID: true
  hostPorts:
    - max: 65535
      min: 0
  privileged: true
  runAsUser:
    rule: RunAsAny
  seLinux:
    rule: RunAsAny
  supplementalGroups:
    rule: RunAsAny
  volumes:
    - '*'
`

const ControllerRBACRole = `
kind: Role
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: {{ .NameOverride }}-controller-psp
rules:
  - apiGroups: ["extensions"]
    resources: ["podsecuritypolicies"]
    resourceNames: ["{{ .NameOverride }}"]
    verbs: ["use"]
`