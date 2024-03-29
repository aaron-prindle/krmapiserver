/*
Copyright 2018 The Kubernetes Authors.

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

package volumepathhandler

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/aaron-prindle/krmapiserver/included/k8s.io/klog"

	"github.com/aaron-prindle/krmapiserver/included/k8s.io/apimachinery/pkg/types"
)

const (
	losetupPath           = "losetup"
	ErrDeviceNotFound     = "device not found"
	ErrDeviceNotSupported = "device not supported"
)

// BlockVolumePathHandler defines a set of operations for handling block volume-related operations
type BlockVolumePathHandler interface {
	// MapDevice creates a symbolic link to block device under specified map path
	MapDevice(devicePath string, mapPath string, linkName string) error
	// UnmapDevice removes a symbolic link to block device under specified map path
	UnmapDevice(mapPath string, linkName string) error
	// RemovePath removes a file or directory on specified map path
	RemoveMapPath(mapPath string) error
	// IsSymlinkExist retruns true if specified symbolic link exists
	IsSymlinkExist(mapPath string) (bool, error)
	// GetDeviceSymlinkRefs searches symbolic links under global map path
	GetDeviceSymlinkRefs(devPath string, mapPath string) ([]string, error)
	// FindGlobalMapPathUUIDFromPod finds {pod uuid} symbolic link under globalMapPath
	// corresponding to map path symlink, and then return global map path with pod uuid.
	FindGlobalMapPathUUIDFromPod(pluginDir, mapPath string, podUID types.UID) (string, error)
	// AttachFileDevice takes a path to a regular file and makes it available as an
	// attached block device.
	AttachFileDevice(path string) (string, error)
	// GetLoopDevice returns the full path to the loop device associated with the given path.
	GetLoopDevice(path string) (string, error)
	// RemoveLoopDevice removes specified loopback device
	RemoveLoopDevice(device string) error
}

// NewBlockVolumePathHandler returns a new instance of BlockVolumeHandler.
func NewBlockVolumePathHandler() BlockVolumePathHandler {
	var volumePathHandler VolumePathHandler
	return volumePathHandler
}

// VolumePathHandler is path related operation handlers for block volume
type VolumePathHandler struct {
}

// MapDevice creates a symbolic link to block device under specified map path
func (v VolumePathHandler) MapDevice(devicePath string, mapPath string, linkName string) error {
	// Example of global map path:
	//   globalMapPath/linkName: plugins/kubernetes.io/{PluginName}/{DefaultKubeletVolumeDevicesDirName}/{volumePluginDependentPath}/{podUid}
	//   linkName: {podUid}
	//
	// Example of pod device map path:
	//   podDeviceMapPath/linkName: pods/{podUid}/{DefaultKubeletVolumeDevicesDirName}/{escapeQualifiedPluginName}/{volumeName}
	//   linkName: {volumeName}
	if len(devicePath) == 0 {
		return fmt.Errorf("Failed to map device to map path. devicePath is empty")
	}
	if len(mapPath) == 0 {
		return fmt.Errorf("Failed to map device to map path. mapPath is empty")
	}
	if !filepath.IsAbs(mapPath) {
		return fmt.Errorf("The map path should be absolute: map path: %s", mapPath)
	}
	klog.V(5).Infof("MapDevice: devicePath %s", devicePath)
	klog.V(5).Infof("MapDevice: mapPath %s", mapPath)
	klog.V(5).Infof("MapDevice: linkName %s", linkName)

	// Check and create mapPath
	_, err := os.Stat(mapPath)
	if err != nil && !os.IsNotExist(err) {
		klog.Errorf("cannot validate map path: %s", mapPath)
		return err
	}
	if err = os.MkdirAll(mapPath, 0750); err != nil {
		return fmt.Errorf("Failed to mkdir %s, error %v", mapPath, err)
	}
	// Remove old symbolic link(or file) then create new one.
	// This should be done because current symbolic link is
	// stale across node reboot.
	linkPath := filepath.Join(mapPath, string(linkName))
	if err = os.Remove(linkPath); err != nil && !os.IsNotExist(err) {
		return err
	}
	err = os.Symlink(devicePath, linkPath)
	return err
}

// UnmapDevice removes a symbolic link associated to block device under specified map path
func (v VolumePathHandler) UnmapDevice(mapPath string, linkName string) error {
	if len(mapPath) == 0 {
		return fmt.Errorf("Failed to unmap device from map path. mapPath is empty")
	}
	klog.V(5).Infof("UnmapDevice: mapPath %s", mapPath)
	klog.V(5).Infof("UnmapDevice: linkName %s", linkName)

	// Check symbolic link exists
	linkPath := filepath.Join(mapPath, string(linkName))
	if islinkExist, checkErr := v.IsSymlinkExist(linkPath); checkErr != nil {
		return checkErr
	} else if !islinkExist {
		klog.Warningf("Warning: Unmap skipped because symlink does not exist on the path: %v", linkPath)
		return nil
	}
	err := os.Remove(linkPath)
	return err
}

// RemoveMapPath removes a file or directory on specified map path
func (v VolumePathHandler) RemoveMapPath(mapPath string) error {
	if len(mapPath) == 0 {
		return fmt.Errorf("Failed to remove map path. mapPath is empty")
	}
	klog.V(5).Infof("RemoveMapPath: mapPath %s", mapPath)
	err := os.RemoveAll(mapPath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// IsSymlinkExist returns true if specified file exists and the type is symbolik link.
// If file doesn't exist, or file exists but not symbolic link, return false with no error.
// On other cases, return false with error from Lstat().
func (v VolumePathHandler) IsSymlinkExist(mapPath string) (bool, error) {
	fi, err := os.Lstat(mapPath)
	if err == nil {
		// If file exits and it's symbolic link, return true and no error
		if fi.Mode()&os.ModeSymlink == os.ModeSymlink {
			return true, nil
		}
		// If file exits but it's not symbolic link, return fale and no error
		return false, nil
	}
	// If file doesn't exist, return false and no error
	if os.IsNotExist(err) {
		return false, nil
	}
	// Return error from Lstat()
	return false, err
}

// GetDeviceSymlinkRefs searches symbolic links under global map path
func (v VolumePathHandler) GetDeviceSymlinkRefs(devPath string, mapPath string) ([]string, error) {
	var refs []string
	files, err := ioutil.ReadDir(mapPath)
	if err != nil {
		return nil, fmt.Errorf("Directory cannot read %v", err)
	}
	for _, file := range files {
		if file.Mode()&os.ModeSymlink != os.ModeSymlink {
			continue
		}
		filename := file.Name()
		fp, err := os.Readlink(filepath.Join(mapPath, filename))
		if err != nil {
			return nil, fmt.Errorf("Symbolic link cannot be retrieved %v", err)
		}
		klog.V(5).Infof("GetDeviceSymlinkRefs: filepath: %v, devPath: %v", fp, devPath)
		if fp == devPath {
			refs = append(refs, filepath.Join(mapPath, filename))
		}
	}
	klog.V(5).Infof("GetDeviceSymlinkRefs: refs %v", refs)
	return refs, nil
}

// FindGlobalMapPathUUIDFromPod finds {pod uuid} symbolic link under globalMapPath
// corresponding to map path symlink, and then return global map path with pod uuid.
// ex. mapPath symlink: pods/{podUid}}/{DefaultKubeletVolumeDevicesDirName}/{escapeQualifiedPluginName}/{volumeName} -> /dev/sdX
//     globalMapPath/{pod uuid}: plugins/kubernetes.io/{PluginName}/{DefaultKubeletVolumeDevicesDirName}/{volumePluginDependentPath}/{pod uuid} -> /dev/sdX
func (v VolumePathHandler) FindGlobalMapPathUUIDFromPod(pluginDir, mapPath string, podUID types.UID) (string, error) {
	var globalMapPathUUID string
	// Find symbolic link named pod uuid under plugin dir
	err := filepath.Walk(pluginDir, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if (fi.Mode()&os.ModeSymlink == os.ModeSymlink) && (fi.Name() == string(podUID)) {
			klog.V(5).Infof("FindGlobalMapPathFromPod: path %s, mapPath %s", path, mapPath)
			if res, err := compareSymlinks(path, mapPath); err == nil && res {
				globalMapPathUUID = path
			}
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	klog.V(5).Infof("FindGlobalMapPathFromPod: globalMapPathUUID %s", globalMapPathUUID)
	// Return path contains global map path + {pod uuid}
	return globalMapPathUUID, nil
}

func compareSymlinks(global, pod string) (bool, error) {
	devGlobal, err := os.Readlink(global)
	if err != nil {
		return false, err
	}
	devPod, err := os.Readlink(pod)
	if err != nil {
		return false, err
	}
	klog.V(5).Infof("CompareSymlinks: devGloBal %s, devPod %s", devGlobal, devPod)
	if devGlobal == devPod {
		return true, nil
	}
	return false, nil
}
