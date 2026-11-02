# Project conversation history

This note summarizes the requests, decisions, proposed approaches, and changes made during the conversation. It is intended to help resume work later; it is not a full transcript.

## Go environment and tests

- Asked for the normal shell's `GOPATH`. Ran `go env GOPATH`; it returned `/home/prism/lab/go`.
- Asked to run `go test .`. The first execution was denied; on retry it failed because there are no Go files at the workspace root.
- Asked to run `go test experimental/profit-triplet/test/util`. It passed (cached).
- Asked to run `go test ./test/profit-triplet`. It passed. The assistant initially answered only that this command ran, though the request had phrased it as running `dag_test`.

## Input generation distribution test

- Requested a test similar to `TestGenerate` that generates 1,000 inputs and checks uniformity of `Input.Margin` and `Input.Price` over `[0,1000]`, allowing up to 1% deviation.
- The assistant noted the ambiguity of interpreting uniformity over 1,001 distinct integer values and asked how to measure it.
- The developer chose ten equal-width bins, 1 percentage point tolerance, and 10,000 samples instead of 1,000.
- The developer then asked to skip bounds checks and only test distribution.
- Added `TestGenerateUniformDistribution` in `test/util/input_test.go`: 10,000 generated inputs, ten bins per field, expected 1,000 samples per bin, and tolerance of ±100. It does not check bounds. `go test ./test/util` passed.
- The developer asked for ten independent attempts where any successful attempt passes. The test was updated so each attempt generates new samples and fresh counters, passing as soon as both fields meet the tolerance. `go test ./test/util` passed.

## DAG tests and graph inspection

- Asked to run `dag_test` similarly to `input_test`; ran `go test ./test/profit-triplet`, which passed.
- Asked whether `TestGraph` could capture output from `fmt.Println`. The assistant proposed redirecting stdout to a temporary file, but the developer rejected that approach as too messy and canceled it. No stdout-capture implementation was retained.
- Instead, the developer requested using `PriceSpanSet.String()` and parsing its output to validate that it represents a DAG.
- The assistant proposed changing `PriceSpanSet.String()` to append node lines rather than index a slice by `item.index`. The developer explained that indices are packed due to a program invariant outside the assistant's scope, rejected that change, and instructed that only `dag_test.go` be changed at that point. No production `dag.go` change was kept.
- The test counter `ItemIndex` was resetting globally across test interactions. The developer requested resetting it to zero at the beginning of each test in `dag_test.go`, rather than saving/restoring it. The reset was added at the start of both tests, and `go test ./test/profit-triplet` passed.
- The developer asked whether string parsing or direct graph access was preferable. The assistant recommended introspective structural testing for the DAG invariant, because parsing display output couples the test to formatting. To avoid exposing private fields, an in-package unit test was suggested.
- Requested replacing the parse-based check in a new `dag_unit_test.go` with an introspective graph traversal. The unit test was moved from `test/profit-triplet/` to `internal/profit-triplet/`, where it can access private fields, and a DFS traverses outgoing edges to detect cycles and references outside the graph. The public/external test was left separate. `go test ./internal/profit-triplet ./test/profit-triplet` passed.

## Genetic input mutations and discovered bug

- Requested a helper in `dag_unit_test.go` that copies a base `[]Input` into ten input sets, mutates each set at least once by changing a price or swapping positions, and treats a price mutation as relevant only when it changes the relative-price relationship as defined in the request.
- Requested running the graph test for the original input plus all ten mutations, releasing per-iteration graph resources so usage does not accumulate.
- The assistant first proposed scanning every price from 0 to 1000. The developer pointed out that scanning an expanded price range could be slow and biased toward low values, and requested a randomized approach using `rand.Intn`.
- Implemented in `internal/profit-triplet/dag_unit_test.go`: ten copied mutations; randomized candidate price changes checked against other inputs with a `relativePrice` helper; swaps as an alternate mutation; and the acyclicity test run as subtests for the original plus mutations. Each graph check resets `ItemIndex` and clears its set's tree reference on return.
- `go test ./internal/profit-triplet` passed once. A repeated `go test -count=20 ./internal/profit-triplet` exposed a panic in `GraphWalk`: the interval tree rejected an interval with equal start and end values (example value: `727`). This was identified as a bug in the production graph-building path. The graph source has not yet been changed to fix it.
- The large repeated-test output was copied into `testout/dag-test-output.txt` at the developer's request to keep it inside the workspace. It contains the panic details.

## Current files and state

