package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	PolicyRecursive string = "recursive"
	PolicyRootOnly         = "root-only"
)

type ChownRequest struct {
	// The name of chown
	Name string
	// The target path
	Path string
	// The user (uid) to set for the path
	User int
	// The group (gid) to set for the path
	Group int
	// The mode of file path to change
	Mode os.FileMode
	// The policy for chown
	Policy string
}

const (
	annotationPrefix    string = "com.launchplatform.oci-hooks.mount-chown."
	annotationPathArg   string = "path"
	annotationOwnerArg  string = "owner"
	annotationPolicyArg string = "policy"
	annotationModeArg   string = "mode"
)

func parseOwner(owner string) (int, int, error) {
	parts := strings.Split(owner, ":")
	if len(parts) < 1 || len(parts) > 2 {
		return 0, 0, fmt.Errorf("Expected only one or two parts in the owner but got %d instead", len(parts))
	}
	uid, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, err
	}
	if len(parts) == 1 {
		return uid, 0, nil
	}
	gid, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, err
	}
	return uid, gid, nil
}

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
			request = ChownRequest{Name: name, User: -1, Group: -1}
		}
		if chownArg == annotationPathArg {
			absPath, err := filepath.Abs(value)
			if err != nil {
				log.Fatal(err)
			}
			if !filepath.IsAbs(value) || absPath != value {
				log.Warnf("Invalid path argument %s for %s, only abs path allowed, ignored", value, name)
				continue
			}
			request.Path = value
		} else if chownArg == annotationOwnerArg {
			uid, gid, err := parseOwner(value)
			if err != nil {
				log.Warnf("Invalid owner argument for %s with error %s, ignored", name, err)
				continue
			}
			if uid < 0 || gid < 0 {
				log.Warnf("Invalid owner argument for %s with negative uid or gid, ignored", name)
				continue
			}
			request.User = uid
			request.Group = gid
		} else if chownArg == annotationPolicyArg {
			request.Policy = value
		} else if chownArg == annotationModeArg {
			mode, err := strconv.ParseInt(value, 8, 32)
			if err != nil {
				log.Warnf("Invalid mode argument %s for request %s, needs to be an octal integer, ignored", value, name)
				continue
			}
			request.Mode = os.FileMode(mode)
		} else {
			log.Warnf("Invalid chown argument %s for request %s, ignored", chownArg, name)
			continue
		}
		requests[name] = request
	}

	filteredRequests := map[string]ChownRequest{}
	for _, request := range requests {
		var emptyValue = false
		if request.Path == "" {
			log.Warnf("Empty path argument value for %s, ignored", request.Name)
			emptyValue = true
		}
		if request.User == -1 || request.Group == -1 {
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
		filteredRequests[request.Path] = request
	}
	return filteredRequests
}
