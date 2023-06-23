package main

import (
	log "github.com/sirupsen/logrus"
	"strings"
)

const (
	PolicyRecursive string = "recursive"
	PolicyRootOnly         = "root-only"
)

type ChownRequest struct {
	// The name of chown
	Name string
	// The "destination" filed of mount point to chown
	MountPoint string
	// The owner to recursively set for the mount point
	Owner string
	// The policy
	Policy string
}

const (
	annotationPrefix        string = "com.launchplatform.oci-hooks.mount-chown."
	annotationMountPointArg string = "mount-point"
	annotationOwnerArg      string = "owner"
	annotationPolicyArg     string = "policy"
)

func parseChownRequests(annotations map[string]string) map[string]ChownRequest {
	requests := map[string]ChownRequest{}
	for key, value := range annotations {
		if !strings.HasPrefix(key, annotationPrefix) {
			continue
		}
		keySuffix := key[len(annotationPrefix):]
		parts := strings.Split(keySuffix, ".")
		name, chownArg := parts[0], parts[1]
		request, ok := requests[name]
		if !ok {
			request = ChownRequest{Name: name}
		}
		if chownArg == annotationMountPointArg {
			request.MountPoint = value
		} else if chownArg == annotationOwnerArg {
			request.Owner = value
		} else if chownArg == annotationPolicyArg {
			request.Policy = value
		} else {
			log.Warnf("Invalid chown argument %s for request %s, ignored", chownArg, name)
			continue
		}
		requests[name] = request
	}

	// Convert map from using name as the key to use mount-point instead
	mountPointRequests := map[string]ChownRequest{}
	for _, request := range requests {
		var emptyValue = false
		if request.MountPoint == "" {
			log.Warnf("Empty mount-point request argument value for %s, ignored", request.Name)
			emptyValue = true
		}
		if request.Owner == "" {
			log.Warnf("Empty owner argument value for %s, ignored", request.Name)
			emptyValue = true
		}
		if request.Policy != "" && request.Policy != PolicyRecursive && request.Policy != PolicyRootOnly {
			log.Warnf("Invalid policy argument value %s for %s, ignored", request.Policy, request.Name)
			emptyValue = true
		}
		if emptyValue {
			continue
		}
		mountPointRequests[request.MountPoint] = request
	}
	return mountPointRequests
}
