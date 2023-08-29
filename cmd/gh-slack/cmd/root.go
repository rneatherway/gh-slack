package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

const sendConfigExample = `
  # Example configuration (add to gh's configuration file at $HOME/.config/gh/config.yml):
  extensions:
    slack:
      team: foo
      channel: ops
      bot: robot        # Can be a user id (most reliable), bot profile name or username`

var rootCmd = &cobra.Command{
	SilenceUsage:  true,
	SilenceErrors: true,
	Use:           "gh-slack [command]",
	Short:         "Command line tool for interacting with Slack through gh cli",
	Long:          `A command line tool for interacting with Slack through the gh cli.`,
	Example: `  gh-slack -i <issue-url> <slack-permalink>  # defaults to read command
  gh-slack read <slack-permalink>
  gh-slack read -i <issue-url> <slack-permalink>
  gh-slack send -m <message> -c <channel-name> -t <team-name>
  ` + sendConfigExample,
}

func Execute() error {
	cmd, _, err := rootCmd.Find(os.Args[1:])
	if err != nil || cmd == nil {
		args := append([]string{"read"}, os.Args[1:]...)
		rootCmd.SetArgs(args)
	}
	return rootCmd.Execute()
}

var verbose bool = false

func init() {
	rootCmd.AddCommand(readCmd)
	rootCmd.AddCommand(sendCmd)
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Show verbose debug information")
	rootCmd.SetHelpTemplate(rootCmdUsageTemplate)
	rootCmd.SetUsageTemplate(rootCmdUsageTemplate)
}

const rootCmdUsageTemplate string = `Usage:{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}

  If no command is specified, the default is "read". The default command also requires a permalink argument <START> for the first message to fetch.
  Use "gh-slack read --help" for more information about the default command behaviour.{{if gt (len .Aliases) 0}}
Aliases:
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

Examples:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}{{$cmds := .Commands}}{{if eq (len .Groups) 0}}

Available Commands:{{range $cmds}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
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
