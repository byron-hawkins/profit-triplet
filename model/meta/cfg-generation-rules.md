# Control-flow graph generation rules

These conventions apply to control-flow diagrams generated from Go source files annotated according to `symbolic-abstraction.bnf`.

1. Create a separate flow graph for each annotated function. Put the function's name in a heading immediately above its Mermaid graph. Give each node a unique diagram ID within its graph, independent of the annotation's element ID.
2. Use the annotation tag and its element content as the node label. Do not add descriptions of the Go statement.
3. If an annotation includes `predicate`, replace that label with the following source condition or loop header, formatted as `predicate: <source text>`. Retain other annotation labels, such as `loophead`.
4. An annotation containing `call` is a call-site node. Label it with its annotation tag and the name of the callee's graph (for example, `GraphWalk#AddItem: call → Item.NewItem`). Do not inline the callee graph or splice its nodes into the caller's graph. The named graph is the reference for the call's behavior. Multiple call sites to the same function refer to the same function graph.
5. Add a synthetic `exit` node for each normal function return, including return due to loop exhaustion or a false guard that skips the remaining body. Use the annotation text `panic` for a panic node and draw it with a distinct shape; panics do not connect to normal exits.
6. Represent branch outcomes and loop continuation/exhaustion on edges. Edge labels are control-flow metadata, not diagram elements.
7. When a callee is described by a summary annotation rather than a body in the source, create a separate graph named for that callee and use the summary alias and substitutions as its contents, with separate outgoing paths for the described substitution cases.
8. Omit unannotated source lines as diagram nodes. Treat them as tacit parts of the arrows between annotated nodes.