- `test/util/input_test.go`: contains the original range/count test and the ten-bin distribution test with ten independent attempts.
- `test/profit-triplet/dag_test.go`: external test with `ItemIndex` reset at the start of each test; contains the earlier string-based graph check if that file has not since been changed.
- `internal/profit-triplet/dag_unit_test.go`: in-package structural DAG traversal plus randomized input mutation helpers.
- `internal/profit-triplet/dag.go`: unchanged in this conversation. The identified equal-endpoint interval panic remains unfixed.
- `testout/dag-test-output.txt`: saved repeated-run output showing the failure.

## November 1: Refining and applying the symbolic-annotation grammar

- The developer asked whether `model/symbolic-abstraction.bnf` could be understood before applying it to Go source annotations. The grammar describes element annotations (`@-`) and summary annotations (`@+`), identifiers, optional package/type prefixes and element IDs, labels, aliases, substitutions, invariants, and inline Go fragments.
- The developer successively clarified the grammar rather than asking for an immediate implementation:
  - Names written as `<*-name>` are freely chosen Go identifiers, and parenthetical/bracketed productions are descriptions rather than literal tokens.
  - The previously undefined function and summary tags were defined.
  - Element lists were generalized to sequences, allowing labels, aliases, and invariants to coexist.
  - The `alias` keyword was made explicit so inline Go in an alias has a recognizable start.
  - `<go>` was specified as line-terminated, excluding newline and `*/`; the grammar then represented newline explicitly.
  - A mandatory single-space separator was introduced for adjacent labels, then the list shape was refined so multi-element lists put at most one element on each line.
  - The developer further specified that substitution and invariant lists begin with a newline; each substitution/invariant occupies its own line; there is at most one alias; aliases precede invariants; and labels precede both, remaining inline.
  - The developer pointed out that aliases must remain legal inline when there are no labels (for example `/* @-package.Function(): alias n = len(v) */`). This led to distinguishing inline elements from newline-started invariant lists.
  - The developer identified that a substitution's terminating newline must not be duplicated: the newline is accounted for by a possible following invariant list. The latest grammar makes the substitution list consume a newline before each substitution, but not after one.
- This was an iterative specification exercise: the developer supplied concrete edge cases and constraints, and the assistant tested each version for token boundaries, separators, empty productions, and line ownership. The resulting grammar is much more directly parseable because keyword starts identify element kinds and line breaks delimit free-form Go text.
- On the latest BNF, parsing was attempted against `internal/profit-triplet/dag.go`. The annotation intent can be extracted, but strict conformance is incomplete:
  - Several `@-` comments omit the required `": "` before element content (for example `@-GraphWalk#AddItem`).
  - Some labels use comma-space (`loophead, predicate`), while the BNF label-list separator is a single space.
  - Several tags include `()` or parameter signatures, although `<function-name>` is specified as an identifier and the tag grammar has no signature production.
  - One tag has multiple `#` components (`@-PriceSpanSet.GraphStep#TreeQuery#MatchCheck`), while the grammar permits at most one element locator.
  - Some summary tags are abbreviated (such as `atomic.Int32`) or include function signatures, so they do not all fit the current `<summary-tag>` form.
  - The block-comment decoration (`* ` before invariant lines) also needs a normalization rule or explicit grammar treatment.
- Despite those mismatches, the annotated control-flow intent in `GraphWalk` and `GraphStep` can be diagrammed by treating the labels as the intended element IDs and the `predicate` label as a request to substitute the following Go condition/loop header. Diagram nodes should not invent operation descriptions beyond annotation content; repeated IDs at distinct source sites need distinct diagram IDs while retaining the same annotation label.
- The control-flow extraction was represented as two graphs:
  - `GraphWalk`: `EachInput` (`for _, i := range input`) → `AddItem` → `Step` → `InsertItem` → `InsertItemErrorCheck` (`if e != nil`); the error-true edge goes to `InsertItemError` (`panic`), and the non-error edge returns to `EachInput`.
  - `GraphStep`: `TreeQuery` → `TreeQuery#MatchCheck` (`if matched`) → `EachMatch` (`for _, match := range matches`) → `MatchBoundCheck` (`if item.Price.Base < match.Price.Bound`). The true branch proceeds through delete (`MatchReplace`), `MatchTruncate`, insert (`MatchReplace`), and `MatchReplaceErrorCheck` (`if e != nil`), then either panic or `MatchLink` and the next match. The false branch enters `GraphForkSamePrice` (`for _, predecessor := range match.AllIn`) → `GraphBranchSamePrice`, repeating for the next predecessor before continuing to the next match.
  - Branch and loop exits that return from a function, and unannotated statements such as `fmt.Println(set)`, were intentionally omitted as diagram nodes. Edges connect only annotated nodes; predicate text is taken from the immediately following source line where the annotation requests `predicate`.

## November 1: Graph generation and symbolic proof strategy

