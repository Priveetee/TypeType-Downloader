package buildinfo

var Version = "0.1.0"
var Revision = "development"
var BuildTime = "unknown"

func ShortRevision() string {
	if len(Revision) <= 12 {
		return Revision
	}
	return Revision[:12]
}
