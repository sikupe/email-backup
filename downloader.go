package main

import (
	"bytes"
	"fmt"
	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"log"
	"math"
	"net/url"
	"os"
	"path/filepath"
)

const maxFileNameLength = 150

type Downloader struct {
	client *client.Client
}

func generateFileName(m *imap.Message) string {
	date := m.InternalDate.String()

	sender := "Unknown"
	if len(m.Envelope.From) > 0 {
		sender = m.Envelope.From[0].MailboxName + "@" + m.Envelope.From[0].HostName
	}
	subject := url.PathEscape(m.Envelope.Subject)
	messageId := m.Uid

	fileNameLength := len(date) + len(sender) + int(math.Floor(math.Log10(float64(messageId)))) + 1 + len(subject) + 3*3 + 4

	trimLength := fileNameLength - maxFileNameLength
	trimmedLength := len(subject) - trimLength

	if fileNameLength > maxFileNameLength {
		subject = subject[:trimmedLength]
	}

	return fmt.Sprintf("%s－%s－%s－%d.eml", date, sender, subject, messageId)
}

func (d Downloader) Sizes(paths []string) map[string]uint64 {
	result := make(map[string]uint64)
	for _, p := range paths {
		mbox, err := d.client.Select(p, false)
		if err != nil {
			log.Fatal(err)
		}

		seqset := new(imap.SeqSet)
		seqset.AddRange(uint32(1), mbox.Messages)

		messages := make(chan *imap.Message, 10)
		done := make(chan error, 1)
		go func() {
			done <- d.client.Fetch(seqset, []imap.FetchItem{imap.FetchFast}, messages)
		}()

		folderSize := uint64(0)
		for m := range messages {
			folderSize += uint64(m.Size)
		}

		if err := <-done; err != nil {
			log.Fatal(err)
		}
		result[p] = folderSize
	}
	return result
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
	//if err := d.client.Logout(); err != nil {
	//	log.Fatal(err)
	//}
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
