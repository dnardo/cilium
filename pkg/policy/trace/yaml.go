// Copyright 2017 Authors of Cilium
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

package trace

import (
	"fmt"
	"io"
	"os"

	"github.com/cilium/cilium/pkg/labels"

	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"
)

const (
	// ReplicationController represents a Kubernetes ReplicationController kind.
	ReplicationController = "ReplicationController"

	// Deployment represents a Kubernetes Deployment kind.
	Deployment = "Deployment"

	// ReplicaSet represents a Kubernetes ReplicaSet kind.
	ReplicaSet = "ReplicaSet"

	// DefaultNamespace represents the default Kubernetes namespace.
	DefaultNamespace = "default"
)

// GetLabelsFromYaml iterates through the provided YAML file and for each
// section  in the YAML, returns the labels or an error if the labels could not
// be parsed.
func GetLabelsFromYaml(file string) ([][]string, error) {

	kinds, err := getKindsFromYaml(file)
	if err != nil {
		return nil, err
	}

	reader, err := os.Open(file)

	if err != nil {
		return nil, err
	}
	defer reader.Close()

	splitYamlLabels := [][]string{}

	yamlDecoder := yaml.NewYAMLToJSONDecoder(reader)
	for _, v := range kinds {
		yamlLabels := []string{}
		switch v {
		case Deployment:
			var deployment v1beta1.Deployment
			err = yamlDecoder.Decode(&deployment)
			if err != nil {
				return nil, err
			}
			var ns string
			if deployment.Namespace != "" {
				ns = deployment.Namespace
			} else {
				ns = DefaultNamespace
			}
			yamlLabels = append(yamlLabels, labels.GenerateK8sLabelString(labels.K8sNamespaceLabel, ns))

			for k, v := range deployment.Spec.Template.Labels {
				yamlLabels = append(yamlLabels, labels.GenerateK8sLabelString(k, v))
			}
		case ReplicationController:
			var controller v1.ReplicationController
			err = yamlDecoder.Decode(&controller)
			if err != nil {
				return nil, err
			}
			var ns string
			if controller.Namespace != "" {
				ns = controller.Namespace
			} else {
				ns = DefaultNamespace
			}
			yamlLabels = append(yamlLabels, labels.GenerateK8sLabelString(labels.K8sNamespaceLabel, ns))

			for k, v := range controller.Spec.Template.Labels {
				yamlLabels = append(yamlLabels, labels.GenerateK8sLabelString(k, v))
			}
		case ReplicaSet:
			var rep v1beta1.ReplicaSet
			err = yamlDecoder.Decode(&rep)
			if err != nil {
				return nil, err
			}
			var ns string
			if rep.Namespace != "" {
				ns = rep.Namespace
			} else {
				ns = DefaultNamespace
			}
			yamlLabels = append(yamlLabels, labels.GenerateK8sLabelString(labels.K8sNamespaceLabel, ns))

			for k, v := range rep.Spec.Template.Labels {
				yamlLabels = append(yamlLabels, labels.GenerateK8sLabelString(k, v))
			}
		default:
			return nil, fmt.Errorf("%s, %s, and %s are the only supported types at this time", Deployment, ReplicationController, ReplicaSet)
		}

		splitYamlLabels = append(splitYamlLabels, yamlLabels)
	}
	return splitYamlLabels, nil
}

// getKindsFromYaml iterates through the provided YAML file, and returns a slice
// of the value of the 'Kind' field in the file. If the file cannot be parsed,
// an error is returned.
func getKindsFromYaml(file string) ([]string, error) {
	reader, err := os.Open(file)

	if err != nil {
		return nil, err
	}
	defer reader.Close()

	yamlDecoder := yaml.NewYAMLToJSONDecoder(reader)
	kinds := []string{}
	for err == nil {
		var yamlData interface{}
		// Each successive call to Decode on yamlDecoder decodes the next YAML
		// section into yamlData.
		err = yamlDecoder.Decode(&yamlData)
		if err != nil {
			// We've reached the end of the file.
			if err == io.EOF {
				break
			} else {
				return nil, err
			}
		}
		m := yamlData.(map[string]interface{})
		// Get the value of the 'Kind' field of each YAML file.
		kind, err := parseKind(m)
		if err != nil {
			return nil, err
		}
		kinds = append(kinds, kind)
	}

	return kinds, nil
}

// parseKind iterates through the provided data from a YAMl file and returns
// the value of the 'Kind' field in the provided data. If no such mapping
// exists, or is not a Deployment, ReplicationController, or ReplicaSet, an
// error is returned.
func parseKind(yamlData map[string]interface{}) (string, error) {
	if v, ok := yamlData["kind"]; ok {

		switch v.(type) {
		case string:
		default:
			return "", fmt.Errorf("Improperly formatted YAML provided; %q did not map to a string", "kind")
		}

		vStr := v.(string)

		switch vStr {
		case Deployment, ReplicationController, ReplicaSet:
			return vStr, nil
		default:
			return "", fmt.Errorf("%s, %s, and %s are the only supported types at this time", Deployment, ReplicationController, ReplicaSet)
		}
	} else {
		return "", fmt.Errorf("Improperly formatted YAML provided")
	}
}