- The developer requested a reusable Mermaid graph artifact for the annotated `dag.go` source, then refined its presentation:
  - Every function return or panic should have an explicit terminal node; panic nodes should have a different shape.
  - Unannotated Go source lines must be tacit, represented only by arrows rather than diagram elements.
  - The BNF added a `call` label. A call node should refer to a callee graph generated from the callee's annotations. Predicates should be visibly labeled `predicate: <source condition>`.
  - A combined graph was first produced with nested call expansions. The developer requested separate graphs per function, with a top label identifying each graph and call-site nodes referring to that label.
  - Extracted reusable conventions into `model/cfg-generation-rules.md`; updated `model/dag.cfg.md` to contain separate Mermaid graphs for `PriceSpanSet.GraphWalk`, `PriceSpanSet.NewItem`, `PriceSpanSet.GraphStep`, summary-described `interval.SearchTree.AllIntersections`, and `Item.Link`.
  - Call-site edges now refer to callee graph labels instead of inlining the callee. This helps keep each function's control flow readable and allows repeated call sites to share one function graph.
- The developer asked whether this graph was enough to prove by symbolic execution that `{Item.in, Item.out}` is acyclic, using `Item.Link`'s invariant as an intermediate lemma. The analysis identified the core mathematical argument:
  - If each `Link(from, to)` call requires `to.out` empty and `from != to`, then adding only `from → to` cannot create a cycle: a cycle using the new edge would require a path out of `to`, but it has no outgoing edges.
  - To establish that precondition at `GraphStep` call sites, the proof also needs the new `item` to be fresh, the search-tree matches to be previously inserted items, and predecessors to come from existing incoming edges.
  - An induction over `GraphWalk` inputs then gives the global property: an initially empty graph is acyclic, and each processed item can only receive incoming edges from older nodes while retaining empty outgoing edges until graph processing is complete.
  - The CFG is an outline only; symbolic execution additionally needs operation semantics, alias/freshness reasoning, and library contracts. The graph alone does not encode the field mutations performed by `Item.Link`.
- The developer asked how to constrain Go evaluation to avoid intractable analysis. Created `model/proof-generation.md`, specifying:
  - Three terminal outcomes: `PROVED`, `REFUTED`, and `UNKNOWN`. Limits, unsupported semantics, missing contracts, and solver failures must never be treated as proof.
  - A deliberately small Go subset, explicit finite limits for call depth, loop unrolling, expression depth, symbolic steps, path count, total steps, wall-clock time, and solver timeouts.
  - Symbolic-state contents, rules for assignments, branch splitting, calls and summaries, induction-based loop proofs, return/postcondition checks, and panic paths.
  - The acyclicity proof plan as obligations: establish `Item.Link`'s call precondition, prove the local edge-addition lemma, then prove `GraphWalk` by induction.
- Next, the developer requested a structured but human-readable acyclicity proof that can be revisited after code changes. Created `model/dag.proof-acyclic.md` as a proof ledger, keyed to annotation tags and source expressions rather than line numbers. It includes:
  - The theorem and graph model (`V`, `E`, and `Acyclic`).
  - Explicit assumptions A1–A8 for initial tree state, vector behavior, tree query membership, predecessor iteration, item freshness, exclusive graph mutation, and the intended precondition meaning of `Item.Link`'s invariant annotation.
  - A dependency map and individual obligations for the initial state, `Link`'s updates, cycle preservation, both call sites, `GraphStep`, and the `GraphWalk` induction.
  - A revalidation checklist and a conditional result statement.
  - It is intentionally not claimed as machine-checked or unconditional: external collection/tree behavior and freshness require validation. The proof specifically calls out that `AllIntersections`'s current summary does not explicitly state that results are stored item pointers.
- In the most recent design discussion, the developer elected to edit annotations manually. Suggestions (not yet implemented) were:
  - Replace ambiguous invariant-only interpretation for `Item.Link` with explicit contract roles such as `requires` and `ensures`, and express both the empty-target/non-alias precondition and edge-update postcondition.
  - Specify the abstract tree property `InTree(tree, x) => EverInserted(tree, x)`, where `EverInserted` is a ghost/history predicate, and pair it with contracts stating that successful `Insert` records provenance and `AllIntersections` returns only current members.
  - Because universal quantification and abstract predicates are not Go expressions, extend the BNF with a small logical-expression grammar and contract productions rather than forcing those facts into `<go>`.
- For readers learning agentic symbolic execution: the central workflow developed here is **annotate behavior → derive per-function CFGs → formulate contracts and invariants → discharge local call obligations → prove loop preservation by induction → preserve uncertainty explicitly**. The CFG supplies path structure, while annotations and source-backed contracts supply semantics. Resource-bounded execution helps discover paths and check local obligations, but unbounded loop properties require inductive arguments; finite random testing or finite unrolling cannot establish them.
