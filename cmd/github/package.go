package main

import "fmt"

type PackageVersion struct {
	Version string `json:"version"`
}

type Package struct {
	Name           string          `json:"name"`
	Type           string          `json:"type"`
	PackageType    string          `json:"package_type"`
	Version        string          `json:"version"`
	HtmlURL        string          `json:"html_url"`
	Owner          *User           `json:"owner"`
	Creator        *User           `json:"creator"`
	PackageVersion *PackageVersion `json:"package_version"`
}

// packageType covers forgejo's "type" and github's "package_type".
func (p *Package) packageType() string {
	if p.Type != "" {
		return p.Type
	}
	return p.PackageType
}

// packageVersion covers forgejo's top-level version and github's nested
// package_version.version.
func (p *Package) packageVersion() string {
	if p.Version != "" {
		return p.Version
	}
	if p.PackageVersion != nil {
		return p.PackageVersion.Version
	}
	return ""
}

type githubPackageHandler struct{}

func (g *githubPackageHandler) handlePackageEvent(data webhook) (messages []string) {
	p := data.Package
	if p == nil {
		return
	}
	origin := data.Repository.FullName
	if origin == "" && p.Owner != nil {
		// Packages uploaded to a user or org have no repository in the payload.
		origin = p.Owner.Login
	}
	actor := data.Sender.Login
	if actor == "" && p.Creator != nil {
		actor = p.Creator.Login
	}
	name := p.Name
	if version := p.packageVersion(); version != "" {
		name = fmt.Sprintf("%s:%s", name, version)
	}
	switch data.Action {
	case "created", "published":
		messages = append(messages, fmt.Sprintf(
			"[%s] %s published %s package %s - %s",
			origin,
			actor,
			p.packageType(),
			name,
			p.HtmlURL,
		))
	case "deleted":
		messages = append(messages, fmt.Sprintf(
			"[%s] %s deleted %s package %s",
			origin,
			actor,
			p.packageType(),
			name,
		))
	}
	return
}
