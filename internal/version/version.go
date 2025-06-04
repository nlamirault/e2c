package version

// Version is the current version of the application.
// This variable is typically set during build time.
var Version = "dev"

// GetVersion returns the current version of the application.
func GetVersion() string {
	return Version
}