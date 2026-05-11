package artifact

import (
	"path/filepath"
	"strings"
)

type Paths struct {
	WorkDir string
	Video   string
	Audio   string
	Output  string
	Key     string
	Name    string
}

func Build(dataDir string, jobID string, title string, container string) Paths {
	name := SafeName(title)
	if name == "" {
		name = jobID
	}
	fileName := name + "." + container
	workDir := filepath.Join(dataDir, "jobs", jobID)
	return Paths{
		WorkDir: workDir,
		Video:   filepath.Join(workDir, "video."+container),
		Audio:   filepath.Join(workDir, "audio."+audioExt(container)),
		Output:  filepath.Join(dataDir, "artifacts", jobID, fileName),
		Key:     "artifacts/" + jobID + "/" + fileName,
		Name:    fileName,
	}
}

func SafeName(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	var builder strings.Builder
	for _, r := range value {
		if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' {
			builder.WriteRune(r)
			continue
		}
		if r == '-' || r == '_' || r == '.' {
			builder.WriteRune(r)
			continue
		}
		if builder.Len() > 0 && builder.String()[builder.Len()-1] != '_' {
			builder.WriteByte('_')
		}
	}
	return strings.Trim(builder.String(), "_.")
}

func audioExt(container string) string {
	if container == "webm" {
		return "webm"
	}
	return "m4a"
}
