package presenters

import (
	"fmt"
	"strings"

	"github.com/sammccord/flyctl/api"
)

type Releases struct {
	Releases []api.Release
	Release  *api.Release
}

func (p *Releases) APIStruct() interface{} {
	return p.Releases
}

func (p *Releases) FieldNames() []string {
	return []string{"Version", "Stable", "Type", "Status", "Description", "User", "Date"}
}

func (p *Releases) Records() []map[string]string {
	out := []map[string]string{}

	if p.Release != nil {
		p.Releases = append(p.Releases, *p.Release)
	}

	for _, release := range p.Releases {
		out = append(out, map[string]string{
			"Version":     fmt.Sprintf("v%d", release.Version),
			"Stable":      fmt.Sprintf("%t", release.Stable),
			"Status":      release.Status,
			"Type":        formatReleaseReason(release.Reason),
			"Description": formatReleaseDescription(release),
			"User":        release.User.Email,
			"Date":        FormatRelativeTime(release.CreatedAt),
		})
	}

	return out
}

func formatReleaseReason(reason string) string {
	switch reason {
	case "change_image":
		return "Image"
	case "change_secrets":
		return "Secrets"
	case "change_code", "change_source": // nodeproxy
		return "Code Change"
	}
	return reason
}

func formatReleaseDescription(r api.Release) string {
	if r.Reason == "change_image" && strings.HasPrefix(r.Description, "deploy image ") {
		return r.Description[13:]
	}
	return r.Description
}
