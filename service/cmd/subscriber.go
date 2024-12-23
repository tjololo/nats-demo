/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"log"

	"github.com/nats-io/nats.go"
	"github.com/spf13/cobra"
	"github.com/tjololo/nats-demo/service/internal/subscriber"
)

// subscriberCmd represents the subscriber command
var subscriberCmd = &cobra.Command{
	Use:   "subscriber",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		sub, err := cmd.Flags().GetString("subscription")
		if err != nil {
			log.Fatalf("Failed to read flag subscription. %s", err)
		}
		natsURL, err := cmd.Flags().GetString("nats")
		if err != nil {
			log.Fatalf("Failed to read flag nats. %s", err)
		}
		//subscriber.StartSubscriber()
		config := subscriber.ClientConfig{
			NatsServerURL: natsURL,
			NatsSubscription: sub,
		}
		subscriber.Subscriber(config)
	},
}

func init() {
	serveCmd.AddCommand(subscriberCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// subscriberCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// subscriberCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	subscriberCmd.Flags().StringP("subscription", "s", "foo", "Subject to subscribe to")
	subscriberCmd.Flags().StringP("nats", "n", nats.DefaultURL, "Nats server url")
}
