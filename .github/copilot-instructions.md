# Copilot instructions — Manalyzer

Purpose: make targeted, safe edits to implement or refactor the CS:GO demo analyzer UI and data pipeline.

Quick start
- Build: `go build` (Go 1.23+)
- Run (dev): `go run .`
- On Windows after build: `./manalyzer.exe`

Key entry points
- `main.go` — program entry and app wiring
- `src/gather.go` — file discovery and demo parsing (exports `GatherAllDemosFromPath`, `GatherDemo`)
- `src/wrangle.go` — transform `[]*api.Match` into player & side-specific stats (primary logic)
- `src/gui.go` — TUI using `tview` (`Form`, `TextView`, `Table`). UI updates must use `Application.QueueUpdate()` or run in main goroutine.

Important patterns & constraints (do not override)
- Use `github.com/akiver/cs-demo-analyzer` to parse demos — do not reimplement parsing or trade/KAST detection.
- Side-specific stats: determine player side per round (see `docs/FIRST_PLAN.md` for reference code). Implement custom aggregation only in `wrangle.go`.
- `GatherAllDemosFromPath` returns `[]*api.Match` and intentionally continues on per-file errors. Preserve that behavior (partial results + aggregated errors).
- SteamID64 inputs are expected as numeric 17-digit strings; forms validate numeric-only input.

Developer workflows
- Iterate quickly: run `go run .` and use a small directory of `.dem` files for fast feedback.
- Long-running analysis should run in a goroutine; post results to UI via `QueueUpdate` to avoid tview race conditions.
- Logging: GUI shows progress in the Event Log (middle panel) and standard library `log` is used for debug messages.

Conventions specific to this repo
- Support 1–5 players; `gui.go` uses a fixed form layout (5 name + 5 SteamID fields). Keep that shape unless UI redesign is approved.
- Filters live in the stats table: map and side (`T`/`CT`) filtering is applied in `StatisticsTable.SetFilter`.
- Preserve the phased implementation described in `docs/FIRST_PLAN.md`: start with `wrangle.go`, then `gather.go`, then `gui.go`.

Integration notes
- Use `*api.Match` fields directly: `match.PlayersBySteamID`, `round.TeamASide`, `round.TeamBSide`.
- cs-demo-analyzer provides per-kill flags (trade, first-kill) — prefer those over heuristic re-detection.

When making changes
- Run the app locally with a small demo set to validate UI behavior and stats correctness.
- Update `README.md` or `docs/FIRST_PLAN.md` if you change a public behavior (CLI, config, or output format).
- Open a PR to `main` and reference `docs/FIRST_PLAN.md` for algorithmic changes.

Files to check before editing
- `README.md`, `src/gather.go`, `src/wrangle.go`, `src/gui.go`.

If anything here is unclear, point to the exact file and line range you want clarified and I will expand.
