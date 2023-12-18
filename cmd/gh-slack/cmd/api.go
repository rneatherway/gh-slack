package cmd

import (
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/cli/go-gh/pkg/config"
	"github.com/rneatherway/gh-slack/internal/slackclient"
	"github.com/spf13/cobra"
)

var apiCmd = &cobra.Command{
	Use:   "api [verb] path",
	Short: "Send an API call to slack",
	Long:  "Send an API call to slack",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Read()
		if err != nil {
			return err
		}

		team, err := getFlagOrElseConfig(cfg, cmd.Flags(), "team")
		if err != nil {
			return err
		}

		fields, err := cmd.Flags().GetStringArray("field")
		if err != nil {
			return err
		}

		mappedFields, err := mapFields(fields)
		if err != nil {
			return err
		}

		logger := log.New(io.Discard, "", log.LstdFlags)
		if verbose {
			logger = log.Default()
		}

		client, err := slackclient.New(team, logger)
		if err != nil {
			return err
		}

		var verb, path string
		if len(args) == 2 {
			verb = strings.ToUpper(args[0])
			path = args[1]
		} else if len(args) == 1 {
			path = args[0]
			if body == "" {
				verb = "GET"
			} else {
				verb = "POST"
			}
		} else {
			return fmt.Errorf("expected 1 or 2 arguments: verb and/or path, see help")
		}

		response, err := client.API(verb, path, mappedFields, []byte(body))
		if err != nil {
			return err
		}

		fmt.Println(string(response))
		return nil
	},
	Example: `  gh-slack api get conversations.list -f types=public_channel,private_channel
  gh-slack api post chat.postMessage -b '{"channel":"123","blocks":[...]}`,
}

var fields []string
var body string

func init() {
	apiCmd.Flags().StringArrayVarP(&fields, "field", "f", nil, "Fields to pass to the api call")
	apiCmd.Flags().StringVarP(&body, "body", "b", "", "Body to send as JSON")
	apiCmd.Flags().StringP("team", "t", "", "Slack team name (required here or in config)")
	apiCmd.SetHelpTemplate(apiCmdUsage)
	apiCmd.SetUsageTemplate(apiCmdUsage)
}

func mapFields(fields []string) (map[string]string, error) {
	mappedFields := map[string]string{}

	for _, field := range fields {
		parts := strings.SplitN(field, "=", 2)

		if len(parts) != 2 || parts[1] == "" {
			return nil, fmt.Errorf("field '%s' is missing a value", field)
		}

		mappedFields[parts[0]] = parts[1]
	}

	return mappedFields, nil
}

const apiCmdUsage string = `Usage:{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}}{{end}}{{if gt (len .Aliases) 0}}
Aliases:
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

Examples:
{{.Example}}

The verb is optional:
- If no body is sent, GET will be used.
- If a body is sent, POST will be used.
{{end}}{{if .HasAvailableSubCommands}}{{$cmds := .Commands}}{{if eq (len .Groups) 0}}

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
