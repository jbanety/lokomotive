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

import (
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/kinvolk/lokomotive/internal"
	"github.com/kinvolk/lokomotive/internal/template"
	"github.com/kinvolk/lokomotive/pkg/components"
	"github.com/kinvolk/lokomotive/pkg/components/util"
	"github.com/kinvolk/lokomotive/pkg/k8sutil"
)

const (
	// Name represents kube-linstor component name as it should be referenced in function calls
	// and in configuration.
	Name = "kube-linstor"
    Indentation = 6
)

type component struct {
	Namespace      string          `hcl:"namespace,optional"`
	NameOverride   string          `hcl:"name_override,optional"`
	Controller     *Controller     `hcl:"controller,block"`
	Satellite      *Satellite      `hcl:"satellite,block"`
	Csi            *Csi            `hcl:"csi,block"`
	HaController   *HaController   `hcl:"ha_controller,block"`
	Stork          *Stork          `hcl:"stork,block"`
	StorkScheduler *StorkScheduler `hcl:"stork_scheduler,block"`
}

// Linstor-controller is main control point for Linstor, it provides API for
// clients and communicates with satellites for creating and monitor DRBD-devices
type Controller struct {
	Enabled         bool 			  `hcl:"enabled,optional"`
	ReplicaCount    int  			  `hcl:"replica_count,optional"`
	Port            int  			  `hcl:"port,optional"`
	SSL			    bool 		      `hcl:"ssl,optional"`
	SSLPort		    int   			  `hcl:"ssl_port,optional"`
	NodeSelector    util.NodeSelector `hcl:"node_selector,optional"`
	NodeSelectorRaw string
	Tolerations     []util.Toleration `hcl:"toleration,block"`
	TolerationsRaw  string
	Db				*Db				  `hcl:"db,block"`
	ImageTag		string			  `hcl:"image_tag,optional"`
}

// Linstor-satellites run on every node, they listen and perform controller tasks
// They operates directly with LVM and ZFS subsystems
type Satellite struct {
	Enabled              bool 			  	  `hcl:"enabled,optional"`
	Port                 int  			  	  `hcl:"port,optional"`
	SSL			         bool 		     	  `hcl:"ssl,optional"`
	SSLPort		         int   				  `hcl:"ssl_port,optional"`
	OverwriteDrbdConf    bool                 `hcl:"overwrite_drbd_conf,optional"`
	UpdateMaxUnavailable int				  `hcl:"update_max_unavailable,optional"`
	NodeSelector         util.NodeSelector    `hcl:"node_selector,optional"`
	NodeSelectorRaw      string
	Tolerations          []util.Toleration    `hcl:"toleration,block"`
	TolerationsRaw       string
	ImageTag		     string			   	  `hcl:"image_tag,optional"`
}

// Linstor CSI driver provides compatibility level for adding Linstor support
// for Kubernetes
type Csi struct {
	Enabled         bool 			  `hcl:"enabled,optional"`
	NodeSelector    util.NodeSelector `hcl:"node_selector,optional"`
	NodeSelectorRaw string
	Tolerations     []util.Toleration `hcl:"toleration,block"`
	TolerationsRaw  string
	ImageTag		string			  `hcl:"image_tag,optional"`
}

// High Availability Controller will speed up the fail over process for stateful
// workloads using Linstor for storage
type HaController struct {
	Enabled         bool 			  `hcl:"enabled,optional"`
	ReplicaCount    int  			  `hcl:"replica_count,optional"`
	NodeSelector    util.NodeSelector `hcl:"node_selector,optional"`
	NodeSelectorRaw string
	Tolerations     []util.Toleration `hcl:"toleration,block"`
	TolerationsRaw  string
	ImageTag		string			  `hcl:"image_tag,optional"`
}

// Stork is a scheduler extender plugin for Kubernetes which allows a storage
// driver to give the Kubernetes scheduler hints about where to place a new pod
// so that it is optimally located for storage performance
type Stork struct {
	Enabled         bool 			  `hcl:"enabled,optional"`
	ReplicaCount    int  			  `hcl:"replica_count,optional"`
	NodeSelector    util.NodeSelector `hcl:"node_selector,optional"`
	NodeSelectorRaw string
	Tolerations     []util.Toleration `hcl:"toleration,block"`
	TolerationsRaw  string
	ImageTag		string			  `hcl:"image_tag,optional"`
}

type StorkScheduler struct {
	Enabled         bool 			  `hcl:"enabled,optional"`
	ReplicaCount    int  			  `hcl:"replica_count,optional"`
	NodeSelector    util.NodeSelector `hcl:"node_selector,optional"`
	NodeSelectorRaw string
	Tolerations     []util.Toleration `hcl:"toleration,block"`
	TolerationsRaw  string
	ImageTag		string			  `hcl:"image_tag,optional"`
}

