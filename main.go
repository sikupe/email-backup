package main

import (
	"fmt"
	"github.com/urfave/cli/v2"
	"log"
	"math"
	"os"
	"sort"
)

func formatBytes(bytes uint64) string {
	step := uint(math.Log10(float64(bytes))) / 3
	mapping := map[uint]string{
		0: " B",
		1: " kB",
		2: " MB",
		3: " GB",
		4: "TB",
	}
	divisor := math.Pow(1024, float64(step))
	adjustedBytes := float64(bytes) / divisor

	return fmt.Sprintf("%.1f %s", adjustedBytes, mapping[step])
}

func main() {
	app := &cli.App{
		Name: "Email backup",
		Flags: []cli.Flag{
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
		Commands: []*cli.Command{
			&cli.Command{
				Name: "backup",

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
			},
			{
				Name: "sizes",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "path",
						Value: "",
						Usage: "Path to measure size",
					},
					&cli.BoolFlag{
						Name: "sort-by-size",
					},
				},
				Action: func(context *cli.Context) error {
					server := context.String("server")
					user := context.String("user")
					password := context.String("password")
					path := context.String("path")
					sortBySize := context.Bool("sort-by-size")

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
					sizes := d.Sizes(paths)

					total := uint64(0)

					keys := make([]string, 0, len(sizes))

					for key := range sizes {
						keys = append(keys, key)
					}

					if sortBySize {
						sort.SliceStable(keys, func(i, j int) bool {
							return sizes[keys[i]] > sizes[keys[j]]
						})
					}

					for _, folder := range keys {
						bytes := sizes[folder]
						fmt.Printf("%s: %s\n", folder, formatBytes(bytes))
						total += bytes
					}
					fmt.Printf("Total space used: %s\n", formatBytes(total))

					log.Println("Done!")

					return nil

				},
			},
		},
	}
	app.Run(os.Args)
}
