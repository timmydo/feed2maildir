#!/bin/sh

fm() {
~/bin/feed2maildir --feed $1 --maildir ~/mail/rss
}
gh() {
~/bin/feed2maildir --feed $1 --maildir ~/mail/github
}

fm "https://www.phoronix.com/rss.php"
fm "https://lwn.net/headlines/rss"
fm "https://timmydouglas.com/feed.xml"
fm "https://news.ycombinator.com/rss"
gh "https://github.com/microsoft/BPerf/commits/dev.atom"
gh "https://github.com/mmcdole/gofeed/commits/master.atom"

