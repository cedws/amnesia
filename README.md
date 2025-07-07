# amnesia

amnesia is a command-line tool for sealing secrets with a set of questions, with a specified number of answers required to decrypt the secret. It uses argon2id and [Shamir's Secret Sharing](https://en.wikipedia.org/wiki/Shamir%27s_secret_sharing) under the hood.

## How it works

Upon sealing a secret, the user is asked to provide a set of questions and answers, and a threshold. The threshold is the number of questions that must be answered correctly to unseal the secret, and must be at least 2. The user is prompted with test questions before sealing the secret to ensure they have been inputted correctly.

## Installation

```bash
go install github.com/cedws/amnesia@latest
```

## Usage

### Sealing a secret
```bash
echo "my-master-password" | amnesia seal -f secret.json
```

### Unsealing a secret
```bash
amnesia unseal -f secret.json
```

### More examples
```bash
# Seal SSH key without compression
cat ~/.ssh/id_rsa | amnesia seal --no-compress -f ssh-key.json

# Unseal to file
amnesia unseal -f sealed.json -o recovered.txt
```

## Cryptography

The cryptography behind amnesia is argon2id, AES-CTR, and [Shamir's Secret Sharing](https://en.wikipedia.org/wiki/Shamir%27s_secret_sharing).

Upon encryption:

1. The secret is optionally compressed with gzip (default behaviour)
2. A SHA256 hash is appended to the secret for integrity verification
3. The secret is split into M shares using Shamir's Secret Sharing, where M is the number of questions
4. An encryption key is derived from each answer using argon2id key derivation function
5. Each share is encrypted with its corresponding key using AES-CTR
6. The encrypted shares are stored alongside their corresponding questions

The size of the sealed secret roughly scales like the following:

```
Total Output Size = M × (N + O)
```

Where:
- **N** = secret size in bytes
- **M** = number of shares (questions)
- **O** = overhead per share (typically 1-10 bytes)

Shares are encrypted with AES-CTR, an unauthenticated cipher, to make shares more difficult to brute force. Integrity of the resulting secret is ensured with the trailing SHA256 checksum.

For strong protection of the secret, enter a good number of difficult questions. An example usage could be to enter your last five passwords as questions.

> [!WARNING]
> *amnesia* has not been cryptographically audited, use at your own risk.
> If you're interested in helping out with this, please get in touch.

## Inspiration

I've been thinking about how I would recover my digital life if I suffered from amnesia and forgot all my passwords. The idea behind *amnesia* is that I could seal a master password with a set of questions about my life I'd be able to answer even with memory loss, and use that master password to recover everything else.
