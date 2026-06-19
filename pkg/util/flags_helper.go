package util

// FlagValues holds plain (non-pointer) values for constructing GeneralFlags and GitFlags.
// This avoids the footgun of manually taking addresses of temporaries.
type FlagValues struct {
	// GeneralFlags fields
	App                  string
	Namespace            string
	Env                  string
	Base                 string
	Cluster              string
	Image                string
	ImageFile            string
	ImageHistoryFile     string
	PathScheme           string
	ExceptionalAppsFile  string
	SrcPath              string
	VersionInfo          string

	// GitFlags fields
	Push       bool
	Shallow    bool
	SshKeyPath string
	User       string
	Email      string
	Url        string
	Branch     string
	Token      string
}

// NewFlagsFromValues constructs GeneralFlags and GitFlags pointer structs from plain values.
// The MCP server must use this instead of GetGeneralFlags()/GetGitFlags() to avoid
// mutating the global flag.CommandLine.
func NewFlagsFromValues(v FlagValues) (GeneralFlags, GitFlags) {
	falseB := false

	gf := GeneralFlags{
		Help:                &falseB,
		App:                 &v.App,
		Namespace:           &v.Namespace,
		Env:                 &v.Env,
		Base:                &v.Base,
		Cluster:             &v.Cluster,
		Image:               &v.Image,
		ImageFile:           &v.ImageFile,
		ImageHistoryFile:    &v.ImageHistoryFile,
		PathScheme:          &v.PathScheme,
		ExceptionalAppsFile: &v.ExceptionalAppsFile,
		SrcPath:             &v.SrcPath,
		VersionInfo:         &v.VersionInfo,
	}

	gitF := GitFlags{
		Push:       &v.Push,
		Shallow:    &v.Shallow,
		SshKeyPath: &v.SshKeyPath,
		User:       &v.User,
		Email:      &v.Email,
		Url:        &v.Url,
		Branch:     &v.Branch,
		Token:      &v.Token,
	}

	return gf, gitF
}
