package terraform

import "regexp"

type Module struct {
	Name    string
	Source  string
	Version string
}

func (m *Module) GithubRepository() string {
	reModuleTfeSource := regexp.MustCompile(`tfe\.doctolib\.net\/[A-Za-z_1-9-]+\/(?P<name>[A-Za-z_1-9-]+)\/(?P<provider>[A-Za-z_1-9_]+)$`)
	if reModuleTfeSource.Match([]byte(m.Source)) {
		matches := reModuleTfeSource.FindStringSubmatch(m.Source)
		provider := matches[reModuleTfeSource.SubexpIndex("provider")]
		name := matches[reModuleTfeSource.SubexpIndex("name")]
		return "terraform-" + provider + "-" + name
	}
	return ""
}

func (m *Module) ModuleName() string {
	reModuleTfeSource := regexp.MustCompile(`tfe\.doctolib\.net\/[A-Za-z_1-9-]+\/(?P<name>[A-Za-z_1-9-]+)\/(?P<provider>[A-Za-z_1-9_]+)$`)
	if reModuleTfeSource.Match([]byte(m.Source)) {
		matches := reModuleTfeSource.FindStringSubmatch(m.Source)
		provider := matches[reModuleTfeSource.SubexpIndex("provider")]
		name := matches[reModuleTfeSource.SubexpIndex("name")]
		return name + "/" + provider
	}
	return ""
}

func (m *Module) IsTfeModule() bool {
	reModuleTfeSource := regexp.MustCompile(`tfe\.doctolib\.net\/[A-Za-z_1-9-]+\/[A-Za-z_1-9-]+\/[A-Za-z_1-9_]+$`)
	return reModuleTfeSource.MatchString(m.Source)
}
