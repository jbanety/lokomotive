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
package kubelinstor

const chartValuesTmpl = `
nameOverride: "{{ .NameOverride }}"
controller:
  enabled: {{ .Controller.Enabled }}
  image:
    repository: docker.io/kvaps/linstor-controller
    tag: {{ .Controller.ImageTag }}
    pullPolicy: IfNotPresent
    pullSecrets: []

  replicaCount: {{ .Controller.ReplicaCount }}

  port: {{ .Controller.Port }}
  ssl:
    enabled: {{ .Controller.SSL }}
    port: {{ .Controller.SSLPort }}

  service:
    annotations:
      prometheus.io/path: "/metrics?error_reports=false"
      prometheus.io/port: "3370"
      prometheus.io/scrape: "true"

  {{- if .Controller.NodeSelector }}
  nodeSelector: {{ .Controller.NodeSelectorRaw }}
  {{- end }}

  {{- if .Controller.Tolerations }}
  tolerations: {{ .Controller.TolerationsRaw }}
  {{- end }}
  
  initSettings:
    enabled: false
    # Set plain connector listen to localhost 
    plainConnectorBindAddress: "127.0.0.1"
    # Disable user security (required for setting global options)
    disableUserSecurity: true

  # Database config
  db:
    {{- if .Controller.Db.User }}
    user: "{{ .Controller.Db.User }}"
	{{- end }}
    {{- if .Controller.Db.Password }}
    password: "{{ .Controller.Db.Password }}"
	{{- end }}
    connectionUrl: "{{ .Controller.Db.ConnectionUrl }}"
	{{- if .Controller.Db.CaCertificate }}
	caCertificate: true
	{{- end }}
	{{- if .Controller.Db.ClientCertificate }}
	clientCertificate: true
	{{- end }}

satellite:
  enabled: {{ .Satellite.Enabled }}
  image:
    repository: docker.io/kvaps/linstor-satellite
    tag: {{ .Satellite.ImageTag }}
    pullPolicy: IfNotPresent
    pullSecrets: []

  port: {{ .Satellite.Port }}
  ssl:
    enabled: {{ .Satellite.SSL }}
    port: {{ .Satellite.SSLPort }}

  # Oerwrite drbd.conf and global_common.conf files. This option will enable
  # usage-count=no and udev-always-use-vnr options by default
  overwriteDrbdConf: {{ .Satellite.OverwriteDrbdConf }}

  # How many nodes can simultaneously download new image
  update:
    maxUnavailable: {{ .Satellite.UpdateMaxUnavailable }}

  {{- if .Satellite.NodeSelector }}
  nodeSelector: {{ .Satellite.NodeSelectorRaw }}
  {{- end }}

  {{- if .Satellite.Tolerations }}
  tolerations: {{ .Satellite.TolerationsRaw }}
  {{- end }}

csi:
  enabled: {{ .Csi.Enabled }}
  image:
    linstorCsiPlugin:
      repository: docker.io/kvaps/linstor-csi
      tag: {{ .Csi.ImageTag }}
      pullPolicy: IfNotPresent
    csiProvisioner:
      repository: k8s.gcr.io/sig-storage/csi-provisioner
      tag: v2.0.4
      pullPolicy: IfNotPresent
    csiAttacher:
      repository: k8s.gcr.io/sig-storage/csi-attacher
      tag: v3.0.2
      pullPolicy: IfNotPresent
    csiResizer:
      repository: k8s.gcr.io/sig-storage/csi-resizer
      tag: v1.0.1
      pullPolicy: IfNotPresent
    csiSnapshotter:
      repository: k8s.gcr.io/sig-storage/csi-snapshotter
      tag: v3.0.2
      pullPolicy: IfNotPresent
    csiNodeDriverRegistrar:
      repository: k8s.gcr.io/sig-storage/csi-node-driver-registrar
      tag: v2.0.1
      pullPolicy: IfNotPresent
    csiLivenessProbe:
      repository: k8s.gcr.io/sig-storage/livenessprobe
      tag: v2.1.0
      pullPolicy: IfNotPresent

  controller:
    replicaCount: {{ .Controller.ReplicaCount }}

    csiProvisioner:
      topology: false

    {{- if .Controller.NodeSelector }}
    nodeSelector: {{ .Controller.NodeSelectorRaw }}
    {{- end }}

    {{- if .Controller.Tolerations }}
    tolerations: {{ .Controller.TolerationsRaw }}
    {{- end }}

  node:
	{{- if or .Satellite.NodeSelector .Satellite.Tolerations }}
	{{- if .Satellite.NodeSelector }}
    nodeSelector: {{ .Satellite.NodeSelectorRaw }}
    {{- end }}
    {{- if .Satellite.Tolerations }}
    tolerations: {{ .Satellite.TolerationsRaw }}
    {{- end }}
	{{- else }}
	{}
    {{- end }}

haController:
  enabled: {{ .HaController.Enabled }}
  image:
    repository: docker.io/kvaps/linstor-ha-controller
    tag: {{ .HaController.ImageTag }}
    pullPolicy: IfNotPresent

  replicaCount: {{ .HaController.ReplicaCount }}
  
  {{- if .HaController.NodeSelector }}
  nodeSelector: {{ .HaController.NodeSelectorRaw }}
  {{- end }}

  {{- if .HaController.Tolerations }}
  tolerations: {{ .HaController.TolerationsRaw }}
  {{- end }}

stork:
  enabled: {{ .Stork.Enabled }}
  image:
    repository: docker.io/kvaps/linstor-stork
    tag: {{ .Stork.ImageTag }}
    pullPolicy: IfNotPresent

  replicaCount: {{ .Stork.ReplicaCount }}

  {{- if .Stork.NodeSelector }}
  nodeSelector: {{ .Stork.NodeSelectorRaw }}
  {{- end }}

  {{- if .Stork.Tolerations }}
  tolerations: {{ .Stork.TolerationsRaw }}
  {{- end }}

storkScheduler:
  enabled: {{ .StorkScheduler.Enabled }}
  image:
    repository: k8s.gcr.io/kube-scheduler
    tag: {{ .StorkScheduler.ImageTag }}
    pullPolicy: IfNotPresent

  replicaCount: {{ .StorkScheduler.ReplicaCount }}

  {{- if .StorkScheduler.NodeSelector }}
  nodeSelector: {{ .StorkScheduler.NodeSelectorRaw }}
  {{- end }}

  {{- if .StorkScheduler.Tolerations }}
  tolerations: {{ .StorkScheduler.TolerationsRaw }}
  {{- end }}

`
