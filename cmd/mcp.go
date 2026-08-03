package cmd

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	kizamicontext "github.com/mskasa/kizami/internal/context"
	"github.com/spf13/cobra"
)

var mcpAllowWrite bool

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Run an MCP server exposing decisions to AI agents",
	Long: `Run an MCP server over stdio exposing three read-only tools:

  kizami_decisions_for_files  Which decisions govern the given files?
  kizami_search_decisions     Has this already been decided?
  kizami_get_decision         Give me a decision's full text.

See docs/decisions/2026-08-03-mcp-tools-as-questions-not-verbs.md for the design.

With --allow-write, a fourth tool is also registered:

  kizami_record_decision      Record a new decision (always Status: Draft, new-file-only).

See docs/decisions/2026-08-03-agent-authored-decision-write-path.md for the safety design.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		root, err := gitRepoRootFn()
		if err != nil {
			return err
		}
		cfg := loadCfg()
		dirs := documentDirs(root, cfg)

		server := newMCPServer(root, dirs, decisionsDir(root, cfg), designDir(root, cfg), mcpAllowWrite)
		return server.Run(cmd.Context(), &mcp.StdioTransport{})
	},
}

func init() {
	mcpCmd.Flags().BoolVar(&mcpAllowWrite, "allow-write", false, "Also register kizami_record_decision (write access)")
	rootCmd.AddCommand(mcpCmd)
}

type decisionsForFilesInput struct {
	Paths []string `json:"paths" jsonschema:"repo-root-relative file paths you are about to read or edit"`
	Full  bool     `json:"full,omitempty" jsonschema:"return each decision's full document text instead of just its summary section"`
}

type searchDecisionsInput struct {
	Query             string `json:"query" jsonschema:"keyword to search for across decisions"`
	Limit             int    `json:"limit,omitempty" jsonschema:"maximum number of results (default 10)"`
	IncludeSuperseded bool   `json:"include_superseded,omitempty" jsonschema:"include decisions that have been superseded by a later one (default: excluded)"`
}

type searchDecisionsOutput struct {
	Results []*kizamicontext.SearchResult `json:"results"`
}

type getDecisionInput struct {
	Slug string `json:"slug" jsonschema:"the decision's slug, as returned by kizami_search_decisions or kizami_decisions_for_files"`
}

type getDecisionOutput struct {
	Matches []*kizamicontext.Document `json:"matches"`
}

type recordDecisionInput struct {
	Kind         string   `json:"kind,omitempty" jsonschema:"\"adr\" (default) or \"design\""`
	Title        string   `json:"title" jsonschema:"the decision's title"`
	RelatedFiles []string `json:"related_files" jsonschema:"files related to this decision"`

	Context      string `json:"context,omitempty" jsonschema:"(adr) why this decision was needed"`
	Decision     string `json:"decision,omitempty" jsonschema:"(adr) what was decided, 1-3 sentences"`
	Consequences string `json:"consequences,omitempty" jsonschema:"(adr) impact, benefits, and trade-offs"`
	Alternatives string `json:"alternatives,omitempty" jsonschema:"(adr) options considered but not adopted"`

	Overview           string `json:"overview,omitempty" jsonschema:"(design) 1-3 sentences summarizing what this design does and why"`
	Background         string `json:"background,omitempty" jsonschema:"(design) why this design was needed"`
	Goals              string `json:"goals,omitempty" jsonschema:"(design) what this design achieves"`
	NonGoals           string `json:"non_goals,omitempty" jsonschema:"(design) what this design explicitly does not cover"`
	DesignNotes        string `json:"design_notes,omitempty" jsonschema:"(design) the actual design: structure, flow, interfaces, etc."`
	ImplementationPlan string `json:"implementation_plan,omitempty" jsonschema:"(design) steps to implement this design"`
	OpenQuestions      string `json:"open_questions,omitempty" jsonschema:"(design) unresolved questions at design time"`
}

type recordDecisionOutput struct {
	Path string `json:"path"`
	Slug string `json:"slug"`
}

// newMCPServer builds the kizami MCP server. dirs are the configured document directories
// (already joined with root, used by the read tools); decisionsDir/designDir are the
// specific directories kizami_record_decision writes into when allowWrite is true; root is
// the repo root, used to compute repo-relative paths and check Related Files existence.
func newMCPServer(root string, dirs []string, decisionsDir, designDir string, allowWrite bool) *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{Name: "kizami", Version: Version}, nil)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "kizami_decisions_for_files",
		Description: "Which decisions (ADRs / design docs) govern the given files? Call this before editing a file to see constraints already recorded for it.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, args decisionsForFilesInput) (*mcp.CallToolResult, *kizamicontext.Result, error) {
		if len(args.Paths) == 0 {
			return nil, nil, fmt.Errorf("paths must not be empty")
		}
		result, err := kizamicontext.Resolve(dirs, root, args.Paths, args.Full)
		if err != nil {
			return nil, nil, err
		}
		return nil, result, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "kizami_search_decisions",
		Description: "Has this already been decided? Keyword search across recorded decisions, useful before implementing something that might duplicate or contradict a past decision.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, args searchDecisionsInput) (*mcp.CallToolResult, *searchDecisionsOutput, error) {
		if args.Query == "" {
			return nil, nil, fmt.Errorf("query must not be empty")
		}
		results, err := kizamicontext.Search(dirs, root, args.Query, args.Limit, args.IncludeSuperseded)
		if err != nil {
			return nil, nil, err
		}
		return nil, &searchDecisionsOutput{Results: results}, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "kizami_get_decision",
		Description: "Get a decision's full text by slug. May return more than one match (e.g. translated variants of the same decision).",
	}, func(_ context.Context, _ *mcp.CallToolRequest, args getDecisionInput) (*mcp.CallToolResult, *getDecisionOutput, error) {
		if args.Slug == "" {
			return nil, nil, fmt.Errorf("slug must not be empty")
		}
		docs, err := kizamicontext.GetBySlug(dirs, root, args.Slug)
		if err != nil {
			return nil, nil, err
		}
		if len(docs) == 0 {
			return nil, nil, fmt.Errorf("no decision found with slug %q", args.Slug)
		}
		return nil, &getDecisionOutput{Matches: docs}, nil
	})

	if allowWrite {
		mcp.AddTool(server, &mcp.Tool{
			Name:        "kizami_record_decision",
			Description: "Record a new decision (ADR or design doc) immediately after making it. Always created as Status: Draft for human review; never edits or deletes an existing file. See docs/decisions/2026-08-03-agent-authored-decision-write-path.md.",
		}, func(_ context.Context, _ *mcp.CallToolRequest, args recordDecisionInput) (*mcp.CallToolResult, *recordDecisionOutput, error) {
			path, slug, err := kizamicontext.Record(decisionsDir, designDir, kizamicontext.RecordInput{
				Kind:               args.Kind,
				Title:              args.Title,
				RelatedFiles:       args.RelatedFiles,
				ADRContext:         args.Context,
				ADRDecision:        args.Decision,
				Consequences:       args.Consequences,
				Alternatives:       args.Alternatives,
				Overview:           args.Overview,
				Background:         args.Background,
				Goals:              args.Goals,
				NonGoals:           args.NonGoals,
				Design:             args.DesignNotes,
				ImplementationPlan: args.ImplementationPlan,
				OpenQuestions:      args.OpenQuestions,
			})
			if err != nil {
				return nil, nil, err
			}
			return nil, &recordDecisionOutput{Path: path, Slug: slug}, nil
		})
	}

	return server
}
