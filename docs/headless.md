# Headless / Non-Interactive Mode

Flashingestor is primarily a TUI tool: you authenticate, then press `Ctrl+l`
(Ingest), `Ctrl+r` (Remote Collection), and `Ctrl+s` (Convert) by hand. For
scripting and automation it can also run **fully non-interactively** from the
command line, producing the BloodHound `.zip` without ever opening the TUI.

## TL;DR

Add `--headless` to your usual authentication command:

```bash
flashingestor \
  -u tbrady@rebound.htb -p '543BOMBOMBUNmanda' \
  --dc 10.129.232.31 --dns 10.129.232.31 \
  --headless
```

This runs **Ingest → Convert**, writes the dump, and exits. The resulting
BloodHound archive is placed under the output directory:

```
output/bloodhound/<timestamp>_BloodHound.zip
```

For example:

```
output/bloodhound/20260731173915_BloodHound.zip
```

No TUI is started and no keypresses are required.

## New command-line flags

| Flag | Default | Description |
| --- | --- | --- |
| `--headless` | `false` | Run non-interactively without the TUI. Executes the steps selected by `--steps`, then exits. |
| `--steps` | `ingest,convert` | Comma-separated steps to run in headless mode. Valid values: `ingest`, `remote`, `convert`. |

`--headless` has no effect on authentication flags — use the same `-u`/`-p`,
`-H`, `--pfx`, `--cert`/`--key`, `--aes-key`, `--ccache`, `--remote-*`, etc.
options you would use in the TUI.

## Selecting steps

Steps always execute in the canonical order **ingest → remote → convert**,
regardless of the order you list them in `--steps`. Unrecognized or missing
steps are ignored; an empty `--steps` falls back to `ingest,convert`.

### DC-only collection (default)

Mirrors the `Ctrl+l` → `Ctrl+s` workflow you would do by hand. This is the
default and needs no extra flags:

```bash
flashingestor -u <USER>@<DOMAIN> -p <PASSWORD> --dc <DC> --dns <DNS> --headless
```

### Full collection including remote (active)

Adds the remote collection step (`Ctrl+r` equivalent) — active RPC/SMB/HTTP
collection against discovered computers. Use the same credentials, or a
separate local-admin account via the `--remote-*` flags:

```bash
flashingestor -u <USER>@<DOMAIN> -p <PASSWORD> \
  --dc <DC> --dns <DNS> \
  --steps ingest,remote,convert --headless
```

> **Note:** Remote collection reaches out to every discovered computer over the
> network and is louder/slower than the LDAP-only ingest+convert path. It is
> skipped automatically if no remote-capable credentials are available.

### Convert only (reuse existing msgpack data)

If you already have `output/ldap/**/*.msgpack` from a previous run, you can skip
re-ingestion and just regenerate the BloodHound zip:

```bash
flashingestor --steps convert --headless
```

## Output

Headless mode uses the same `--outdir` (`./output` by default) and the same
intermediate layout as the TUI:

```
output/
├── ldap/         # intermediate msgpack files (per domain / per forest)
├── remote/       # intermediate remote-collection msgpack files
└── bloodhound/   # final BloodHound JSON + <timestamp>_BloodHound.zip
```

By default the conversion step both **compresses** the JSON into the `.zip` and
**cleans up** the individual JSON files afterward (`compress_output` and
`cleanup_after_compression` are both `true` in the default config). The
`output/ldap` and `output/remote` msgpack files are intentionally retained so
they can be re-converted or inspected (e.g. via `cmd/ingest2json`) without
re-collecting. Delete them manually if you don't need them.

## Logging

In headless mode, log lines are written to **stdout** (with tview color tags
stripped) and timestamped, for example:

```
[2026-07-31 17:39:11] 🚀 Starting LDAP ingestion of "REBOUND.HTB"...
[2026-07-31 17:39:15] ✅ Ingestion of "REBOUND.HTB" completed in 3 seconds
[2026-07-31 17:39:15] 🔀 Starting BloodHound conversion...
[2026-07-31 17:39:15] ✅ BloodHound dump: "output/bloodhound/20260731173915_BloodHound.zip" (74.2 KB)
[2026-07-31 17:39:15] ✅ Headless run finished.
```

`--log <file>` still works and mirrors the same output to a file.
`-v` / `-vv` increase verbosity as in the TUI.

## Exit behavior

The process returns **0** after all selected steps complete and logs are
flushed. Steps that cannot run are skipped with a warning rather than aborting
the whole run, for example:

- `ingest` is skipped if no ingestion credentials are supplied or the initial
  domain cannot be determined.
- `remote` is skipped if no remote-capable credentials are supplied.
- `convert` is skipped if it is not listed in `--steps`.

Because there is no overwrite prompt in headless mode, existing intermediate
`msgpack` files are overwritten on `ingest`/`remote` (the same as choosing
“Yes” in the interactive overwrite dialog).

## How it relates to the TUI

Nothing about the interactive TUI changes — `--headless` simply selects a
different execution path:

- The same underlying engine performs ingestion, remote collection, and
  conversion (the progress-update channels are nil-guarded, so the engine runs
  identically with or without a UI).
- In headless mode the TUI is never started; the UI update primitives that would
  otherwise wait on the (non-running) event loop are neutralized, and log output
  is redirected to stdout.
- The three steps run sequentially to completion before the program exits.

In short: if you can collect something with the TUI, you can collect the same
thing with `--headless` and the same flags.
