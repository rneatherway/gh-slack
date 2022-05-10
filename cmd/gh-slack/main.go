package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"regexp"

	"github.com/jessevdk/go-flags"
	"github.com/rneatherway/gh-slack/internal/gh"
	"github.com/rneatherway/gh-slack/internal/markdown"
	"github.com/rneatherway/gh-slack/internal/slackclient"
	"github.com/rneatherway/gh-slack/internal/version"
	"r00t2.io/gosecret"
)

var (
	permalinkRE = regexp.MustCompile("https://[^./]+.slack.com/archives/([A-Z0-9]+)/p([0-9]+)([0-9]{6})")
	nwoRE       = regexp.MustCompile("^/[^/]+/[^/]+/?$")
	issueRE     = regexp.MustCompile("^/[^/]+/[^/]+/issues/[0-9]+/?$")
)

// https://github.slack.com/archives/CP9GMKJCE/p1648028606962719
// returns (CP9GMKJCE, 1648028606.962719, nil)
func parsePermalink(link string) (string, string, error) {
	result := permalinkRE.FindStringSubmatch(link)
	if result == nil {
		return "", "", fmt.Errorf("not a permalink: %q", link)
	}

	return result[1], result[2] + "." + result[3], nil
}

var opts struct {
	Args struct {
		Start string `description:"Required. Permalink for the first message to fetch. Following messages are then fetched from that channel (or thread if applicable)"`
	} `positional-args:"yes"`
	Limit   int    `short:"l" long:"limit" default:"20" description:"Number of _channel_ messages to be fetched after the starting message (all thread messages are fetched)"`
	Verbose bool   `short:"v" long:"verbose" description:"Show verbose debug information"`
	Version bool   `long:"version" description:"Output version information"`
	Details bool   `short:"d" long:"details" description:"Wrap the markdown output in HTML <details> tags"`
	Issue   string `short:"i" long:"issue" description:"The URL of a repository to post the output as a new issue, or the URL of an issue to add a comment to that issue"`
}

const (
	// The default collection; it should be available on all SecretService implementations.
	collectionName string = "login"
	// A label for an Item used in examples below.
	exampleLabel string = "Some Website Credentials"
)

func test() error {

	var err error
	var service *gosecret.Service
	//var collection *gosecret.Collection
	// var item *gosecret.Item
	// var itemAttrs map[string]string
	var itemLabel string
	//var secret *gosecret.Secret

	// All interactions with SecretService start with initiating a Service connection.
	if service, err = gosecret.NewService(); err != nil {
		log.Panicln(err)
	}
	defer service.Close()

	// And unless operating directly on a Service via its methods, you probably need a Collection as well.
	// if collection, err = service.GetCollection(collectionName); err != nil {
	// 	log.Panicln(err)
	// }

	/*
		Create a Secret which gets stored in an Item which gets stored in a Collection.
		See the documentation for details.
	*/
	// Incidentally, I believe this is the only exported function/method that does not return an error returner.
	// secret = gosecret.NewSecret(
	// 	service.Session,                   // The session associated with this Secret. You're likely fine with the automatically-created *(Service).Session.
	// 	[]byte{},                          // The "parameters". Likely this is an empty byteslice.
	// 	[]byte("a super secret password"), // The actual secret value.
	// 	"text/plain",                      // The content type (MIME type/media type). See https://www.iana.org/assignments/media-types/media-types.xhtml.
	// )

	/*
		Item attributes are a map[string]string of *metadata* about a Secret/Item.
		Do *NOT* store sensitive information in these.
		They're primarily used for searching for Items.
	*/
	itemAttrs := map[string]string{
		"xdg:schema":  "chrome_libsecret_os_crypt_password_v2",
		"application": "Slack",
	}

	// // And create the Item (and add it to SecretService).
	// if item, err = collection.CreateItem(
	// 	exampleLabel, // The label of the item. This should also be considered not secret.
	// 	itemAttrs,    // Attributes for the item; see above.
	// 	secret,       // The actual secret.
	// 	true,         // Whether to replace an existing item with the same label or not.
	// ); err != nil {
	// 	log.Panicln(err)
	// }

	/*
		Now let's fetch the same Item via its attributes.
		The results are split into locked items and unlocked items.
	*/
	var unlockedItems []*gosecret.Item
	//var lockedItems []*gosecret.Item

	if unlockedItems, _, err = service.SearchItems(itemAttrs); err != nil {
		log.Panicln(err)
	}

	// We should only have one Item that matches the search attributes, and unless the item or collection is locked, ...
	item := unlockedItems[0]
	if itemLabel, err = item.Label(); err != nil {
		log.Panicln(err)
	}
	fmt.Printf("Found item: %v %v\n", itemLabel, string(item.Secret.Value))

	// Alternatively if you are unsure of the attributes but know the label of the item you want, you can iterate through them.
	// var itemResults []*gosecret.Item

	// if itemResults, err = collection.Items(); err != nil {
	// 	log.Panicln(err)
	// }
	// fmt.Printf("Found %d items\n", len(itemResults))
	// for idx, i := range itemResults {
	// 	if itemLabel, err = i.Label(); err != nil {
	// 		fmt.Printf("Cannot read label for item at path '%v'\n", i.Dbus.Path())
	// 		continue
	// 	}
	// 	// if itemLabel != exampleLabel { // Matching against a desired label - exampleLabel, in this case.
	// 	// 	continue
	// 	// }
	// 	fmt.Printf("Found item labeled '%v'! Index number %v at path '%v'\n", itemLabel, idx, i.Dbus.Path())
	// 	fmt.Println(i.Attrs)
	// 	//fmt.Printf("Password: %v\n", string(i.Secret.Value))
	// 	//break
	// }
	return nil
}

func realMain() error {
	test()
	return nil
	_, err := flags.NewParser(&opts, flags.HelpFlag|flags.PassDoubleDash).Parse()
	if err != nil {
		return err
	}

	if opts.Version {
		fmt.Printf("gh-slack %s (%s)\n", version.Version(), version.Commit())
		return nil
	}

	if opts.Args.Start == "" {
		return errors.New("the required argument `Start` was not provided")
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

	channelID, timestamp, err := parsePermalink(opts.Args.Start)
	if err != nil {
		return err
	}

	logger := log.New(io.Discard, "", log.LstdFlags)
	if opts.Verbose {
		logger = log.Default()
	}

	client, err := slackclient.New(
		"github", // This could be made configurable at some point
		logger)
	if err != nil {
		return err
	}

	history, err := client.History(channelID, timestamp, opts.Limit)
	if err != nil {
		return err
	}

	output, err := markdown.FromMessages(client, history)
	if err != nil {
		return err
	}

	if opts.Details {
		output = markdown.WrapInDetails(output)
	}

	if repoUrl != "" {
		channelInfo, err := client.ChannelInfo(channelID)
		if err != nil {
			return err
		}
		gh.NewIssue(repoUrl, channelInfo, output)
	} else if issueUrl != "" {
		channelInfo, err := client.ChannelInfo(channelID)
		if err != nil {
			return err
		}
		gh.AddComment(issueUrl, channelInfo, output)
	} else {
		os.Stdout.WriteString(output)
	}

	return nil
}

func main() {
	err := realMain()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
