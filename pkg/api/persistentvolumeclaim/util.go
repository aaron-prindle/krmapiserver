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

package persistentvolumeclaim

import (
	utilfeature "github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/util/feature"
	"github.com/aaron-prindle/krmapiserver/pkg/apis/core"
	"github.com/aaron-prindle/krmapiserver/pkg/features"
)

const (
	pvc            string = "PersistentVolumeClaim"
	volumeSnapshot string = "VolumeSnapshot"
)

// DropDisabledFields removes disabled fields from the pvc spec.
// This should be called from PrepareForCreate/PrepareForUpdate for all resources containing a pvc spec.
func DropDisabledFields(pvcSpec, oldPVCSpec *core.PersistentVolumeClaimSpec) {
	if !utilfeature.DefaultFeatureGate.Enabled(features.BlockVolume) && !volumeModeInUse(oldPVCSpec) {
		pvcSpec.VolumeMode = nil
	}
	if !dataSourceIsEnabled(pvcSpec) && !dataSourceInUse(oldPVCSpec) {
		pvcSpec.DataSource = nil
	}
}

func volumeModeInUse(oldPVCSpec *core.PersistentVolumeClaimSpec) bool {
	if oldPVCSpec == nil {
		return false
	}
	if oldPVCSpec.VolumeMode != nil {
		return true
	}
	return false
}

func dataSourceInUse(oldPVCSpec *core.PersistentVolumeClaimSpec) bool {
	if oldPVCSpec == nil {
		return false
	}
	if oldPVCSpec.DataSource != nil {
		return true
	}
	return false
}

func dataSourceIsEnabled(pvcSpec *core.PersistentVolumeClaimSpec) bool {
	if pvcSpec.DataSource != nil {
		if pvcSpec.DataSource.Kind == pvc &&
			*pvcSpec.DataSource.APIGroup == "" &&
			utilfeature.DefaultFeatureGate.Enabled(features.VolumePVCDataSource) {
			return true

		}
		if pvcSpec.DataSource.Kind == volumeSnapshot &&
			*pvcSpec.DataSource.APIGroup == "snapshot.storage.k8s.io" &&
			utilfeature.DefaultFeatureGate.Enabled(features.VolumeSnapshotDataSource) {
			return true
		}
	}
	return false
}
