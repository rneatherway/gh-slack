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
	Use:   "api verb path",
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

		if len(args) != 2 {
			return fmt.Errorf("Expected 2 arguments: verb and path, see help")
		}

		verb := strings.ToUpper(args[0])
		path := args[1]

		fields, err := cmd.Flags().GetStringSlice("field")
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

		response, err := client.API(verb, path, mappedFields, body)
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
	apiCmd.Flags().StringSliceVarP(&fields, "field", "f", nil, "Fields to pass to the api call")
	apiCmd.Flags().StringVarP(&body, "body", "b", "{}", "Body to send as JSON")
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

const apiCmdUsage string = `TODO`
