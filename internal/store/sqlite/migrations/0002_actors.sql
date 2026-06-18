-- Actors integration (ADR 0023). Forward-only and additive (ADR 0018):
-- align the lease holder terminology with the actor model and add the node
-- routing metadata that the work-tree model (T-0308) and ticket export now
-- carry. Behavioral actor-aware routing (selection, snapshots, approvals) is
-- Phase 2; this migration only lands the schema slice.

-- The lease holder is an actor (human or AI), not specifically an agent.
ALTER TABLE leases RENAME COLUMN agent_id TO actor_id;

-- work_type: organization-defined operational metadata for SOP/policy/actor
-- routing and validation (not a status). requested_actor: optional routing hint
-- that policy must still authorize.
ALTER TABLE tickets ADD COLUMN work_type TEXT;
ALTER TABLE tickets ADD COLUMN requested_actor TEXT;
