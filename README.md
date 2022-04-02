# slack-to-md

This project provides a means of archiving a Slack conversation or thread as markdown.

## Installation

    go install github.com/rneatherway/slack-to-md/cmd/slack-to-md@latest

## Usage

```
Usage:
  main [OPTIONS] [Start]

Application Options:
  -l, --limit=   Number of _channel_ messages to be fetched after the starting message (all thread messages are fetched) (default: 20)
  -v, --verbose  Show verbose debug information

Help Options:
  -h, --help     Show this help message

Arguments:
  Start:         Permalink for the first message to fetch. Following messages are then fetched from that channel(or thread if applicable)
```

## Limitations

Many and varied, but at least:

* No paging is used when fetching messages, so if the conversation is too long the output will be truncated.