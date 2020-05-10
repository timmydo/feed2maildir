package main

import (
	"context"
	"encoding/base64"
	"flag"
	"fmt"
	"mime"
	"os"
	"path"
	"strings"
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

func filesInDirToHashSet(dirname string) (map[string]bool, error) {
	f, err := os.Open(dirname)
	if err != nil {
		return nil, err
	}
	list, err := f.Readdir(-1)
	f.Close()
	if err != nil {
		return nil, err
	}

	files := map[string]bool{}
	for _, file := range list {
		splits := strings.SplitN(file.Name(), ".", 3)
		if len(splits) == 3 && strings.HasPrefix(splits[2], "feed2maildir") {
			files[splits[1]] = true
		}
	}
	return files, nil
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
	tmpDir := path.Join(*flagMaildir, "tmp")
	if !dirExists(curDir) {
		fmt.Println("Maildir/cur directory does not exist: ", *flagMaildir)
		os.Exit(1)
	}
	if !dirExists(newDir) {
		fmt.Println("Maildir/new directory does not exist: ", *flagMaildir)
		os.Exit(1)
	}
	if !dirExists(newDir) {
		fmt.Println("Maildir/tmp directory does not exist: ", *flagMaildir)
		os.Exit(1)
	}

	// we read all the files in the format we write in the new and cur dirs
	// if the hash matches of the new mail we'd create, skip writing new mail
	curDirFiles, err := filesInDirToHashSet(curDir)
	if err != nil {
		fmt.Println("Cannot read files from ", curDir)
		os.Exit(1)
	}

	newDirFiles, err := filesInDirToHashSet(newDir)
	if err != nil {
		fmt.Println("Cannot read files from ", newDir)
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	fp := gofeed.NewParser()
	// "http://feeds.twit.tv/twit.xml"
	feed, err := fp.ParseURLWithContext(*flagFeed, ctx)
	if err != nil {
		fmt.Printf("Error requesting %s: %s", *flagFeed, err.Error())
	}

	h := xxhash.New()

	for _, item := range feed.Items {
		h.Reset()
		h.Write([]byte(item.Title))
		h.Write([]byte(item.Link))
		sum := h.Sum64()
		sumStr := fmt.Sprintf("%d", sum)
		mailFilename := fmt.Sprintf("0.%d.feed2maildir", sum)

		fullNewFilename := path.Join(newDir, mailFilename)
		fullTmpFilename := path.Join(tmpDir, mailFilename)

		if curDirFiles[sumStr] {
			// fmt.Printf("%s exists in cur dir\n", mailFilename)
			continue
		}

		if newDirFiles[sumStr] {
			// fmt.Printf("%s exists in new dir\n", mailFilename)
			continue
		}

		f, err := os.Create(fullTmpFilename)
		if err != nil {
			fmt.Printf("Error writing %s: %s", fullNewFilename, err)
		}

		fmt.Fprintf(f, "Content-Transfer-Encoding: base64\n")
		timeNowString := time.Now().UTC().Format(time.RFC1123Z)
		fmt.Fprintf(f, "Date: %s\n", timeNowString)
		fmt.Fprintf(f, "Content-Type: text/plain; charset=UTF-8\n")
		fmt.Fprintf(f, "Message-Id: <%s@local>\n", mailFilename)
		fmt.Fprintf(f, "From: %s <feed@local>\n", mimeEncode(feed.Title))
		fmt.Fprintf(f, "Subject: %s\n", mimeEncode(item.Title))
		fmt.Fprintf(f, "\n")
		content := fmt.Sprintf("\n%s\n\n%s\n%s", item.Link, item.Description, item.Content)
		fmt.Fprintf(f, "%s", base64.StdEncoding.EncodeToString([]byte(content)))
		f.Close()
		err = os.Rename(fullTmpFilename, fullNewFilename)
		if err != nil {
			fmt.Printf("error renaming: %s -> %s\n", fullTmpFilename, fullNewFilename)
			continue
		}

		fmt.Printf("%s: %s -> %s\n", feed.Title, item.Title, fullNewFilename)
	}

}
