# gh-slack

[![CI](https://github.com/rneatherway/gh-slack/actions/workflows/ci.yml/badge.svg)](https://github.com/rneatherway/gh-slack/actions/workflows/ci.yml) [![Release](https://github.com/rneatherway/gh-slack/actions/workflows/release.yml/badge.svg)](https://github.com/rneatherway/gh-slack/actions/workflows/release.yml)

This project provides a means of archiving a Slack conversation or thread as markdown. For convenience it is installable as a [`gh`](https://github.com/cli/cli) [extension]([url](https://cli.github.com/manual/gh_extension)).

## Installation

    gh extension install https://github.com/rneatherway/gh-slack

## Upgrading

    gh extension upgrade gh-slack

## Usage

    Usage:
      gh slack [command]

      If no command is specified, the default is "read". The default command also requires a permalink argument <START> for the first message to fetch.
      Use "gh slack read --help" for more information about the default command behaviour.

    Examples:
      gh slack -i <issue-url> <slack-permalink>  # defaults to read command
      gh slack read <slack-permalink>
      gh slack read -i <issue-url> <slack-permalink>
      gh slack send -m <message> -c <channel-name> -t <team-name>

      # Example configuration (add to gh's configuration file at $HOME/.config/gh/config.yml):
      extensions:
        slack:
          team: github
          channel: ops
          bot: hubot        # Can be a user id (most reliable), bot profile name or username

    Available Commands:
      completion  Generate the autocompletion script for the specified shell
      help        Help about any command
      read        Reads a Slack channel and outputs the messages as markdown
      send        Sends a message to a Slack channel

    Flags:
      -h, --help      help for gh-slack
      -v, --verbose   Show verbose debug information

    Use "gh slack [command] --help" for more information about a command.

## Configuration

The `send` subcommand supports storing default values for the `team`, `bot` and
`channel` required parameters in gh's own configuration file using a block like:

```yaml
extensions:
  slack:
    team: foo
    channel: ops
    bot: robot        # Can be a user id (most reliable), bot profile name or username
```

This is particularly useful if you want to use the `send` subcommand to interact
with a bot serving chatops in a standard operations channel.

## Limitations

Many and varied, but at least:

* No paging is used when fetching messages, so if the conversation is too long the output will be truncated.

## Development

To release a new version, simply tag it. The `goreleaser` workflow will take care of the rest. E.g:

    git tag 0.0.6
    git push origin 0.0.6
