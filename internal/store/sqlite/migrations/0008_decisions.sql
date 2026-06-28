-- 0008_decisions.sql
-- Ticket-attached durable decision records (ADR 0051, docs/contracts/decision-records.md).
-- The authoritative copy is the per-ticket sidecar .groundwork/tickets/<id>/decisions.ndjson;
-- this table is the live projection that lets pending input/approval/decision queues
-- rebuild after a store purge and be queried by status. doc_json holds the full canonical
-- record; seq is the append order within a ticket (unique per ticket).
CREATE TABLE decisions (
  ticket_id   TEXT    NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
  seq         INTEGER NOT NULL,
  decision_id TEXT    NOT NULL DEFAULT '',
  event_type  TEXT    NOT NULL,
  status      TEXT    NOT NULL,
  doc_json    TEXT    NOT NULL,
  created_at  TEXT    NOT NULL,
  PRIMARY KEY (ticket_id, seq)
);
CREATE INDEX idx_decisions_ticket ON decisions(ticket_id);
CREATE INDEX idx_decisions_status ON decisions(status);
CREATE INDEX idx_decisions_did ON decisions(ticket_id, decision_id);
