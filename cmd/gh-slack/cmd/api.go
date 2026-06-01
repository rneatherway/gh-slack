package cmd

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/cli/go-gh/v2/pkg/config"
	"github.com/rneatherway/gh-slack/internal/slackclient"
	"github.com/spf13/cobra"
)

var apiCmd = &cobra.Command{
	Use:   "api [verb] path",
	Short: "Send an API call to slack",
	Long:  "Send an API call to slack",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Read(nil)
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

		fileFlags, err := cmd.Flags().GetStringArray("file")
		if err != nil {
			return err
		}
		if len(fileFlags) > 1 {
			return fmt.Errorf("only one file upload is supported")
		}
		if len(fileFlags) > 0 && body != "" {
			return fmt.Errorf("--file cannot be used with --body")
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
			if body == "" && len(fileFlags) == 0 {
				verb = "GET"
			} else {
				verb = "POST"
			}
		} else {
			return fmt.Errorf("expected 1 or 2 arguments: verb and/or path, see help")
		}

		var response []byte
		if len(fileFlags) > 0 {
			if verb != "POST" {
				return fmt.Errorf("--file only supports POST requests")
			}

			fileParam, file, err := fileParamFromFlag(fileFlags[0])
			if err != nil {
				return err
			}
			defer file.Close()

			response, err = client.APIMultipart(path, mappedFields, fileParam)
		} else {
			response, err = client.API(verb, path, mappedFields, []byte(body))
		}
		if err != nil {
			return err
		}

		fmt.Println(string(response))
		return nil
	},
	Example: `  gh-slack api get conversations.list -f types=public_channel,private_channel
  gh-slack api post chat.postMessage -b '{"channel":"123","blocks":[...]}'
  gh-slack api post users.setPhoto -F image=@photo.jpg -f crop_x=0 -f crop_y=0`,
}

var fields []string
var body string
var files []string

func init() {
	apiCmd.Flags().StringArrayVarP(&fields, "field", "f", nil, "Fields to pass to the api call")
	apiCmd.Flags().StringVarP(&body, "body", "b", "", "Body to send as JSON")
	apiCmd.Flags().StringArrayVarP(&files, "file", "F", nil, "File to upload as multipart form data, in key=@path format")
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

func fileParamFromFlag(fileFlag string) (slackclient.FileParam, *os.File, error) {
	parts := strings.SplitN(fileFlag, "=", 2)

	if len(parts) != 2 || parts[0] == "" {
		return slackclient.FileParam{}, nil, fmt.Errorf("file '%s' must be in key=@path format", fileFlag)
	}
	if !strings.HasPrefix(parts[1], "@") || parts[1] == "@" {
		return slackclient.FileParam{}, nil, fmt.Errorf("file '%s' must be in key=@path format", fileFlag)
	}

	filePath := strings.TrimPrefix(parts[1], "@")
	file, err := os.Open(filePath)
	if err != nil {
		return slackclient.FileParam{}, nil, fmt.Errorf("opening file %q: %w", filePath, err)
	}

	return slackclient.FileParam{
		Fieldname: parts[0],
		Filename:  filepath.Base(filePath),
		Reader:    file,
	}, file, nil
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
