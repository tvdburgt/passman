# passman security

- scrypt kdf (default work params for now) (planning to make this variable and
  store-dependant)
- aes-256 in ctr mode
- hmac: encrypt-then-mac
- new salt is generated each time a store mutation is made
- json entries (show `passman export`)
- plaintext passwords and keys are stored as mutable types and cleared from
  memory when the program is done processing them
- password generator methods

# memory issues
- sensitive data can be swapped to disk
- sensitive data can be copied by GC
