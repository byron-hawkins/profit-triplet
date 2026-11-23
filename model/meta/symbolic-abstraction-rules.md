# Symbolic-abstraction naming and semantic rules

This document supplements
[`symbolic-abstraction.bnf`](./symbolic-abstraction.bnf). The BNF describes
the annotation's surface syntax; the rules here define naming conventions
and constraints that ordinary BNF cannot express.

## Name nonterminals and freshness

A nonterminal whose name ends in `-name` denotes an identifier token whose
spelling is freely chosen by the author, subject to the identifier syntax of
the language or notation in which it is used. It is not a literal string
equal to the nonterminal's name.

For example, `<abstraction-name>` can be instantiated as `PriceBounds`
or `InputSizeBounds`.

Freshness is a contextual constraint, not a property ordinary BNF can enforce:

- When a name is introduced as a declaration, its spelling must be fresh in
  the namespace and scope of that declaration. It must not collide with a
  different declaration already in scope.
- When a name occurs as a reference, it resolves to the matching declaration;
  it is not a new fresh choice. For example, an abstraction name in an
  invariant or alias refers to the abstract-model declaration with that name.
- Two references to the same declaration must use the same spelling.
- Name freshness does not mean that every occurrence must have a different
  spelling. It applies when a declaration introduces a name; references
  intentionally reuse that spelling.

The applicable namespace and scope are determined by the construct that
introduces the name. Unless a more specific rule says otherwise, an
abstraction name is unique among abstract-model declarations in the
annotation set being parsed.

## Why this is not encoded in the BNF

BNF can describe the shape of an identifier token, but it cannot by itself
express that a declaration's identifier has not appeared before, or that a
later reference must equal the spelling of a particular earlier
declaration. Those are contextual, cross-occurrence constraints. A grammar
tool may enforce them with semantic actions, symbol-table checks, or an
attribute grammar, but they are specified here rather than encoded as
ordinary productions.

## Abstract-model declaration and reference

An abstract-model declaration introduces a fresh abstract-model name and
associates it with its body. A reference resolves to an abstract-model
declaration in the same Go source file as the annotation containing the
reference; declarations may appear before or after their references.

A standalone model reference in an invariant definition or model alias uses
`<abstraction-invocation>`, without an `@` prefix. Within `<go>` or `<go+>`,
a model reference must instead use `<abstract-model-call>`, with an `@`
prefix, in an expression position that accepts a `bool`. For example:

```go
if @TwinChainSamePrice(i) {
    return true
}
```

An unprefixed Go identifier or function call does not reference an abstract
model. An `@` in a Go string or rune literal is not a model call.

Each argument must be a variable identifier in scope at the call site, as
specified by `<abstraction-invocation>`. The number and types of arguments
must be valid for the model's declared parameters under Go function-call
rules. A call evaluates the model body as a function returning `bool`, with
arguments evaluated and bound to parameters as in a Go function call. The
body is interpreted in the applicable modeled state; pointer and other
reference values retain their ordinary Go behavior.

Before making proof progress, statically inspect the abstract-model call
graph for the source file. Each model is a vertex, and every
`<abstract-model-call>` in its body is an edge to the referenced model,
regardless of whether the call is reachable during evaluation. If a
self-call or any other cycle is discovered, report a specification error
and stop immediately; make no further proof progress or proof claim. If a
proof outcome is required, report `UNKNOWN`, not `REFUTED`. If the graph
cannot be checked, do not assume it is acyclic or proceed with the proof.

Parsing still requires the source annotation to conform to the productions
in `symbolic-abstraction.bnf`. The naming rules do not relax syntactic
requirements or make a malformed annotation valid.
