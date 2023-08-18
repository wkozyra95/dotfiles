package semver

import (
	"github.com/blang/semver"
)

func ShouldUpdate(current string, latest string) (bool, error) {
	currentSemver, currentSemverErr := semver.Make(current)
	if currentSemverErr != nil {
		return false, currentSemverErr
	}

	latestSemver, latestSemverErr := semver.Make(latest)
	if latestSemverErr != nil {
		return false, latestSemverErr
	}

	return latestSemver.GT(currentSemver), nil
}
