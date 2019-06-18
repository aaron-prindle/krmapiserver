/*
Copyright 2014 The Kubernetes Authors.

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

package options

// This file exists to force the desired plugin implementations to be linked.
// This should probably be part of some configuration fed into the build for a
// given binary target.
import (
	// Admission policies
	"github.com/aaron-prindle/krmapiserver/plugin/pkg/admission/admit"
	"github.com/aaron-prindle/krmapiserver/plugin/pkg/admission/alwayspullimages"
	"github.com/aaron-prindle/krmapiserver/plugin/pkg/admission/antiaffinity"
	"github.com/aaron-prindle/krmapiserver/plugin/pkg/admission/defaulttolerationseconds"
	"github.com/aaron-prindle/krmapiserver/plugin/pkg/admission/deny"
	"github.com/aaron-prindle/krmapiserver/plugin/pkg/admission/eventratelimit"
	"github.com/aaron-prindle/krmapiserver/plugin/pkg/admission/exec"
	"github.com/aaron-prindle/krmapiserver/plugin/pkg/admission/extendedresourcetoleration"
	"github.com/aaron-prindle/krmapiserver/plugin/pkg/admission/gc"
	"github.com/aaron-prindle/krmapiserver/plugin/pkg/admission/imagepolicy"
	"github.com/aaron-prindle/krmapiserver/plugin/pkg/admission/limitranger"
	"github.com/aaron-prindle/krmapiserver/plugin/pkg/admission/namespace/autoprovision"
	"github.com/aaron-prindle/krmapiserver/plugin/pkg/admission/namespace/exists"
	"github.com/aaron-prindle/krmapiserver/plugin/pkg/admission/noderestriction"
	"github.com/aaron-prindle/krmapiserver/plugin/pkg/admission/nodetaint"
	"github.com/aaron-prindle/krmapiserver/plugin/pkg/admission/podnodeselector"
	"github.com/aaron-prindle/krmapiserver/plugin/pkg/admission/podpreset"
	"github.com/aaron-prindle/krmapiserver/plugin/pkg/admission/podtolerationrestriction"
	podpriority "github.com/aaron-prindle/krmapiserver/plugin/pkg/admission/priority"
	"github.com/aaron-prindle/krmapiserver/plugin/pkg/admission/resourcequota"
	"github.com/aaron-prindle/krmapiserver/plugin/pkg/admission/security/podsecuritypolicy"
	"github.com/aaron-prindle/krmapiserver/plugin/pkg/admission/securitycontext/scdeny"
	"github.com/aaron-prindle/krmapiserver/plugin/pkg/admission/serviceaccount"
	"github.com/aaron-prindle/krmapiserver/plugin/pkg/admission/storage/persistentvolume/label"
	"github.com/aaron-prindle/krmapiserver/plugin/pkg/admission/storage/persistentvolume/resize"
	"github.com/aaron-prindle/krmapiserver/plugin/pkg/admission/storage/storageclass/setdefault"
	"github.com/aaron-prindle/krmapiserver/plugin/pkg/admission/storage/storageobjectinuseprotection"

	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apimachinery/pkg/util/sets"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/admission"
	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/admission/plugin/namespace/lifecycle"
	mutatingwebhook "github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/admission/plugin/webhook/mutating"
	validatingwebhook "github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/admission/plugin/webhook/validating"
	utilfeature "github.com/aaron-prindle/krmapiserver/included/k8s.io/apiserver/pkg/util/feature"
	"github.com/aaron-prindle/krmapiserver/pkg/features"
)

// AllOrderedPlugins is the list of all the plugins in order.
var AllOrderedPlugins = []string{
	admit.PluginName,                        // AlwaysAdmit
	autoprovision.PluginName,                // NamespaceAutoProvision
	lifecycle.PluginName,                    // NamespaceLifecycle
	exists.PluginName,                       // NamespaceExists
	scdeny.PluginName,                       // SecurityContextDeny
	antiaffinity.PluginName,                 // LimitPodHardAntiAffinityTopology
	podpreset.PluginName,                    // PodPreset
	limitranger.PluginName,                  // LimitRanger
	serviceaccount.PluginName,               // ServiceAccount
	noderestriction.PluginName,              // NodeRestriction
	nodetaint.PluginName,                    // TaintNodesByCondition
	alwayspullimages.PluginName,             // AlwaysPullImages
	imagepolicy.PluginName,                  // ImagePolicyWebhook
	podsecuritypolicy.PluginName,            // PodSecurityPolicy
	podnodeselector.PluginName,              // PodNodeSelector
	podpriority.PluginName,                  // Priority
	defaulttolerationseconds.PluginName,     // DefaultTolerationSeconds
	podtolerationrestriction.PluginName,     // PodTolerationRestriction
	exec.DenyEscalatingExec,                 // DenyEscalatingExec
	exec.DenyExecOnPrivileged,               // DenyExecOnPrivileged
	eventratelimit.PluginName,               // EventRateLimit
	extendedresourcetoleration.PluginName,   // ExtendedResourceToleration
	label.PluginName,                        // PersistentVolumeLabel
	setdefault.PluginName,                   // DefaultStorageClass
	storageobjectinuseprotection.PluginName, // StorageObjectInUseProtection
	gc.PluginName,                           // OwnerReferencesPermissionEnforcement
	resize.PluginName,                       // PersistentVolumeClaimResize
	mutatingwebhook.PluginName,              // MutatingAdmissionWebhook
	validatingwebhook.PluginName,            // ValidatingAdmissionWebhook
	resourcequota.PluginName,                // ResourceQuota
	deny.PluginName,                         // AlwaysDeny
}

// RegisterAllAdmissionPlugins registers all admission plugins and
// sets the recommended plugins order.
func RegisterAllAdmissionPlugins(plugins *admission.Plugins) {
	admit.Register(plugins) // DEPRECATED as no real meaning
	alwayspullimages.Register(plugins)
	antiaffinity.Register(plugins)
	defaulttolerationseconds.Register(plugins)
	deny.Register(plugins) // DEPRECATED as no real meaning
	eventratelimit.Register(plugins)
	exec.Register(plugins)
	extendedresourcetoleration.Register(plugins)
	gc.Register(plugins)
	imagepolicy.Register(plugins)
	limitranger.Register(plugins)
	autoprovision.Register(plugins)
	exists.Register(plugins)
	noderestriction.Register(plugins)
	nodetaint.Register(plugins)
	label.Register(plugins) // DEPRECATED, future PVs should not rely on labels for zone topology
	podnodeselector.Register(plugins)
	podpreset.Register(plugins)
	podtolerationrestriction.Register(plugins)
	resourcequota.Register(plugins)
	podsecuritypolicy.Register(plugins)
	podpriority.Register(plugins)
	scdeny.Register(plugins)
	serviceaccount.Register(plugins)
	setdefault.Register(plugins)
	resize.Register(plugins)
	storageobjectinuseprotection.Register(plugins)
}

// DefaultOffAdmissionPlugins get admission plugins off by default for kube-apiserver.
func DefaultOffAdmissionPlugins() sets.String {
	defaultOnPlugins := sets.NewString(
		lifecycle.PluginName,                    //NamespaceLifecycle
		limitranger.PluginName,                  //LimitRanger
		serviceaccount.PluginName,               //ServiceAccount
		setdefault.PluginName,                   //DefaultStorageClass
		resize.PluginName,                       //PersistentVolumeClaimResize
		defaulttolerationseconds.PluginName,     //DefaultTolerationSeconds
		mutatingwebhook.PluginName,              //MutatingAdmissionWebhook
		validatingwebhook.PluginName,            //ValidatingAdmissionWebhook
		resourcequota.PluginName,                //ResourceQuota
		storageobjectinuseprotection.PluginName, //StorageObjectInUseProtection
	)

	if utilfeature.DefaultFeatureGate.Enabled(features.PodPriority) {
		defaultOnPlugins.Insert(podpriority.PluginName) //PodPriority
	}

	if utilfeature.DefaultFeatureGate.Enabled(features.TaintNodesByCondition) {
		defaultOnPlugins.Insert(nodetaint.PluginName) //TaintNodesByCondition
	}

	return sets.NewString(AllOrderedPlugins...).Difference(defaultOnPlugins)
}
