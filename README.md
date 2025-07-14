# amnesia

amnesia is a command-line tool for sealing secrets with a set of questions, with a specified number of answers required to decrypt the secret. It uses argon2id and [Shamir's Secret Sharing](https://en.wikipedia.org/wiki/Shamir%27s_secret_sharing) under the hood.

I've been thinking about how I would recover my digital life if I suffered from amnesia and forgot all my passwords. The idea behind *amnesia* is that I could seal a master password with a set of questions about my life I'd be able to answer even with memory loss, and use that master password to recover everything else.

## How it works

Upon sealing a secret, the user is asked to provide a set of questions and answers, and a threshold. The threshold is the number of questions that must be answered correctly to unseal the secret, and must be at least 2. The user is prompted with test questions before sealing the secret to ensure they have been inputted correctly.

## Demo

![Demo](docs/amnesia.gif)

## Installation

```bash
go install github.com/cedws/amnesia@latest
```

## Usage

### Sealing a secret
```bash
# Basic usage
echo "my-master-password" | amnesia seal -f sealed.json

# Seal without test questions
echo "my-master-password" | amnesia seal -f sealed.json -t
```

### Unsealing a secret
```bash
# Basic usage
amnesia unseal -f sealed.json

# Unseal to output file
amnesia unseal -f sealed.json -o secret.json
```

For strong protection of the secret, enter a good number of difficult questions. An example usage could be to enter your last five passwords as questions.

Answers are used verbatim in key-derivation, so be mindful of usage of casing and punctuation.

## Cryptography

> [!WARNING]
> amnesia has not been cryptographically audited, use at your own risk.
> If you're interested in helping out with this, please get in touch.

The cryptography used in amnesia is argon2id, AES-CTR, AES-GCM and [Shamir's Secret Sharing](https://en.wikipedia.org/wiki/Shamir%27s_secret_sharing).

1. A 32 byte DEK (data encryption key) is generated
2. The DEK is split into N shares using Shamir's Secret Sharing, where N is the number of questions
3. A 32 byte KEK (key encryption key) is derived from each answer using argon2id KDF
4. Each share of the DEK is encrypted with an answer KEK using AES-CTR
5. The encrypted shares are stored alongside the corresponding questions
6. The secret is encrypted with the DEK using AES-GCM

This hybrid method of encrypting a secret with a DEK and splitting the DEK into parts with SSS means very large secrets can be protected with minimal overhead.
