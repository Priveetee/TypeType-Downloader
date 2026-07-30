package buildinfo

var Version = "1.2.4-dev"
var Revision = "development"
var BuildTime = "unknown"

func ShortRevision() string {
	if len(Revision) <= 12 {
		return Revision
	}
	return Revision[:12]
}
