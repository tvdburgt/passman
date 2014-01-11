# passman

passman is an offline command-line tool for managing your passwords.

## Installation
[Install Go](http://golang.org/doc/install), and install from source:

    $ go get github.com/tvdburgt/passman

`passman` only works for linux/386 and linux/amd64 for now.

## Getting started

### Creating a passman store

`passman` stores the passwords and other related data in an encrypted file
called "store". In order to create a store, simply type:

    $ passman init
    Initialized empty passman store: '/home/tman/.pass_store'


You will first be prompted for passphrase. This passphrase will indirectly be
used as key to encrypt your store. As you can see, the default store file is
`$HOME/.pass_store`. An alternate file location can be enforced using the
`$PASSMAN_STORE` environment variable or the `-file` command-line flag (in
increasing order of precedence). Defining the store file works in exactly the
same way as the other `passman` subcommands (if applicable).

Or import from an existing file:

    $ passman import -format keepass2 ~/secret.xml 

### Adding entries

Adding new password entries and modifying existing entries is done with the
`passman set` command. Entries are distinguished and referenced by a unique
identifier. For example, to create an entry for my GitHub account with username
`tvdburgt`:

    $ passman set -name tvdburgt github

This command will prompt for a password to be associated with this entry. You
can fill this in manually or generate one to your liking.

Aside from name and password, arbitrary key-value entry data can be attached
using the `-meta` flag:

    $ passman set -meta url=https://github.com/login \
        -meta description='My favorite coding site!' github

### Querying entries

To show the contents of an individual entry use `passman get`:

    $ passman get github

Listing all entries is handled by the `passman list` command.

    $ passman list

For simplicity reasons, `passman` does not provide built-in support for
maintaining a hierarchy for entries. you can however, group entries by id. The
`list` command accepts a search pattern that can be used to filter entries by
matching against a given regex pattern:


unix directory name convention

    $ passman list '^news/'
    news/hn
    news/reddit
    news/slashdot

clipboard:

    $ passman clip -timeout 20 foo

### Additional information

The complete list of commands can viewed with `passman help`. Technical
information about the design and security can be found [here](SECURITY.md).

LICENSE
