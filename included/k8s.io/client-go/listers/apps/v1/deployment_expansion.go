/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1

import (
	"fmt"

	apps "github.com/aaron-prindle/krmapiserver/included/k8s.io/api/apps/v1"
	metav1 "github.com/aaron-prindle/krmapiserver/included/k8s.io/apimachinery/pkg/apis/meta/v1"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apimachinery/pkg/labels"
)

// DeploymentListerExpansion allows custom methods to be added to
// DeploymentLister.
type DeploymentListerExpansion interface {
	GetDeploymentsForReplicaSet(rs *apps.ReplicaSet) ([]*apps.Deployment, error)
}

// DeploymentNamespaceListerExpansion allows custom methods to be added to
// DeploymentNamespaceLister.
type DeploymentNamespaceListerExpansion interface{}

// GetDeploymentsForReplicaSet returns a list of Deployments that potentially
// match a ReplicaSet. Only the one specified in the ReplicaSet's ControllerRef
// will actually manage it.
// Returns an error only if no matching Deployments are found.
func (s *deploymentLister) GetDeploymentsForReplicaSet(rs *apps.ReplicaSet) ([]*apps.Deployment, error) {
	if len(rs.Labels) == 0 {
		return nil, fmt.Errorf("no deployments found for ReplicaSet %v because it has no labels", rs.Name)
	}

	// TODO: MODIFY THIS METHOD so that it checks for the podTemplateSpecHash label
	dList, err := s.Deployments(rs.Namespace).List(labels.Everything())
	if err != nil {
		return nil, err
	}

	var deployments []*apps.Deployment
	for _, d := range dList {
		selector, err := metav1.LabelSelectorAsSelector(d.Spec.Selector)
		if err != nil {
			return nil, fmt.Errorf("invalid label selector: %v", err)
		}
		// If a deployment with a nil or empty selector creeps in, it should match nothing, not everything.
		if selector.Empty() || !selector.Matches(labels.Set(rs.Labels)) {
			continue
		}
		deployments = append(deployments, d)
	}

	if len(deployments) == 0 {
		return nil, fmt.Errorf("could not find deployments set for ReplicaSet %s in namespace %s with labels: %v", rs.Name, rs.Namespace, rs.Labels)
	}

	return deployments, nil
}
