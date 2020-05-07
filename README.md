# feed2maildir

Convert RSS/Atom feeds using gofeed to maildir mail files.

I wrote this because I want to read my mail with gnu/emacs and notmuch.

Install to BINDIR(`BINDIR=$(HOME)/bin`): `make install`

Sample usage: `./fetch.sh`

Sample systemd files in `systemd`. Edit, then install with:
 `mkdir -p ~/.config/systemd/user && cp systemd/rss.* ~/.config/systemd/user/ && systemctl --user daemon-reload && systemctl --user start rss`
