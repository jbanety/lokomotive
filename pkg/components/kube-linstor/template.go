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
    repository: {{ .Controller.Image.Repository }}
    tag: {{ .Controller.Image.Tag }}
    pullPolicy: {{ .Controller.Image.PullPolicy }}

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
    tls: {{ .Controller.Db.TLS }}
	{{- if .Controller.Db.CaCertificate }}
    ca: |
{{ .Controller.Db.CaCertificateRaw }}
	{{- end }}
	{{- if .Controller.Db.ClientCertificate }}
    cert: |
{{ .Controller.Db.ClientCertificateRaw }}
	{{- end }}
	{{- if .Controller.Db.ClientKey }}
    key: |
{{ .Controller.Db.ClientKeyRaw }}
	{{- end }}
    {{- if .Controller.Db.EtcdPrefix }}
    etcdPrefix: "{{ .Controller.Db.EtcdPrefix }}"
	{{- end }}

satellite:
  enabled: {{ .Satellite.Enabled }}
  image:
    repository: {{ .Satellite.Image.Repository }}
    tag: {{ .Satellite.Image.Tag }}
    pullPolicy: {{ .Satellite.Image.PullPolicy }}

  port: {{ .Satellite.Port }}
  ssl:
    enabled: {{ .Satellite.SSL }}
    port: {{ .Satellite.SSLPort }}

  # Overwrite drbd.conf and global_common.conf files. This option will enable
  # usage-count=no and udev-always-use-vnr options by default
  overwriteDrbdConf: {{ .Satellite.OverwriteDrbdConf }}

  autoJoinCluster: {{ .Satellite.AutoJoinCluster }}
{{- if or .Satellite.StoragePools.LVMPools .Satellite.StoragePools.LVMThinPools .Satellite.StoragePools.ZFSPools }}
  storagePools:
{{- if .Satellite.StoragePools.LVMPools }}
    lvmPools:
{{- range .Satellite.StoragePools.LVMPools }}
      - name: {{ .Name }}
        volumeGroup: {{ .VolumeGroup }}
{{- if .DevicePaths }}
		devicePaths:
{{- range .DevicePaths }}
		  - {{ . }}
{{- end }}
{{- end }}
{{- if .RaidLevel }}
        raidLevel: {{ .RaidLevel }}
{{- end }}
        vdo: {{ .VDO }}
{{- if .VDO }}
		vdoLogicalSizeKib: {{ .VdoLogicalSizeKib }}
		vdoSlabSizeKib: {{ .VdoSlabSizeKib }}
{{- end }}	
{{- end }}
{{- end }}
{{- if .Satellite.StoragePools.LVMThinPools }}
    lvmThinPools:
{{- range .Satellite.StoragePools.LVMThinPools }}
      - name: {{ .Name }}
        volumeGroup: {{ .VolumeGroup }}
        thinVolume: {{ .ThinVolume }}
{{- if .RaidLevel }}
        raidLevel: {{ .RaidLevel }}
{{- end }}
{{- end }}
{{- end }}
{{- if .Satellite.StoragePools.ZFSPools }}
    zfsPools:
{{- range .Satellite.StoragePools.ZFSPools }}
      - name: {{ .Name }}
        zPool: {{ .ZPool }}
        thin: {{ .Thin }}
{{- end }}
{{- end }}
{{- end }}

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
{{- range $name, $image := .Csi.Images }}
    {{ $name }}:
      repository: {{ $image.Repository }}
      tag: {{ $image.Tag }}
      pullPolicy: {{ $image.PullPolicy }}
{{- end }}

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
    repository: {{ .HaController.Image.Repository }}
    tag: {{ .HaController.Image.Tag }}
    pullPolicy: {{ .HaController.Image.PullPolicy }}

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
    repository: {{ .Stork.Image.Repository }}
    tag: {{ .Stork.Image.Tag }}
    pullPolicy: {{ .Stork.Image.PullPolicy }}

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
    repository: {{ .StorkScheduler.Image.Repository }}
    tag: {{ .StorkScheduler.Image.Tag }}
    pullPolicy: {{ .StorkScheduler.Image.PullPolicy }}

  replicaCount: {{ .StorkScheduler.ReplicaCount }}

  {{- if .StorkScheduler.NodeSelector }}
  nodeSelector: {{ .StorkScheduler.NodeSelectorRaw }}
  {{- end }}

  {{- if .StorkScheduler.Tolerations }}
  tolerations: {{ .StorkScheduler.TolerationsRaw }}
  {{- end }}

`
