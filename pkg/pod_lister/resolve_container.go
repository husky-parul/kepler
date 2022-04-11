/*
Copyright 2021.

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

package pod_lister

import (
	"encoding/binary"
	"fmt"
	"io/fs"
	"path/filepath"
	"regexp"
	"strings"

	"golang.org/x/sys/unix"

	bpf "github.com/iovisor/gobpf/bcc"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	cache      = map[uint64]string{}
	cgroupMaps = map[string]*metav1.ObjectMeta{}
	re         = regexp.MustCompile(`crio-(.*?)\.scope`)
	cgroupPath = "/sys/fs/cgroup"
	byteOrder  binary.ByteOrder
)

func init() {
	byteOrder = bpf.GetHostByteOrder()
}

func CgroupToPod(path string) (*metav1.ObjectMeta, error) {
	if re.MatchString(path) {
		containerId := ""
		sub := re.FindAllString(path, -1)
		for _, element := range sub {
			containerId = strings.Trim(element, "crio-")
			containerId = strings.Trim(containerId, ".scope")
		}
		containerId = "cri-o://" + containerId
		if meta, ok := cgroupMaps[containerId]; ok {
			return meta, nil
		}
		podLister := KubeletPodLister{}
		pods, err := podLister.ListPods()
		if err != nil {
			return nil, err
		}
		cgroupMaps = map[string]*metav1.ObjectMeta{}
		for _, pod := range *pods {
			meta := &metav1.ObjectMeta{
				Name:      pod.Name,
				Namespace: pod.Namespace,
			}
			statuses := pod.Status.ContainerStatuses
			for _, status := range statuses {
				id := status.ContainerID
				cgroupMaps[id] = meta
			}
			statuses = pod.Status.InitContainerStatuses
			for _, status := range statuses {
				id := status.ContainerID
				cgroupMaps[id] = meta
			}
		}
		if meta, ok := cgroupMaps[containerId]; ok {
			return meta, nil
		}
	}
	return nil, nil
}

// CgroupIdToName uses cgroupfs to get cgroup path from id
// it needs cgroup v2 (per https://github.com/iovisor/bpftrace/issues/950) and kernel 4.18+ (https://github.com/torvalds/linux/commit/bf6fa2c893c5237b48569a13fa3c673041430b6c)
func CgroupIdToName(cgroupId uint64) (string, error) {
	if p, ok := cache[cgroupId]; ok {
		return p, nil
	}

	err := filepath.WalkDir(cgroupPath, func(path string, dentry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !dentry.IsDir() {
			return nil
		}
		handle, _, err := unix.NameToHandleAt(unix.AT_FDCWD, path, 0)
		if err != nil {
			return fmt.Errorf("Error resolving handle: %v", err)
		}
		cache[byteOrder.Uint64(handle.Bytes())] = path
		return nil
	})

	if err != nil {
		return "", fmt.Errorf("failed to find cgroup id: %v", err)
	}
	if p, ok := cache[cgroupId]; ok {
		return p, nil
	}
	cache[cgroupId] = "unknown"
	return cache[cgroupId], nil
}
