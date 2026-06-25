---
id: T-1004
kind: epic
node_type: composite
work_type: technical_design
title: Align configuration with 12-factor app principles
status: backlog
assignee: null
requested_actor: null
priority: 0.2
labels: []
parent: T-1074
depends_on: []
created_at: "2026-06-21T12:51:31Z"
updated_at: "2026-06-25T01:24:34Z"
---

## Problem

Adopt 12-factor configuration, chiefly 'store config in the environment'. Today the coordinator bind address is configurable only via .groundwork/config.yaml and the 'gw server --addr' flag; there is no environment-variable override, and --addr moves only the server bind, not the CLI client target. Add env-var config (e.g. GW_ADDR / GW_SERVER_ADDR) honored in config.Open so server and client stay consistent, define precedence (flag > env > file > default), and review other settings for env-overridability. Tracked as a root for later development; decompose when scheduled.

## Acceptance Criteria

- The coordinator bind address can be set via an environment variable, applied in config loading so server and CLI client agree.
- Config precedence is documented and deterministic (flag > env > config file > default).
- A 12-factor review identifies any other configuration that should be environment-overridable.

## Design / Contract

_No contract recorded._

## Escalations

_No escalations._
