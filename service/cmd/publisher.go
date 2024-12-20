/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/tjololo/nats-demo/service/internal/publisher"
)

// publisherCmd represents the publisher command
var publisherCmd = &cobra.Command{
	Use:   "publisher",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		port, err := cmd.Flags().GetInt16("port")
		if err != nil {
			fmt.Printf("Failed to read flag port. %s", err)
			os.Exit(1)
		}
		every, err := cmd.Flags().GetDuration("every")
		if err != nil {
			fmt.Printf("Failed to read flag every. %s", err)
			os.Exit(1)
		}
		subject, err := cmd.Flags().GetString("subject")
		if err != nil {
			fmt.Printf("Failed to read flag subject. %s", err)
			os.Exit(1)
		}
		reply, err := cmd.Flags().GetString("reply")
		if err != nil {
			fmt.Printf("Failed to read flag reply. %s", err)
			os.Exit(1)
		}
		nats, err := cmd.Flags().GetString("nats")
		if err != nil {
			fmt.Printf("Failed to read flag nats. %s", err)
			os.Exit(1)
		}
		pc := publisher.PublisherConfig{
			NatsServerURL: nats,
			DefaultSubject: subject,
			Port: port,
			Every: every,
			ReplySubject: reply,
		}
		publisher.Publisher(pc)
	},
}

func init() {
	serveCmd.AddCommand(publisherCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// publisherCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// publisherCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	publisherCmd.Flags().Duration("every", time.Duration(0), "Publish every X. 0 Means no messages are sendt automatically")
	publisherCmd.Flags().Int16P("port", "p", 8181, "Port the api should be exposed. Used to publish custom messages")
	publisherCmd.Flags().StringP("subject", "s", "foo", "Default subject to publish to")
	publisherCmd.Flags().StringP("reply", "r", "bar", "Default subject to reply to")
	publisherCmd.Flags().StringP("nats", "n", "nats://localhost:4222", "Nats server to connect to")
}
