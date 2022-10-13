# Email Backup
A small tool for administrate and backup (locally) your email account

## Installation
You can install the tooling via (go must be installed):
1. `git clone git@github.com:sikupe/email-backup.git`
2. `cd email-backup`
3. `go install`

## Functionality
### Backup
Downloads all emails of your email account or a folder in your email account to a local output folder.
This is purely based on IMAP and can be used via:

`email-backup backup --output {output dir, optional} --path {folder to backup, optional} --server {host:port for IMAP} --user {username for IMAP} --password {password for IMAP}`

Server, user and password can also be provided through env variables called `IMAP_SERVER`, `IMAP_USER` nad `IMAP_PASSWORD`.

### Sizes
This shows the size of each folder in your email account as well as the total space used in the email account.
This is purely based on IMAP and can be used via:

`email-backup sizes {--sort-by-sizes} --path {folder to backup, optional} --server {host:port for IMAP} --user {username for IMAP} --password {password for IMAP}`

Server, user and password can also be provided through env variables called `IMAP_SERVER`, `IMAP_USER` nad `IMAP_PASSWORD`.
