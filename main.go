/*
Copyright © 2025 WizardOps LLC headwizard@wizardops.dev
*/
package main

import "github.com/wizardopstech/conjure/cmd"

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	cmd.Execute(version, commit, date)
}
