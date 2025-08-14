package admincommands

type AdminCommand struct {
	DevEnvironment DevEnvironment `cmd:"dev-environment" help:"Manage Dev Environments for Devgraph"`
}
