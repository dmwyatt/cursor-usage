# cursor-usage

> **Warning: This project was vibe coded.** An AI agent wrote virtually all of
> this code with minimal human review. It works on my machine. It might work on
> yours. It might not. Use at your own risk, read the source, and don't blame me
> when it breaks.

A CLI tool for querying your [Cursor](https://cursor.com) IDE usage and costs
via the unofficial dashboard API.

## Why

The Cursor dashboard shows usage data in a web UI, but there's no official API
for programmatic access. This tool wraps the reverse-engineered dashboard
endpoints so you can check your usage from the terminal, pipe it into scripts,
or aggregate costs by model.

## Install

Requires [Go 1.23+](https://go.dev/dl/).

```bash
go install github.com/dmwyatt/cursor-usage@latest
```

Or build from source:

```bash
git clone https://github.com/dmwyatt/cursor-usage.git
cd cursor-usage
go build -o cursor-usage .
```

## Setup

The tool authenticates using the `WorkosCursorSessionToken` cookie from your
browser session. To get it:

1. Open https://cursor.com/dashboard/usage in Chrome
2. Open DevTools (F12) > Application > Cookies > `https://cursor.com`
3. Copy the value of `WorkosCursorSessionToken`
4. Run:

```bash
cursor-usage config set-token YOUR_TOKEN_HERE
```

The token is stored locally at the platform-appropriate config path
(`~/.config/cursor-usage/config.json` on Linux,
`~/Library/Application Support/cursor-usage/config.json` on macOS,
`%LOCALAPPDATA%\cursor-usage\config.json` on Windows).

## Usage

### Billing summary

```bash
cursor-usage summary
```

Shows your plan type, billing period, request usage, and on-demand spending.

```
┌─────────────────────────────────────────────┐
│ Cursor Usage Summary                        │
├────────────────────────┬────────────────────┤
│ FIELD                  │ VALUE              │
├────────────────────────┼────────────────────┤
│ Membership             │ enterprise         │
│ Billing Start          │ 2026-04-02 09:11   │
│ Billing End            │ 2026-05-02 09:11   │
├────────────────────────┼────────────────────┤
│ Plan Used / Limit      │ 2000 / 2000 (100%) │
│ Plan Included          │ 2000               │
│ Plan Bonus             │ 6121               │
│ Plan Total Allowance   │ 8121               │
├────────────────────────┼────────────────────┤
│ On-Demand Used / Limit │ 2408 / unlimited   │
└────────────────────────┴────────────────────┘
```

### Usage events

Events default to the current billing cycle and automatically fetch all pages.

```bash
# Current billing cycle (all pages)
cursor-usage events

# Last 7 days (first page only; add --all for everything)
cursor-usage events --since 7d

# Specific date range
cursor-usage events --since 2026-04-01 --until 2026-04-03

# All billing cycles
cursor-usage events --all-time

# Fetch every page
cursor-usage events --all
```

### Aggregated costs

```bash
cursor-usage events --aggregate
```

Fetches all events in the current billing cycle, groups by model, and
calculates cost per active hour:

```
Total events: 289 (usage-based: 37, included: 251, headless: 0)
Total cost:   $105.32
Total tokens: 1443067 input, 656744 output, 5770500 cache write
Active time:  9.4h ($11.24/hr, sessions split by 30m+ gaps)

┌───────────────────────────────────┬────────┬────────┬───────────┬────────────┬─────────────────┬──────────┐
│ MODEL                             │ EVENTS │ COST   │ INPUT TOK │ OUTPUT TOK │ CACHE WRITE TOK │ HEADLESS │
├───────────────────────────────────┼────────┼────────┼───────────┼────────────┼─────────────────┼──────────┤
│ claude-4.6-opus-high-thinking     │    213 │ $87.29 │    380829 │     477616 │         4895366 │        0 │
│ claude-4.6-sonnet-medium-thinking │      9 │ $4.76  │       106 │      40686 │          416760 │        0 │
│ claude-4.6-opus-high              │      6 │ $3.48  │        35 │       8474 │          149315 │        0 │
│ ...                               │        │        │           │            │                 │          │
└───────────────────────────────────┴────────┴────────┴───────────┴────────────┴─────────────────┴──────────┘
```

Active time is inferred from event timestamps: consecutive events less than
30 minutes apart are grouped into the same session. Adjust the threshold with
`--session-gap` (in minutes):

```bash
cursor-usage events --aggregate --session-gap 15
```

### JSON output

Any command supports `--json` for machine-readable output:

```bash
cursor-usage summary --json
cursor-usage events --aggregate --json
```

### Configuration

```bash
cursor-usage config set-token TOKEN   # save token
cursor-usage config show              # show config (token redacted)
cursor-usage config path              # print config file path
```

## Caveats

- **Unofficial API.** These endpoints are reverse-engineered from the Cursor
  dashboard. They can change or break without notice.
- **Cookie auth.** The session token is an httpOnly cookie with an expiration.
  When it expires, you'll need to grab a fresh one from your browser.
- **Rate limits.** Unknown. The `--all` and `--aggregate` flags add a 200ms
  delay between page requests, but be reasonable.
- **No warranty.** See the warning at the top.

## License

MIT
