---

## Agents

The `agents/` directory contains custom-built telemetry collectors and helpers, primarily written in Go, with supporting shell tooling where appropriate.

These agents are designed to:
- Be consumed by Telegraf or similar collectors
- Expose system and platform state not available via standard plugins
- Run safely via cron, systemd, or scheduled execution
- Produce predictable, parseable output for ingestion

Current agent capabilities include:
- DNF/YUM update availability checks
- Last patch/update timestamp tracking
- Chrony time synchronization health and drift
- RHEL-compatible static binaries for simplified deployment

Each agent includes its own documentation describing usage, output, and integration patterns.

---

## Configurations

The `configs/` directory contains reference and production-tested configuration examples for:
- Telegraf inputs and exec integrations
- VictoriaMetrics agent, metrics, and logs pipelines
- Loki log ingestion and labeling strategies

These configurations are intended to be **opinionated but adaptable**, reflecting patterns proven in real environments.

---

## Philosophy

This repository follows an **automation-first and observability-by-design** approach:

- Build small, focused components
- Favor reliability over complexity
- Design for failure, retries, and partial execution
- Document intent and behavior clearly

Where useful, projects are made public to share patterns and reduce duplication of effort for others facing similar challenges.

---

## Status

This repository is actively evolving as new agents, integrations, and configurations are added over time.