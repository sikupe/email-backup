package main

import (
	"bytes"
	"fmt"
	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"log"
	"net/url"
	"os"
	"path/filepath"
)

type Downloader struct {
	client *client.Client
}

func generateFileName(m *imap.Message) string {
	date := m.InternalDate.String()

	sender := m.Envelope.From[0].MailboxName + "@" + m.Envelope.From[0].HostName
	subject := url.PathEscape(m.Envelope.Subject)
	messageId := m.Uid

	return fmt.Sprintf("%s－%s－%s－%d.eml", date, sender, subject, messageId)
}

func (d Downloader) Download(paths []string, output string) error {
	for _, p := range paths {
		outputDir := filepath.Join(output, p)
		if err := os.MkdirAll(outputDir, 0777); err != nil {
			return err
		}
		mbox, err := d.client.Select(p, false)
		if err != nil {
			log.Fatal(err)
		}

		seqset := new(imap.SeqSet)
		seqset.AddRange(uint32(1), mbox.Messages)

		messages := make(chan *imap.Message, 10)
		done := make(chan error, 1)
		go func() {
			done <- d.client.Fetch(seqset, []imap.FetchItem{imap.FetchEnvelope, imap.FetchInternalDate, imap.FetchUid, "BODY.PEEK[]"}, messages)
		}()

		for m := range messages {
			totalBuf := &bytes.Buffer{}
			for _, value := range m.Body {
				length := value.Len()
				buf := make([]byte, length)
				n, err := value.Read(buf)
				if err != nil {
					log.Fatal(err)
				}
				if n != length {
					log.Fatal("Didn't read correct length")
				}

				totalBuf.Write(buf)
			}
			filename := generateFileName(m)
			outputFilePath := filepath.Join(outputDir, filename)
			if file, err := os.Create(outputFilePath); err != nil {
				log.Fatal(err)
			} else {
				if _, err := file.Write(totalBuf.Bytes()); err != nil {
					log.Fatal(err)
				}
			}
			log.Printf("Downloaded %s\n", outputFilePath)
		}

		if err := <-done; err != nil {
			log.Fatal(err)
		}
	}
	return nil
}

func (d Downloader) ListFolders(path string) []string {
	name := ""
	if path == "" {
		name = "*"
	}

	mailboxes := make(chan *imap.MailboxInfo, 10)
	done := make(chan error, 1)

	go func() {
		done <- d.client.List(path, name, mailboxes)
	}()

	paths := []string{}

	log.Println("Mailboxes:")
	for m := range mailboxes {
		paths = append(paths, m.Name)
	}

	if err := <-done; err != nil {
		log.Fatal(err)
	}

	return paths
}

func (d Downloader) Logout() {
	d.client.Logout()
}

func CreateDownloader(server, user, password string) *Downloader {
	log.Println("Connecting to server...")
	// Connect to server
	c, err := client.DialTLS(server, nil)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Connected")

	// Login
	if err := c.Login(user, password); err != nil {
		log.Fatal(err)
	}
	log.Println("Logged in")

	return &Downloader{
		client: c,
	}
}
