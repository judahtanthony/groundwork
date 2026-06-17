# Workflow Contract

`.groundwork/WORKFLOW.md` is the repo-owned agent operating contract.

It should follow the Symphony-style shape: optional YAML front matter plus a Markdown prompt body rendered with ticket and run context.

## Purpose

`WORKFLOW.md` tells agents how to operate in the managed project:

- how to interpret work nodes,
- how to triage a claimed node as leaf or composite,
- how to decompose a composite node into children and a parent contract, and submit it as a proposal,
- how to follow task-type SOPs and context,
- how to update progress,
- how to request approvals, including the `decompose` gate,
- how to escalate revisions upward,
- how to validate work,
- how to hand off for review,
- how to land when policy permits.

## V1 Boundary

The workflow is durable committed state. It is not runtime state. Changes to workflow should be reviewed like other project policy changes.

