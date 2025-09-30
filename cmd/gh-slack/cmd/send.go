package cmd

import (
	"fmt"
	"io"
	"log"

	"github.com/cli/go-gh/v2/pkg/config"
	"github.com/rneatherway/gh-slack/internal/slackclient"
	"github.com/spf13/cobra"
)

var sendCmd = &cobra.Command{
	Use:   "send [flags]",
	Short: "Sends a message to a Slack channel",
	Long:  `Sends a message to a Slack channel.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Read(nil)
		if err != nil {
			return err
		}

		channelName, err := getFlagOrElseConfig(cfg, cmd.Flags(), "channel")
		if err != nil {
			return err
		}

		team, err := getFlagOrElseConfig(cfg, cmd.Flags(), "team")
		if err != nil {
			return err
		}

		message, err := cmd.Flags().GetString("message")
		if err != nil {
			return err
		}

		wait, err := cmd.Flags().GetBool("wait")
		if err != nil {
			return err
		}

		bot, err := cmd.Flags().GetString("bot")
		if err != nil {
			return err
		}

		if wait && bot == "" {
			bot, err = getGHSlackConfigValue(cfg, "bot")
			if err != nil {
				return err
			}
		}

		logger := log.New(io.Discard, "", log.LstdFlags)
		if verbose {
			logger = log.Default()
		}
		return sendMessage(team, channelName, message, bot, logger)
	},
	Example: `  gh-slack send -t <team-name> -c <channel-name> -m <message> -b <bot-name>
  gh-slack send -m <message> -w # If bot is specified in config
` + sendConfigEample,
}

// sendMessage sends a message to a Slack channel.
func sendMessage(team, channelName, message, bot string, logger *log.Logger) error {
	client, err := slackclient.New(team, logger)
	if err != nil {
		return err
	}

	var rtmClient *slackclient.RTMClient
	if bot != "" {
		rtmClient, err = client.ConnectToRTM()
		if err != nil {
			return err
		}
		defer rtmClient.Close()
	}

	channelID, err := client.ChannelIDForName(channelName)
	if err != nil {
		return err
	}

	// We get back the permalink to the message we just sent, but I don't
	// currently see a use for that.
	_, err = client.SendMessage(channelID, message)
	if err != nil {
		return err
	}

	if bot != "" {
		err = rtmClient.ListenForMessagesFromBot(channelID, bot)
		if err != nil {
			return fmt.Errorf("failed to listen to messages: %w", err)
		}
	}

	return nil
}

func init() {
	sendCmd.Flags().StringP("channel", "c", "", "Channel name to send the message to (required here or in config)")
	sendCmd.Flags().StringP("message", "m", "", "Message to send (required here or in config)")
	sendCmd.Flags().StringP("team", "t", "", "Slack team name (required here or in config)")
	sendCmd.MarkFlagRequired("message")
	sendCmd.Flags().StringP("bot", "b", "", "User id (most reliable), profile name or username to wait for a response from (implies --wait)")
	sendCmd.Flags().BoolP("wait", "w", false, "Wait for message responses")
	sendCmd.MarkFlagsRequiredTogether("message")
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