type Db struct {
	User                 string `hcl:"user,optional"`
	Password             string `hcl:"password,optional"`
	ConnectionUrl        string `hcl:"connection_url"`
	TLS                  bool   `hcl:"tls,optional"`
	CaCertificate        string `hcl:"ca_certificate,optional"`
	CaCertificateRaw     string
	ClientCertificate    string `hcl:"client_certificate,optional"`
	ClientCertificateRaw string
	ClientKey            string `hcl:"client_key,optional"`
	ClientKeyRaw         string
	EtcdPrefix			 string `hcl:"etcd_prefix,optional"`
}

// NewConfig returns new cert-manager component configuration with default values set.
//
//nolint:golint
func NewConfig() *component {
	return &component{
		Namespace: "linstor",
		NameOverride: "linstor",
		Controller: &Controller{
			Enabled: true,
			ImageTag: "v1.11.0",
			ReplicaCount: 2,
			Port: 3370,
			SSL: true,
			SSLPort: 3371,
			Db: &Db{
				TLS: false,
			},
		},
		Satellite: &Satellite{
			Enabled: true,
			ImageTag: "v1.11.0",
			Port: 3366,
			SSL: true,
			SSLPort: 3367,
			OverwriteDrbdConf: true,
			UpdateMaxUnavailable: 40,
		},
		Csi: &Csi{
			Enabled: true,
			ImageTag: "v1.11.0",
		},
		HaController: &HaController{
			Enabled: true,
			ImageTag: "v1.11.0",
			ReplicaCount: 2,
		},
		Stork: &Stork{
			Enabled: true,
			ImageTag: "v1.11.0",
			ReplicaCount: 2,
		},
		StorkScheduler: &StorkScheduler{
			Enabled: true,
			ImageTag: "v1.20.1",
			ReplicaCount: 2,
		},
	}
}

func (c *component) LoadConfig(configBody *hcl.Body, evalContext *hcl.EvalContext) hcl.Diagnostics {
	if configBody == nil {
		return hcl.Diagnostics{
			components.HCLDiagConfigBodyNil,
		}
	}
	return gohcl.DecodeBody(*configBody, evalContext, c)
}

