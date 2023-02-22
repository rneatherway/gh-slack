package cmd

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"regexp"

	"github.com/rneatherway/gh-slack/internal/gh"
	"github.com/rneatherway/gh-slack/internal/markdown"
	"github.com/rneatherway/gh-slack/internal/slackclient"
	"github.com/rneatherway/gh-slack/internal/version"
	"github.com/spf13/cobra"
)

var readCmd = &cobra.Command{
	Use:   "read [flags] <START>",
	Short: "Reads a Slack channel and outputs the messages as markdown",
	Long:  `Reads a Slack channel and outputs the messages as markdown for GitHub issues.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return readSlack(args)
	},
	Example: `  gh-slack read <slack-permalink>
  gh-slack read -i <issue-url> <slack-permalink>`,
}

var (
	permalinkRE = regexp.MustCompile("https://([^./]+).slack.com/archives/([A-Z0-9]+)/p([0-9]+)([0-9]{6})")
	nwoRE       = regexp.MustCompile("^/[^/]+/[^/]+/?$")
	issueRE     = regexp.MustCompile("^/[^/]+/[^/]+/issues/[0-9]+/?$")
)

type linkParts struct {
	team      string
	channelID string
	timestamp string
}

// https://github.slack.com/archives/CP9GMKJCE/p1648028606962719
// returns (github, CP9GMKJCE, 1648028606.962719, nil)
func parsePermalink(link string) (linkParts, error) {
	result := permalinkRE.FindStringSubmatch(link)
	if result == nil {
		return linkParts{}, fmt.Errorf("not a permalink: %q", link)
	}

	return linkParts{
		team:      result[1],
		channelID: result[2],
		timestamp: result[3] + "." + result[4],
	}, nil
}

var opts struct {
	Args struct {
		Start string
	}
	Limit   int
	Version bool
	Details bool
	Issue   string
}

func init() {
	readCmd.Flags().IntVarP(&opts.Limit, "limit", "l", 20, "Number of _channel_ messages to be fetched after the starting message (all thread messages are fetched)")
	readCmd.Flags().BoolVar(&opts.Version, "version", false, "Output version information")
	readCmd.Flags().BoolVarP(&opts.Details, "details", "d", false, "Wrap the markdown output in HTML <details> tags")
	readCmd.Flags().StringVarP(&opts.Issue, "issue", "i", "", "The URL of a repository to post the output as a new issue, or the URL of an issue to add a comment to that issue")
	readCmd.SetHelpTemplate(readCmdUsage)
	readCmd.SetUsageTemplate(readCmdUsage)
}

func readSlack(args []string) error {
	if opts.Version {
		fmt.Printf("gh-slack %s (%s)\n", version.Version(), version.Commit())
		return nil
	}

	if len(args) == 0 {
		return errors.New("the required argument <START> was not provided")
	}
	opts.Args.Start = args[0]
	if opts.Args.Start == "" {
		return errors.New("the required argument <START> was not provided")
	}

	var repoUrl, issueUrl string
	if opts.Issue != "" {
		u, err := url.Parse(opts.Issue)
		if err != nil {
			return err
		}

		if nwoRE.MatchString(u.Path) {
			repoUrl = opts.Issue
		} else if issueRE.MatchString(u.Path) {
			issueUrl = opts.Issue
		} else {
			return fmt.Errorf("not a repository or issue URL: %q", opts.Issue)
		}
	}

	linkParts, err := parsePermalink(opts.Args.Start)
	if err != nil {
		return err
	}

	logger := log.New(io.Discard, "", log.LstdFlags)
	if verbose {
		logger = log.Default()
	}

	client, err := slackclient.New(
		linkParts.team,
		logger)
	if err != nil {
		return err
	}

	history, err := client.History(linkParts.channelID, linkParts.timestamp, opts.Limit)
	if err != nil {
		return err
	}

	output, err := markdown.FromMessages(client, history)
	if err != nil {
		return err
	}

	var channelName string
	if opts.Details {
		channelInfo, err := client.ChannelInfo(linkParts.channelID)
		if err != nil {
			return err
		}

		channelName = channelInfo.Name
		output = markdown.WrapInDetails(channelName, opts.Args.Start, output)
	}

	if repoUrl != "" {
		if channelName == "" {
			channelInfo, err := client.ChannelInfo(linkParts.channelID)
			if err != nil {
				return err
			}
			channelName = channelInfo.Name
		}

		err := gh.NewIssue(repoUrl, channelName, output)
		if err != nil {
			return err
		}
	} else if issueUrl != "" {
		err := gh.AddComment(issueUrl, output)
		if err != nil {
			return err
		}
	} else {
		os.Stdout.WriteString(output)
	}

	return nil
}

const readCmdUsage string = `Usage:{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command] <START>{{end}}

  where <START> is a required argument which should be permalink for the first message to fetch. Following messages are then fetched from that channel (or thread if applicable).{{if gt (len .Aliases) 0}}
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
