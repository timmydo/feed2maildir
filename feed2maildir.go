package main

import (
	"context"
	"flag"
	"fmt"
	"mime"
	"os"
	"path"
	"time"

	"github.com/cespare/xxhash"
	"github.com/mmcdole/gofeed"
)

var (
	// Command line flags
	flagFeed    = flag.String("feed", "", "Feed url")
	flagMaildir = flag.String("maildir", "", "Maildir path")
	flagVersion = flag.Bool("version", false, "Show the version number and information")
	optionFlags map[string]*string
	Version     string
)

func printHelp() {
	fmt.Println("Help")
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func dirExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}

func mimeEncode(txt string) string {
	return mime.QEncoding.Encode("utf-8", txt)
}

func main() {
	flag.Parse()
	if *flagVersion {
		fmt.Println("feed2maildir " + Version)
		return
	}

	if *flagFeed == "" {
		fmt.Println("--feed required")
		os.Exit(1)
	}

	if *flagMaildir == "" {
		fmt.Println("--maildir required")
		os.Exit(1)
	}

	if !dirExists(*flagMaildir) {
		fmt.Println("Maildir does not exist: ", *flagMaildir)
		os.Exit(1)
	}

	curDir := path.Join(*flagMaildir, "cur")
	newDir := path.Join(*flagMaildir, "new")
	if !dirExists(curDir) {
		fmt.Println("Maildir/cur directory does not exist: ", *flagMaildir)
	}
	if !dirExists(newDir) {
		fmt.Println("Maildir/new directory does not exist: ", *flagMaildir)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	fp := gofeed.NewParser()
	// "http://feeds.twit.tv/twit.xml"
	feed, _ := fp.ParseURLWithContext(*flagFeed, ctx)

	h := xxhash.New()

	for _, item := range feed.Items {
		h.Reset()
		h.Write([]byte(item.Title))
		h.Write([]byte(item.Link))
		sum := h.Sum64()
		mailFilename := fmt.Sprintf("feed2maildir.%d", sum)
		fullCurFilename := path.Join(curDir, mailFilename)
		fullNewFilename := path.Join(newDir, mailFilename)
		if fileExists(fullCurFilename) || fileExists(fullNewFilename) {
			// email already written
			continue
		}

		f, err := os.Create(fullNewFilename)
		if err != nil {
			fmt.Printf("Error writing %s: %s", fullNewFilename, err)
		}

		defer f.Close()

		fmt.Fprintf(f, "Content-Transfer-Encoding: quoted-printable\n")
		timeNowString := time.Now().UTC().Format(time.RFC1123Z)
		fmt.Fprintf(f, "Date: %s\n", timeNowString)
		fmt.Fprintf(f, "Content-Type: text/plain; charset=UTF-8\n")
		fmt.Fprintf(f, "Message-Id: <%s@local>\n", mailFilename)
		fmt.Fprintf(f, "From: (%s) <feed@local>\n", mimeEncode(feed.Title))
		fmt.Fprintf(f, "Subject: %s\n", mimeEncode(item.Title))
		fmt.Fprintf(f, "\n%s\n\n%s\n%s", item.Link, item.Description, item.Content)

		fmt.Printf("%s: %s -> %s\n", feed.Title, item.Title, fullNewFilename)
	}

}
