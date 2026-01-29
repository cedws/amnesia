# amnesia

amnesia is a command-line tool for sealing secrets with a set of questions, with a specified number of answers required to decrypt the secret. It uses argon2id and [Shamir's Secret Sharing](https://en.wikipedia.org/wiki/Shamir%27s_secret_sharing) under the hood.

I've been thinking about how I would recover my digital life if I suffered from amnesia and forgot all my passwords. The idea behind *amnesia* is that I could seal a master password with a set of questions about my life I'd be able to answer even with memory loss, and use that master password to recover everything else.

## How it works

Upon sealing a secret, the user is asked to provide a set of questions and answers, and a threshold. The threshold is the number of questions that must be answered correctly to unseal the secret, and must be at least 2. The user is prompted with test questions before sealing the secret to ensure they have been inputted correctly.

## Demo

![Demo](docs/amnesia.gif)

## Installation

### Brew

```bash
brew install cedws/tap/amnesia
```

### Scoop

```powershell
scoop bucket add cedws https://github.com/cedws/scoop-bucket.git
scoop install amnesia
```

### Go

```bash
go install github.com/cedws/amnesia@latest
```

## Usage

For strong protection of the secret, enter a good number of difficult questions. An example usage could be to enter your last five passwords as questions.

Answers are used verbatim in key-derivation, so be mindful of usage of casing and punctuation.

### Sealing a secret
```bash
# Seal a secret, output to stdout
echo "my-master-password" | amnesia seal

# Seal a secret to a file
echo "my-master-password" | amnesia seal -o sealed.json

# Seal without test questions
echo "my-master-password" | amnesia seal -o sealed.json -t
```

### Unsealing a secret
```bash
# Unseal a secret, output to stdout
amnesia unseal -f sealed.json

# Unseal a secret to a file
amnesia unseal -f sealed.json -o secret.txt
```

### Resealing a secret

Resealing allows you to replace the encrypted secret in an existing sealed file while keeping the same questions and answers. You must provide the correct answers to derive the encryption key.

```bash
# Reseal with a new secret, output to stdout
echo "new-master-password" | amnesia reseal -f sealed.json

# Reseal with a new secret to a file
echo "new-master-password" | amnesia reseal -f sealed.json -o resealed.json
```

### Opening a secret for editing

Opens a sealed secret to a file for editing. Press Ctrl+C to reseal the modified contents. The secret file is deleted on exit.

```bash
amnesia open -f sealed.json -o secret.txt
```

## age Plugin (experimental)

amnesia has experimental support for [age](https://github.com/FiloSottile/age) as an identity plugin. amnesia can generate an *age*-compatible X25519 identity sealed with questions. When age wants to decrypt data using this identity, it will prompt the user for the required answers to unseal the identity.

To install amnesia as an *age* plugin, create a symlink in your PATH which links `age-plugin-amnesia` to amnesia. You can also just clone the binary. If you've installed amnesia via brew, you don't need to do anything.

```bash
ln -s amnesia age-plugin-amnesia
# OR (not recommended)
cp amnesia age-plugin-amnesia
```

To generate an *age*-compatible identity, run:

```bash
amnesia age-keygen
```

This will interactively prompt for questions to seal the identity with, and output something like this:

```bash
# created 2025-07-16T23:55:59+09:00
# public key: age1enfsyp6vug3l4xt65jysvlpl076xkw4cxup89rmteakfzre8uajqlnya3u
AGE-PLUGIN-AMNESIA-10V9ZQGPZWEJHYUMFDAHZYW3QYGCJYTQ2YQSZYUM9V9KX2EZLW35K6ETNW3SK6UPZ8GSZYV3SXG6J6VPH95CNV4PJXVAR2DF6X5UJKVPE8GCRQG3VPGSZQGNNDPSHYETNYGAZQKC2YQSZQGRMPGSZQGPQYQSZY6TYYGAZQVPVPGSZQGPQYQSZYUT4V4EHG6T0DC3R5GPZDESK6EFZ9S9ZQGPQYQSZQGNNV9K8GG36YQ3YSN2EF5MKKD6JFATK632FFEFHYVZ3WD4NV5NR2G6NJ5R9DSUN2WT42FT9WAPH2FXKYTMN853ZCZ3QYQSZQGPQYFEKSCTJV53R5GPZ29VNSDF5DA4RVC2HXP4XSJNKG36XX3N5GAZHV5PNXAPRQKNSXEXRXC26V9F8G6TFWFHYU7NTW39YWD2VVAX9Y6TV2YU42S2YW49KYJFSGY7N6GS2YQSZQGRA9S9ZQGPQYPAS5GPQYQSZQGPZD9JZYW3QXYKQ5GPQYQSZQGPZW96K2UM5D9HKUG36YQ3XZEM9YGKQ5GPQYQSZQGPZWDSKCAPZ8GSZYK3EVFJ9V635W9XKYJT8WDYYX6TND4EXKN6RGA28YMJ5VD9YV73HWDU57UM02A2KS6N4VV7JYTQ2YQSZQGPQYQ38X6RPWFJJYW3QYF6KJ6RKW39XVJE5W9SH5S2GDEX5Y3PETPJNVET3VAZ5CJ2Y8P6XCUR4WGHKUJMS2EA926602FZ8S3RXXV69Q43ND46NQMNCD934J3NDX3GN60FZPGSZQGPQ059ZQGZA9S9ZQGPZV4HXXUNEWP6X2EPZ8GSZYW23T9M5XKZX2F68SJECFC45KMMC2434Z6J6W3GYKKR5TQ45S4MYVDC5C3T3G4RXUNN9TFC8J7PCD9C4G7Z5VEQNSVR3TG6KS4MKDDK5WDNGXPHH56M30FFRZ5T62PCY6CTR0P8575P3DUHNSU6VWAHHVU6G23TKVE30FPVJ766P24E95V6JTFDXYNMGV4U8QCMXWEKYVDNG2E2RXVTVF3MXJ3PZPF7SMZA27H
```

The public key is a "recipient" and can be shared with anyone. You can use it to encrypt data like so:

```bash
echo "secret" | age -r age1enfsyp6vug3l4xt65jysvlpl076xkw4cxup89rmteakfzre8uajqlnya3u > secret.enc
```

The long line beginning with `AGE-PLUGIN-AMNESIA-...` is the "identity." This is an encrypted X25519 key which can only be unsealed with sufficient answers to the input questions.

You can decrypt data with the identity like so:

```bash
age --decrypt -i identity.txt secret.enc
```

This will interactively prompt for answers via *age*.

> [!IMPORTANT]
> It is not currently possible to leave an answer blank due to an *age* limitation. You must provide the correct answer to all questions for successful decryption.

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
