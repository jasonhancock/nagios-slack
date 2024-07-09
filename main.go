package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/jasonhancock/slack-go-webhook"
)

const defaultIcon = ":white_medium_square:"

var icons = map[string]string{
	"CRITICAL":        ":exclamation:",
	"DOWN":            ":exclamation:",
	"WARNING":         ":warning:",
	"OK":              ":white_check_mark:",
	"UP":              ":white_check_mark:",
	"UNKNOWN":         ":question:",
	"ACKNOWLEDGEMENT": ":heart:",
}

func main() {
	if len(os.Args) != 5 {
		fmt.Printf("%s <nagios_url> <slack webhook url> <slack channel> <slack bot name>\n", os.Args[0])
		os.Exit(1)
	}

	var (
		nagiosURL       = os.Args[1]
		slackWebhookURL = os.Args[2]
		slackChannel    = os.Args[3]
		slackBotName    = os.Args[4]

		icon      = defaultIcon
		msg       = ""
		msgPrefix = ""
	)
	notificationType := "host"
	if os.Getenv("NAGIOS_SERVICEATTEMPT") != "" {
		notificationType = "service"
	}

	if notificationType == "service" {
		if i, ok := icons[os.Getenv("NAGIOS_SERVICESTATE")]; ok {
			icon = i
		}
		msg = fmt.Sprintf("MESSAGE: %s", os.Getenv("NAGIOS_SERVICEOUTPUT"))
		msgPrefix = fmt.Sprintf("Service: %s", os.Getenv("NAGIOS_SERVICEDISPLAYNAME"))
	} else {
		if i, ok := icons[os.Getenv("NAGIOS_HOSTSTATE")]; ok {
			icon = i
		}
		msgPrefix = fmt.Sprintf("is %s", os.Getenv("NAGIOS_HOSTSTATE"))
	}

	// override if it's an Ack
	if os.Getenv("NAGIOS_NOTIFICATIONTYPE") == "ACKNOWLEDGEMENT" {
		icon = icons["ACKNOWLEDGEMENT"]
		msg = fmt.Sprintf(
			"ACKNOWLEDGED BY: %s  COMMENT: %s",
			os.Getenv("NAGIOS_NOTIFICATIONAUTHOR"),
			os.Getenv("NAGIOS_NOTIFICATIONCOMMENT"),
		)
	}

	url := fmt.Sprintf(
		"%s/cgi-bin/status.cgi?navbarsearch=1&host=%s",
		strings.TrimSuffix(nagiosURL, "/"),
		os.Getenv("NAGIOS_HOSTNAME"),
	)

	payload := slack.Payload{
		Channel:  slackChannel,
		Username: slackBotName,
		Text: fmt.Sprintf(
			"%s HOST: %s  %s  %s  <%s|See Nagios>",
			icon,
			os.Getenv("NAGIOS_HOSTNAME"),
			msgPrefix,
			msg,
			url,
		),
	}

	// Slack package returns a slice of errors. Convert into a multierror
	if err := slack.Send(slackWebhookURL, "", payload); err != nil {
		fmt.Fprintf(os.Stderr, "sending slack notification: %s\n", err.Error())
		os.Exit(1)
	}
}
