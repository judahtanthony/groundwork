//go:build ignore

// bootstrap_import is a ONE-SHOT transcription tool (not part of the gw binary):
// it renders docs/plan/work-tree.yaml into canonical Markdown ticket exports under
// .groundwork/tickets/ so `gw ticket import` can ingest them (ADR 0032). The gw
// binary never reads YAML; this script is the mechanical authoring pass and the
// durable artifacts are the committed ticket.md files. Run from the repo root:
//
//	go run scripts/bootstrap_import.go
//
// Statuses reflect what M1/M2 actually delivered: completed work is `done`,
// Phase 4/5 and deferred surfaces are `backlog`, the dogfood capstone is `todo`.
package main

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
	"groundwork/internal/config"
	"groundwork/internal/exporter"
	"groundwork/internal/ticket"
)

type node struct {
	ID         string   `yaml:"id"`
	Kind       string   `yaml:"kind"`
	Title      string   `yaml:"title"`
	Acceptance []string `yaml:"acceptance"`
	Children   []node   `yaml:"children"`
}

type doc struct {
	Root node `yaml:"root"`
}

func set(ids ...string) map[string]bool {
	m := map[string]bool{}
	for _, id := range ids {
		m[id] = true
	}
	return m
}

// Status partitions. Anything not listed defaults to `done` (completed M1/M2 work).
var (
	backlog = set(
		"E-0006", // Codex runtime adapter — Phase 4
		"T-0501", "T-0502", "T-0503", "T-0504", "T-0505",
		"T-0904", // resume from checkpoint — Phase 4
		"T-0801", "T-0802", // dashboard HTML — deferred web surface
		"T-0903", // generated view export — ties to dashboard
		"T-1003", // first Codex-assisted ticket — Phase 4
	)
	inProgress = set("G-0001", "E-0009", "E-0010", "E-0011")
	todo       = set("T-1002") // the dogfood capstone, run live through gw
	docWork    = set("T-0001", "T-0002", "T-0003", "T-1001", "T-1002")
)

func statusFor(id string) ticket.Status {
	switch {
	case backlog[id]:
		return ticket.StatusBacklog
	case inProgress[id]:
		return ticket.StatusInProgress
	case todo[id]:
		return ticket.StatusTodo
	default:
		return ticket.StatusDone
	}
}

func main() {
	data, err := os.ReadFile("docs/plan/work-tree.yaml")
	check(err)
	var d doc
	check(yaml.Unmarshal(data, &d))

	proj := &config.Project{Root: "."}
	ticketsDir := proj.TicketsDir()
	check(os.MkdirAll(ticketsDir, 0o755))

	n := walk(d.Root, "", ticketsDir)
	fmt.Printf("wrote %d ticket exports under %s\n", n, ticketsDir)
}

func walk(nd node, parentID, ticketsDir string) int {
	nodeType := ticket.NodeComposite
	workType := ""
	if nd.Kind == "ticket" {
		nodeType = ticket.NodeLeaf
		workType = "technical_implementation"
	}
	if docWork[nd.ID] {
		workType = "documentation"
	}
	t := &ticket.Ticket{
		ID:         nd.ID,
		Kind:       nd.Kind,
		Title:      nd.Title,
		NodeType:   nodeType,
		WorkType:   workType,
		Status:     statusFor(nd.ID),
		ParentID:   parentID,
		Acceptance: nd.Acceptance,
	}
	_, err := exporter.WriteTo(ticketsDir, t, nil)
	check(err)
	count := 1
	for _, c := range nd.Children {
		count += walk(c, nd.ID, ticketsDir)
	}
	return count
}

func check(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "bootstrap_import:", err)
		os.Exit(1)
	}
}
