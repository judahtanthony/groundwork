# ADR 0008: Prefer Open Source Local Dependencies

Status: Accepted
Implemented: Implemented

## Context

Avoiding all dependencies would waste effort. Mandatory external services would increase cost and lock-in.

## Decision

Minimize required services, not useful libraries. Prefer open-source local dependencies and open standards where practical.

## Consequences

SQLite drivers, CLI frameworks, YAML parsers, filesystem watchers, and Markdown renderers are acceptable if they reduce risk. Postgres, Redis, Temporal, Kubernetes, Docker, hosted trackers, and hosted control planes are not required for v1.
