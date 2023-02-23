package cmd

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/rneatherway/gh-slack/internal/slackclient"
	"github.com/spf13/cobra"
)

var sendCmd = &cobra.Command{
	Use:   "send [flags]",
	Short: "Sends a message to a Slack channel",
	Long:  `Sends a message to a Slack channel.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		channelID, err := cmd.Flags().GetString("channel")
		if err != nil {
			return err
		}
		message, err := cmd.Flags().GetString("message")
		if err != nil {
			return err
		}
		team, err := cmd.Flags().GetString("team")
		if err != nil {
			return err
		}
		logger := log.New(io.Discard, "", log.LstdFlags)
		if verbose {
			logger = log.Default()
		}
		bot, err := cmd.Flags().GetString("bot")
		if err != nil {
			return err
		}
		return sendMessage(team, channelID, message, bot, logger)
	},
	Example: `  gh-slack send -t <team-name> -c <channel-id> -m <message>`,
}

// sendMessage sends a message to a Slack channel.
func sendMessage(team, channelID, message, bot string, logger *log.Logger) error {
	client, err := slackclient.New(team, logger)
	if err != nil {
		return err
	}
	defer client.Close()
	resp, err := client.SendMessage(channelID, message)
	if err != nil {
		return err
	}
	// only listen to messages when bot is specified
	// TODO: maybe we should move this to a separate function (SoC)
	if bot != "" {
		err = client.ListenForMessagesFromBot(channelID, bot)
		if err != nil {
			return fmt.Errorf("failed to listen to messages: %w", err)
		}
	}
	fmt.Fprintln(os.Stdout, resp.Output(team, channelID))
	return nil
}

func init() {
	sendCmd.Flags().StringP("channel", "c", "", "Channel ID to send the message to (required)")
	sendCmd.Flags().StringP("message", "m", "", "Message to send (required)")
	sendCmd.Flags().StringP("team", "t", "", "Slack team name (required)")
	sendCmd.MarkFlagRequired("channel")
	sendCmd.MarkFlagRequired("message")
	sendCmd.MarkFlagRequired("team")
	sendCmd.Flags().StringP("bot", "b", "", "Name of the bot to listen to for message responses")
	sendCmd.MarkFlagsRequiredTogether("channel", "message", "team")
	sendCmd.SetUsageTemplate(sendCmdUsage)
	sendCmd.SetHelpTemplate(sendCmdUsage)
}

const sendCmdUsage string = `Usage:{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}}{{end}}{{if gt (len .Aliases) 0}}
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
