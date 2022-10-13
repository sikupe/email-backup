package main

import (
	"github.com/urfave/cli/v2"
	"log"
	"os"
)

func main() {
	app := &cli.App{
		Name: "Email backup",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "path",
				Value: "",
				Usage: "Path to backup",
			},
			&cli.StringFlag{
				Name:  "output",
				Value: ".",
				Usage: "Download path",
			},
			&cli.StringFlag{
				Name:    "server",
				EnvVars: []string{"IMAP_SERVER"},
				Usage:   "Mail server hostname and port",
			},
			&cli.StringFlag{
				Name:    "user",
				EnvVars: []string{"IMAP_USER"},
				Usage:   "User for the imap server",
			},
			&cli.StringFlag{
				Name:    "password",
				EnvVars: []string{"IMAP_PASSWORD"},
				Usage:   "Password for the imap server",
			},
		},
		Action: func(context *cli.Context) error {
			server := context.String("server")
			user := context.String("user")
			password := context.String("password")
			path := context.String("path")
			output := context.String("output")

			paths := []string{}

			if path != "" {
				paths = append(paths, path)
			}

			d := CreateDownloader(server, user, password)

			// Don't forget to logout
			defer d.Logout()

			// List mailboxes
			listedPaths := d.ListFolders(path)
			paths = append(paths, listedPaths...)

			// Dowloading messages
			if err := d.Download(paths, output); err != nil {
				log.Fatal(err)
			}
			log.Println("Done!")

			return nil
		},
	}
	app.Run(os.Args)
}
