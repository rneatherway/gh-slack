# gh-slack

[![CI](https://github.com/rneatherway/gh-slack/actions/workflows/ci.yml/badge.svg)](https://github.com/rneatherway/gh-slack/actions/workflows/ci.yml) [![Release](https://github.com/rneatherway/gh-slack/actions/workflows/release.yml/badge.svg)](https://github.com/rneatherway/gh-slack/actions/workflows/release.yml)

This project provides a means of archiving a Slack conversation or thread as markdown. For convenience it is installable as a `gh` extension.

## Installation

    gh extension install https://github.com/rneatherway/gh-slack

## Upgrading

    gh extension upgrade gh-slack

## Usage

    Usage:
      gh slack [OPTIONS] [Start]

    Application Options:
      -l, --limit=   Number of _channel_ messages to be fetched after the starting message (all thread messages are fetched) (default: 20)
      -v, --verbose  Show verbose debug information
          --version  Output version information
      -d, --details  Wrap the markdown output in HTML <details> tags
      -i, --issue=   The URL of a repository to post the output as a new issue, or the URL of an issue to add a comment to that issue

    Help Options:
      -h, --help     Show this help message

    Arguments:
      Start:         Required. Permalink for the first message to fetch. Following messages are then fetched from that channel (or thread if applicable)

## Limitations

Many and varied, but at least:

* No paging is used when fetching messages, so if the conversation is too long the output will be truncated.

## Development

To release a new version, simply tag it. The `goreleaser` workflow will take care of the rest. E.g:

    git tag 0.0.6
    git push origin 0.0.6
