VERSION=0.1.0

BINDIR=$HOME/bin

GO?=go
GOFLAGS?=

GOSRC!=find . -name '*.go'
GOSRC+=go.mod go.sum

feed2maildir: $(GOSRC)
	$(GO) build $(GOFLAGS) \
		-ldflags "-X main.Version=$(VERSION)" \
		-o $@

all: feed2maildir

# Exists in GNUMake but not in NetBSD make and others.
RM?=rm -f

clean:
	$(RM) feed2maildir

install: all
	mkdir -m755 -p $(BINDIR)
	install -m755 feed2maildir $(BINDIR)/feed2maildir

RMDIR_IF_EMPTY:=sh -c '\
if test -d $$0 && ! ls -1qA $$0 | grep -q . ; then \
	rmdir $$0; \
fi'

uninstall:
	$(RM) $(BINDIR)/feed2maildir
	${RMDIR_IF_EMPTY} $(BINDIR)

.DEFAULT_GOAL := all

.PHONY: all clean install uninstall
