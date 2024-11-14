package main

import "fmt"

func (m *model) headerView() string {
	title := titleStyle.Render("🚀 Launchrail")
	desc := headerStyle.Render("Risk-neutral trajectory simulation for sounding rockets via the Black-Scholes model.\n'ctrl+c' or 'q' to quit.")
	return fmt.Sprintf("%s\n%s\n", title, desc)
}

func (m *model) footerView() string {
	githubText := footerLinkStyle.Render(m.cfg.App.Repo)
	licenseText := footerStyle.Render(m.cfg.App.License)
	versionText := footerStyle.Render(m.cfg.App.Version)
	return fmt.Sprintf("%s | %s | %s\n", versionText, licenseText, githubText)
}