func (c *component) RenderManifests() (map[string]string, error) {
	helmChart, err := components.Chart(Name)
	if err != nil {
		return nil, fmt.Errorf("retrieving chart from assets: %w", err)
	}

	c.Controller.TolerationsRaw, err = util.RenderTolerations(c.Controller.Tolerations)
	if err != nil {
		return nil, fmt.Errorf("rendering tolerations failed: %w", err)
	}

	c.Controller.NodeSelectorRaw, err = c.Controller.NodeSelector.Render()
	if err != nil {
		return nil, fmt.Errorf("rendering node selector failed: %w", err)
	}

	c.Satellite.TolerationsRaw, err = util.RenderTolerations(c.Satellite.Tolerations)
	if err != nil {
		return nil, fmt.Errorf("rendering tolerations failed: %w", err)
	}

	c.Satellite.NodeSelectorRaw, err = c.Satellite.NodeSelector.Render()
	if err != nil {
		return nil, fmt.Errorf("rendering node selector failed: %w", err)
	}

	c.Csi.TolerationsRaw, err = util.RenderTolerations(c.Csi.Tolerations)
	if err != nil {
		return nil, fmt.Errorf("rendering tolerations failed: %w", err)
	}

	c.Csi.NodeSelectorRaw, err = c.Csi.NodeSelector.Render()
	if err != nil {
		return nil, fmt.Errorf("rendering node selector failed: %w", err)
	}

	c.HaController.TolerationsRaw, err = util.RenderTolerations(c.HaController.Tolerations)
	if err != nil {
		return nil, fmt.Errorf("rendering tolerations failed: %w", err)
	}

	c.HaController.NodeSelectorRaw, err = c.HaController.NodeSelector.Render()
	if err != nil {
		return nil, fmt.Errorf("rendering node selector failed: %w", err)
	}

	c.Stork.TolerationsRaw, err = util.RenderTolerations(c.Stork.Tolerations)
	if err != nil {
		return nil, fmt.Errorf("rendering tolerations failed: %w", err)
	}

	c.Stork.NodeSelectorRaw, err = c.Stork.NodeSelector.Render()
	if err != nil {
		return nil, fmt.Errorf("rendering node selector failed: %w", err)
	}

	c.StorkScheduler.TolerationsRaw, err = util.RenderTolerations(c.StorkScheduler.Tolerations)
	if err != nil {
		return nil, fmt.Errorf("rendering tolerations failed: %w", err)
	}

	c.StorkScheduler.NodeSelectorRaw, err = c.StorkScheduler.NodeSelector.Render()
	if err != nil {
		return nil, fmt.Errorf("rendering node selector failed: %w", err)
	}

	if len(c.Controller.Db.CaCertificate) > 0 {
		c.Controller.Db.CaCertificateRaw = internal.Indent(c.Controller.Db.CaCertificate, Indentation)
	}

	if len(c.Controller.Db.ClientCertificate) > 0 {
		c.Controller.Db.ClientCertificateRaw = internal.Indent(c.Controller.Db.ClientCertificate, Indentation)
	}

	if len(c.Controller.Db.ClientKey) > 0 {
		ConvertedClientKey, err := ensurePKCS8Key(c.Controller.Db.ClientKey)
		if err != nil {
			return nil, fmt.Errorf("converting private key: %w", err)
		}
		c.Controller.Db.ClientKeyRaw = internal.Indent(ConvertedClientKey, Indentation)
	}

	values, err := template.Render(chartValuesTmpl, c)
	if err != nil {
		return nil, fmt.Errorf("rendering chart values template: %w", err)
	}

	renderedFiles, err := util.RenderChart(helmChart, Name, c.Namespace, values)
	if err != nil {
		return nil, fmt.Errorf("rendering chart: %w", err)
	}

	podSecurityPolicy, err := template.Render(PodSecurityPolicy, c)
	if err != nil {
		return nil, fmt.Errorf("rendering pod security policy manifest: %w", err)
	}
	renderedFiles["pod-security-policy.yaml"] = podSecurityPolicy

	if c.Controller.Enabled {
		controllerRBACRole, err := template.Render(ControllerRBACRole, c)
		if err != nil {
			return nil, fmt.Errorf("rendering controller rbac role manifest: %w", err)
		}
		renderedFiles["controller-rbac-role.yaml"] = controllerRBACRole
		controllerRBACRoleBinding, err := template.Render(ControllerRBACRoleBinding, c)
		if err != nil {
			return nil, fmt.Errorf("rendering controller rbac role binding manifest: %w", err)
		}
		renderedFiles["controller-rbac-role-binding.yaml"] = controllerRBACRoleBinding
		csiControllerRBACRoleBinding, err := template.Render(CsiControllerRBACRoleBinding, c)
		if err != nil {
			return nil, fmt.Errorf("rendering csi controller rbac role binding manifest: %w", err)
		}
		renderedFiles["csi-controller-rbac-role-binding.yaml"] = csiControllerRBACRoleBinding
		haControllerRBACRoleBinding, err := template.Render(HaControllerRBACRoleBinding, c)
		if err != nil {
			return nil, fmt.Errorf("rendering ha controller rbac role binding manifest: %w", err)
		}
		renderedFiles["ha-controller-rbac-role-binding.yaml"] = haControllerRBACRoleBinding
		storkSchedulerRBACRoleBinding, err := template.Render(StorkSchedulerRBACRoleBinding, c)
		if err != nil {
			return nil, fmt.Errorf("rendering stork scheduler rbac role binding manifest: %w", err)
		}
		renderedFiles["stork-scheduler-rbac-role-binding.yaml"] = storkSchedulerRBACRoleBinding
	}

	return renderedFiles, nil
}

func (c *component) Metadata() components.Metadata {
	return components.Metadata{
		Name: Name,
		Namespace: k8sutil.Namespace{
			Name: c.Namespace,
		},
	}
}

func ensurePKCS8Key(der string) (string, error) {

	privPem, _ := pem.Decode([]byte(der))

	if key, err := x509.ParsePKCS8PrivateKey(privPem.Bytes); err == nil {
		switch key.(type) {
		case *rsa.PrivateKey, *ecdsa.PrivateKey:
			return der, nil
		default:
			return "", errors.New("crypto/tls: found unknown private key type in PKCS#8 wrapping")
		}
	}

	if key, err := x509.ParsePKCS1PrivateKey(privPem.Bytes); err == nil {
		pkcs8Key, err := x509.MarshalPKCS8PrivateKey(key)
		if err != nil {
			return "", fmt.Errorf("marshaling pkcs1 to pkcs8: %w", err)
		}
		return string(pem.EncodeToMemory(
			&pem.Block{
				Type:  "PRIVATE KEY",
				Bytes: pkcs8Key,
			},
		)), nil
	}

	return "", errors.New("unknown private key type")
}