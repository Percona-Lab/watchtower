package validation

import (
	"strings"

	"github.com/containrrr/watchtower/pkg/container"
	"github.com/containrrr/watchtower/pkg/types"

	"golang.org/x/net/context"
)

func ValidateParams(client container.Client, params types.UpdateParams) error {
	containers, err := client.ListContainers(params.Filter)
	if err != nil {
		return err
	}
	if len(containers) == 0 {
		return types.NewValidationError("no containers found")
	}
	if params.NewImageName != "" {
		for _, c := range containers {
			if !c.IsPMM() {
				return types.NewValidationError("container is not a PMM server")
			}
		}
	}

	if !isImageAllowed(params.AllowedImageRepos, params.NewImageName) {
		return types.NewValidationError("image not allowed")
	}
	pullNeeded, err := client.PullNeeded(context.TODO(), containers[0])
	if err != nil {
		return err
	}
	// if pull is needed, we don't need to check for new image locally.
	if pullNeeded {
		return nil
	} else {
		hasNew, _, err := client.HasNewImage(context.TODO(), containers[0])
		if err != nil {
			return err
		}
		if !hasNew {
			return types.NewValidationError("no new image available")
		}
	}
	return nil
}

func isImageAllowed(repos []string, newImageName string) bool {
	if newImageName == "" || len(repos) == 0 {
		return true
	}
	for _, repo := range repos {
		if strings.HasPrefix(newImageName, repo) {
			return true
		}
	}
	return false
}