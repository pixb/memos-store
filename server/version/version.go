package version

// Version is the service current released version.
// Semantic versioning: https://semver.org/
var Version = "0.24.3"

// DevVersion is the service current development version.
var DevVersion = "0.24.3"

func GetCurrentVersion(mode string) string {
	if mode == "dev" || mode == "demo" {
		return DevVersion
	}
	return Version
}
