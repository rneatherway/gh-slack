package cmd

import (
	"fmt"
	"net/url"

	"github.com/cli/go-gh/pkg/config"
	"github.com/rneatherway/slack"
	"github.com/spf13/cobra"
)

var authCmd = &cobra.Command{
	Use:   "auth [flags]",
	Short: "Prints authentication information for the Slack API (treat output as secret)",
	Long:  "Prints authentication information for the Slack API (treat output as secret).",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Read()
		if err != nil {
			return err
		}

		team, err := getFlagOrElseConfig(cfg, cmd.Flags(), "team")
		if err != nil {
			return err
		}

		auth, err := slack.GetCookieAuth(team)
		if err != nil {
			return err
		}

		vals := url.Values{}
		for k, v := range auth.Cookies {
			vals.Add(k, v)
		}

		fmt.Printf("export SLACK_TOKEN=%s\n", auth.Token)
		fmt.Printf("export SLACK_COOKIES=%s\n", vals.Encode())
		return nil
	},
	Example: `  eval $(gh-slack auth [-t <team-name>])

  # Example configuration (add to gh's configuration file at $HOME/.config/gh/config.yml):
  extensions:
    slack:
      team: foo`,
}

func init() {
	authCmd.Flags().StringP("team", "t", "", "Slack team name (required here or in config)")
	authCmd.SetHelpTemplate(authCmdUsageTemplate)
	authCmd.SetUsageTemplate(authCmdUsageTemplate)
}

const authCmdUsageTemplate string = `Usage:{{if .Runnable}}
{{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
{{.CommandPath}}{{end}}{{if gt (len .Aliases) 0}}
Aliases:
{{.NameAndAliases}}{{end}}{{if .HasExample}}

Security:
  Treat the output of this command as secret and do not share it with anyone!
  It can be used to impersonate you. If you suspect it has been compromised,
  log out of the Slack app to revoke the token and cookies.

Examples:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}{{$cmds := .Commands}}{{if eq (len .Groups) 0}}

Available Commands:{{range $cmds}}{{if (or .IsAvailableCommand)}}
{{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{else}}{{range $group := .Groups}}

{{.Title}}{{range $cmds}}{{if (and (eq .GroupID $group.ID) (or .IsAvailableCommand (eq .Name "help")))}}
{{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if not .AllChildCommandsHaveGroup}}

Additional Commands:{{range $cmds}}{{if (and (eq .GroupID "") (or .IsAvailableCommand (eq .Name "help")))}}
{{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

Global Flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
{{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`
