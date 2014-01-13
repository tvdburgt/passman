# passman

`passman` is an offline command-line tool for managing your passwords.

## Installation

To install from source, you need to have [Go](http://golang.org/doc/install) installed.
Issuing the following command will download and install `passman` and its
dependencies:

    $ go get github.com/tvdburgt/passman

Currently, only `go1.2` and greater is supported for the following systems:
`linux/386` and `linux/amd64`.

## Getting started

### Creating a passman store

`passman` stores the passwords and other related data in an encrypted file
called "store". In order to create a store, simply type:

    $ passman init
    [enter passphrase]
    [verify passphrase]
    Initialized empty passman store at '/home/tman/.pass_store'.

You will first be prompted for passphrase. This passphrase will indirectly be
used as a key to encrypt and decrypt your store. As you can see, the default
store file is `$HOME/.pass_store`. An alternate file location can be enforced
using the `$PASSMAN_STORE` environment variable or the `-file` or `-f`
command-line flag (in increasing order of precedence). Defining the store file
works in exactly the same way for all other store-related `passman` subcommands.

If you want to migrate from a different password manager, say KeePassX, you can
use `passman import` to import entries from an exported XML file:

    $ passman import -format keepassx ~/kpx.xml 
    Imported 42 entries to '/home/tman/.pass_store'.

### Adding and modifying entries

Adding new password entries and modifying existing entries is done with the
`passman set` command. Entries are distinguished and referenced by a unique
human-readable identifier. For example, to create an entry for my GitHub
account, I can use the id `github` accompanied with my username `tvdburgt`:

    $ passman set -name tvdburgt github

This command will prompt for a password to be associated with this entry. You
can fill this in manually or generate one to your liking.

Aside from name and password, arbitrary key-value entry data can be attached
using the `-meta` flag:

    $ passman set -meta url=https://github.com/login \
        meta description='My favorite coding site!' github

### Querying entries

To show the contents of an individual entry use `passman get`:

    $ passman get github

Listing all entries is handled by the `passman list` command.

    $ passman list

For simplicity reasons, `passman` does not maintain an entry hierarchy. It
encourages the user to devise a personal scheme by incorporating hierarchy
information in the naming of entry identifiers.

The `passman list` command conveniently accepts a regex pattern argument, which
can be used to group entries. For example, by using a directory-like identifier
scheme, I can easily list all entries that start with a given subpath:

```bash
$ passman list ^news/
news/hn
news/reddit
news/slashdot
````

[replace with actual list output]

### Exchanging entry data

To put the password of my `github` entry on the system clipboard, simply type:

````bash
$ passman clip github
Listening for clipboard requests...            
[wait 20 seconds]
Timeout reached. Exiting.
````

This will make the password available for selection requests for a limited
amount of time. Other entry fields can be copied as well:

````bash
$ passman clip -fields url,name,password github
Listening for clipboard requests...            
Field "url" requested by "Iceweasel"
Field "name" requested by "Iceweasel"
Field "password" requested by "Iceweasel"
All field values are copied. Exiting.
````

The above command will subsequently copy the metadata value of key `url` and
field `name` and `password` for the following consecutive selection requests.

### Additional information

The complete list of commands can viewed with `passman help`. Use `passman help
<cmd>` for more information about a particular subcommand.  Technical
information about the design and security can be found [here](SECURITY.md).

[add license info]
