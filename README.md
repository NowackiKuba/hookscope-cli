<div align="center">

<img src="https://app.hookscope.dev/og-image.png" alt="HookScope" width="100%" />

<br />
<br />

<h1>HookScope CLI</h1>

<p>Expose your local server to the internet and inspect every webhook in real-time.</p>

<br />

[![Release](https://img.shields.io/github/v/release/nowackikuba/hookscope?style=flat-square&color=7c3aed)](https://github.com/nowackikuba/hookscope/releases)
[![License](https://img.shields.io/github/license/nowackikuba/hookscope?style=flat-square&color=7c3aed)](LICENSE)
[![Dashboard](https://img.shields.io/badge/dashboard-app.hookscope.dev-7c3aed?style=flat-square)](https://app.hookscope.dev)

</div>

---

## What is HookScope?

You're integrating Stripe. You get a `400`. No idea what payload was sent. No idea what your server received. No logs. Nothing.

**HookScope fixes this.**

- 🔍 **Inspect** every incoming webhook in real-time
- 🚇 **Tunnel** webhooks directly to your localhost
- ⚡ **Detect** schema drift before it crashes your production
- 🤖 **Generate** TypeScript, Zod, Go, Python types automatically

---

## Installation

### macOS (Homebrew)

```bash
brew tap nowackikuba/hookscope
brew install hookscope
```

### Linux

```bash
curl -sSL https://raw.githubusercontent.com/nowackikuba/hookscope/main/install.sh | bash
```

### Manual

Download the latest binary from [GitHub Releases](https://github.com/nowackikuba/hookscope/releases).

| Platform              | Binary                   |
| --------------------- | ------------------------ |
| macOS (Apple Silicon) | `hookscope-darwin-arm64` |
| macOS (Intel)         | `hookscope-darwin-amd64` |
| Linux                 | `hookscope-linux-amd64`  |

---

## Quick Start

```bash
# 1. Login to your HookScope account
hookscope login

# 2. Forward webhooks to your local server
hookscope forward 3000
```

That's it. Open [app.hookscope.dev](https://app.hookscope.dev) to inspect incoming webhooks.

---

## Commands

### `hookscope login`

Authenticate with your HookScope account.

```bash
hookscope login
```

Opens a browser window to complete authentication. Your session is saved locally.

---

### `hookscope forward <port>`

Start a tunnel from HookScope to your local server.

```bash
hookscope forward 3000
```

```
 ✓ Tunnel established
 ✓ Forwarding to http://localhost:3000

 Webhook URL  https://app.hookscope.dev/webhooks/abc-123-xyz
 Dashboard    https://app.hookscope.dev/endpoints/abc-123-xyz

 Waiting for webhooks...
```

Copy the **Webhook URL** and paste it into Stripe, GitHub, Shopify — or any provider.
Every request will appear in your dashboard instantly.

---

## How it works

```
Stripe / GitHub / Shopify
         │
         │  POST webhook
         ▼
  app.hookscope.dev
         │
         │  inspect · verify signature · detect drift
         ▼
  hookscope forward 3000
         │
         │  forward
         ▼
  localhost:3000
```

1. Provider sends a webhook to your HookScope URL
2. HookScope inspects the payload, verifies the signature, runs scanners
3. CLI forwards the request to your local server
4. You see everything in real-time in the dashboard

---

## Dashboard features

Once you're forwarding, open [app.hookscope.dev](https://app.hookscope.dev):

| Feature                    | Description                                                             |
| -------------------------- | ----------------------------------------------------------------------- |
| **Live Inspector**         | See every webhook as it arrives, with full headers and payload          |
| **Schema Drift Detection** | Get alerted when a provider silently changes their payload              |
| **DTO Generator**          | Auto-generate TypeScript, Zod, Go, Python types from any schema         |
| **Signature Verification** | Automatic HMAC verification for Stripe, GitHub, Shopify, Clerk and more |
| **Duplicate Detection**    | Know when the same webhook is delivered multiple times                  |
| **Silence Detection**      | Get alerted when webhooks stop coming unexpectedly                      |
| **Alerts**                 | Email, Slack and Discord notifications                                  |

---

## Supported providers

Signature verification is built-in for:

- **Stripe**
- **GitHub**
- **Shopify**
- **Clerk**
- **Przelewy24**
- **Twilio** _(coming soon)_
- **Paddle** _(coming soon)_
- **Lemon Squeezy** _(coming soon)_

---

## Requirements

- macOS or Linux
- A free [HookScope account](https://app.hookscope.dev)

---

## License

MIT © [HookScope](https://app.hookscope.dev)
