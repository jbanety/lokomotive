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

// Package dex has code related to deployment of dex component.
package kubelinstorstorageclass

const storageClassTmpl = `
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: {{ .Name }}
  annotations:
    storageclass.kubernetes.io/is-default-class: "{{ .Default }}"
provisioner: linstor.csi.linbit.com
parameters:
  # CSI related parameters
  {{- if .FsType }}
  csi.storage.k8s.io/fstype: {{ .FsType }}
  {{- end }}
  # LINSTOR parameters
  {{- if .AutoPlace }}
  autoPlace: "{{ .AutoPlace }}"
  {{- end }}
  {{- if .ResourceGroup }}
  resourceGroup: "{{ .ResourceGroup }}"
  {{- end }}
  {{- if .StoragePool }}
  storagePool: "{{ .StoragePool }}"
  {{- end }}	
  {{- if .DisklessStoragePool }}
  disklessStoragePool: "{{ .DisklessStoragePool }}"
  {{- end }}
  {{- if .LayerList }}
  layerList: "{{ .LayerList }}"
  {{- end }}
  placementPolicy: "{{ .PlacementPolicy }}"
  allowRemoteVolumeAccess: "{{ .AllowRemoteVolumeAccess }}"
  encryption: "{{ .Encryption }}"
  {{- if .NodeList }}
  nodeList: "{{ .NodeListRaw }}"
  {{- end }}
  {{- if .ClientList }}
  clientList: "{{ .ClientListRaw }}"
  {{- end }}
  {{- if .ReplicasOnSame }}
  replicasOnSame: "{{ .ReplicasOnSameRaw }}"
  {{- end }}
  {{- if .ReplicasOnDifferent }}
  replicasOnDifferent: "{{ .ReplicasOnDifferentRaw }}"
  {{- end }}
  disklessOnRemaining: "{{ .DisklessOnRemaining }}"
  {{- if .DoNotPlaceWithRegex }}
  doNotPlaceWithRegex: "{{ .DoNotPlaceWithRegex }}"
  {{- end }}
  {{- if .FsOpts }}
  fsOpts: "{{ .FsOpts }}"
  {{- end }}
  {{- if .MountOpts }}
  mountOpts: "{{ .MountOpts }}"
  {{- end }}
  {{- if .PostMountXfsOpts }}
  postMountXfsOpts: "{{ .PostMountXfsOpts }}"
  {{- end }}
  # DRBD parameters
  {{- if .DrbdOptions }}
  {{- range $k, $v := .DrbdOptions }}
  DrbdOptions/{{ $k }}: "{{ $v }}"
  {{- end }}
  {{- end}}
reclaimPolicy: {{ .ReclaimPolicy }}
`
