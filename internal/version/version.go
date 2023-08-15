package version

import (
	"fmt"

	"github.com/hibare/GoCommon/pkg/version"
	"github.com/hibare/GoS3Backup/internal/constants"
)

var (
	CurrentVersion = "0.0.0"
	V              = version.Version{
		CurrentVersion: fmt.Sprintf("v%s", CurrentVersion),
		GithubOwner:    constants.GithubOwner,
		GithubRepo:     constants.ProgramIdentifier,
	}
)
