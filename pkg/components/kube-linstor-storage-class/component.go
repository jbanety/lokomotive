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

package kubelinstorstorageclass

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"

	"github.com/kinvolk/lokomotive/pkg/components"
	"github.com/kinvolk/lokomotive/pkg/k8sutil"
)

const (
	// Name represents kube-linstor-storage-class component name as it should be referenced in function calls
	// and in configuration.
	Name = "kube-linstor-storage-class"
	Namespace = "kube-linstor"
)

// StorageClass represents single Linstor storage class with properties.
type StorageClass struct {
	// Name of the storage class
	Name                    string              `hcl:"name,label"`
	// Make the storage class as default
	Default                 bool                `hcl:"default,optional"`
	// Reclaim policy of volume
	ReclaimPolicy		    string              `hcl:"reclaim_policy"`
	// Sets the file system type to create for volumeMode: FileSystem PVCs
	FsType                  string              `hcl:"fs_type,optional"`
	// Determines the amount of replicas a volume of this StorageClass will have
	// If you use this option, you must not use nodeList.
	AutoPlace               int                 `hcl:"auto_place,optional"`
	// The LINSTOR Resource Group (RG) to associate with this StorageClass.
	// If not set, a new RG will be created for each new PVC.
	ResourceGroup           string              `hcl:"resource_group,optional"`
	// Name of the LINSTOR storage pool that will be used to provide storage to the newly-created volumes.
	StoragePool             string              `hcl:"storage_pool,optional"`
	// DisklessStoragePool is an optional parameter that only effects LINSTOR volumes assigned disklessly to
	// kubelets i.e., as clients. If you have a custom diskless storage pool defined in LINSTOR, you’ll specify that here.
	DisklessStoragePool     string              `hcl:"diskless_storage_pool,optional"`
	// A comma-seperated list of layers to use for the created volumes.
	LayerList               string              `hcl:"layer_list,optional"`
	// Name of the scheduler used to place volumes
	PlacementPolicy	        string              `hcl:"placement_policy"`
	// Disable remote access to volumes.
	AllowRemoteVolumeAccess bool                `hcl:"allow_remote_volume_access,optional"`
	// Determines whether to encrypt volumes
	Encryption              bool                `hcl:"encryption,optional"`
	// List of nodes for volumes to be assigned to
	// If you use this option, you must not use autoPlace.
	NodeList				[]string            `hcl:"node_list,optional"`
	NodeListRaw             string
	// List of nodes for diskless volumes to be assigned to.
	// Use in conjunction with nodeList.
	ClientList				[]string            `hcl:"client_list,optional"`
	ClientListRaw           string
	// List of key or key=value items used as autoplacement selection labels
	// when autoplace is used to determine where to provision storage.
	// These labels correspond to LINSTOR node properties.
	ReplicasOnSame			[]string            `hcl:"replicas_on_same,optional"`
	ReplicasOnSameRaw		string
	// List of properties to consider, same as replicasOnSame.
	ReplicasOnDifferent     []string            `hcl:"replicas_on_different,optional"`
	ReplicasOnDifferentRaw	string
	// Create a diskless resource on all nodes that were not assigned a diskful resource.
	DisklessOnRemaining     bool                `hcl:"diskless_on_remaining,optional"`
	// Do not place the resource on a node which has a resource with a name matching the regex.
	DoNotPlaceWithRegex     string              `hcl:"do_not_place_with_regex,optional"`
	// Parameter that passes options to the volume’s filesystem at creation time.
	FsOpts					string              `hcl:"fs_opts,optional"`
	// Parameter that passes options to the volume’s filesystem at mount time.
	MountOpts			    string              `hcl:"mount_opts,optional"`
	// Extra arguments to pass to xfs_io, which gets called before right before first use of the volume.
	PostMountXfsOpts        string              `hcl:"post_mount_xfs_opts,optional"`
	// Advanced DRBD options to pass to LINSTOR.
	// The full list of options is available at https://app.swaggerhub.com/apis-docs/Linstor/Linstor/1.5.0#/developers/resourceDefinitionModify
	DrbdOptions				[]map[string]string `hcl:"drbd_options,optional"`
}

type component struct {
	StorageClasses []StorageClass `hcl:"storage-class,block"`
}

func defaultStorageClass() StorageClass {
	return StorageClass{
		Name: "linstor-retain-3-replicas",
		Default: true,
		FsType: "ext4",
		AutoPlace: 3,
	}
}

// NewConfig returns new OpenEBS Storage Class component configuration with default values set.
//
//nolint:golint
func NewConfig() *component {
	return &component{
		StorageClasses: []StorageClass{},
	}
}

func (c *component) LoadConfig(configBody *hcl.Body, evalContext *hcl.EvalContext) hcl.Diagnostics {
	if configBody == nil {
		return hcl.Diagnostics{}
	}

	if diagnostics := gohcl.DecodeBody(*configBody, evalContext, c); diagnostics != nil {
		return diagnostics
	}

	// if empty config body is provided, default component storage details are still preserved.
	if len(c.StorageClasses) == 0 {
		c.StorageClasses = append(c.StorageClasses, defaultStorageClass())
	}

	if err := c.validateConfig(); err != nil {
		return hcl.Diagnostics{
			&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Validation of the config failed",
				Detail:   fmt.Sprintf("validation failed: %v", err),
			},
		}
	}

	return nil
}

func (c *component) validateConfig() error {
	maxDefaultStorageClass := 0

	for _, sc := range c.StorageClasses {
		if sc.Default == true {
			maxDefaultStorageClass++
		}
		if maxDefaultStorageClass > 1 {
			return fmt.Errorf("cannot have more than one default storage class")
		}
	}

	return nil
}

// TODO: Convert to Helm chart.
func (c *component) RenderManifests() (map[string]string, error) {

	scTmpl, err := template.New(Name).Parse(storageClassTmpl)
	if err != nil {
		return nil, fmt.Errorf("parsing storage class template: %w", err)
	}

	var manifestsMap = make(map[string]string)

	for _, sc := range c.StorageClasses {
		var scBuffer bytes.Buffer

		sc.NodeListRaw = RenderList(sc.NodeList, ",")
		sc.ClientListRaw = RenderList(sc.ClientList, ",")
		sc.ReplicasOnSameRaw = RenderList(sc.ReplicasOnSame, " ")
		sc.ReplicasOnDifferentRaw = RenderList(sc.ReplicasOnDifferent, " ")

		if err := scTmpl.Execute(&scBuffer, sc); err != nil {
			return nil, fmt.Errorf("executing storage class %q template: %w", sc.Name, err)
		}

		filename := fmt.Sprintf("%s-%s.yml", Name, sc.Name)
		manifestsMap[filename] = scBuffer.String()

	}

	return manifestsMap, nil
}

func (c *component) Metadata() components.Metadata {
	return components.Metadata{
		Name: Name,
		// Return the same namespace which the openebs-operator component is using.
		Namespace: k8sutil.Namespace{
			Name: Namespace,
		},
	}
}

func RenderList(s []string, sep string) string {
	if len(s) == 0 {
		return ""
	}

	b := strings.Join(s[:], sep)

	return b
}


