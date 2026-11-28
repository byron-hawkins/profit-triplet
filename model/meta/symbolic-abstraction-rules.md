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

## Abstract collection annotation placement

An `<abstract-collection-annotation>` (`@=`) must appear immediately before
the Go type declaration to which it applies. It associates the declared
abstract collection name with that following type.

An `<abstract-collection-model>` (`@==`) must appear immediately after
the struct field declaration to which it applies. When it contains an
`<abstract-collection-instantiation>`, that qualifier associates the
named abstract collection with the field.

If the instantiation names a data model, resolve the `<data-model-name>` by
exact spelling against the data-model declarations for the annotated
field's type. The reference denotes that modeled member for the particular
field instance, not a Go field with the same name. A missing or non-unique
match is a specification error. If no data-model name is supplied, the
instantiation refers directly to the annotated field.

These are source-placement constraints: both forms are annotations in the
BNF, but their required Go source contexts are enforced by these rules.

## Field mutation policies

An `<abstract-immutable>` qualifier associates an immutable value property
with each instance of the annotated field. Its initial value is established
when the containing struct instance is initialized, including initialization
to the zero value. Thereafter, the field's value must not change.

An `<abstract-interface>` qualifier associates a partial immutability policy
with each instance of the annotated field. Let `I` be its named interface.
Mutation of the field is permitted only during an invocation of a method
defined by `I`, including synchronous calls made during that invocation.
The permission is not restricted to the receiver's own field instance.
Any other attempt to mutate the field fails analysis with an interface
policy violation.

In analyzed code, a call to a method defined by `I` on a receiver whose
static type implements `I` is governed by this policy. The method must be
selected through `I`. If the receiver expression has a different static
type, convert it to `I` before selecting the method, for example
`I(foo).M()`. A direct call such as `foo.M()`, or a call through a method
value selected as `foo.M`, is an interface policy violation, even if a
surrounding invocation would otherwise permit the mutation. An expression
already typed as `I` may select `M` directly. The conversion and call must
be valid Go.

These policies cover the field and its owned modeled contents, including
changes made through aliases or by replacing an enclosing value. For example,
the contents of a tree referenced by an annotated field are covered, but
independent objects merely referenced by entries in the tree are not thereby
covered. Initializing a new struct instance establishes its initial state;
subsequent changes to an existing instance remain subject to its policies.
An abstract collection state update does not itself grant permission to
mutate the corresponding field or its owned contents.

A known policy violation invalidates policy-dependent analysis. If compliance
cannot be established, including because aliasing or ownership cannot be
resolved, report `UNKNOWN` rather than assuming compliance.

## Abstract collection semantics

An abstract collection (AC) is declared with a kind (`set` or `list`), an
element name, and a Go element type. Its state is associated with each
instance of the Go type annotated by `@=`. Unless a corresponding abstract
collection instantiation (ACI) supplies an initial value, the AC starts
empty. An ACI supplies the initial abstract state by interpreting the initial
value of its referenced field or modeled member according to the AC kind.

An ACI associates its referenced field or modeled member with the named AC
and states that its contents represent the AC. For a set AC, correspondence
compares membership and ignores order and duplicates. For a list AC,
correspondence compares the sequence, including order and duplicates. An
ACI is optional; an AC without an ACI is purely abstract.

A state annotation updates the named AC as if by assignment:

```text
AC = AC operator operand
```

Classify each operand from its statically resolved Go type; do not infer its
role from runtime behavior:

- If the operand's type is compatible with the AC element type, interpret it
  as one element, i.e. a singleton operand.
- Otherwise, an operand may be interpreted as a collection only when its
  static type is a supported array, slice, or vector type and its element
  type is compatible with the AC element type. A vector type is supported
  only when its element type and ordered contents are explicitly available
  from its declaration or declared model.
- Do not interpret maps, structs, interfaces, iterators, or other aggregate
  types as collection operands.
- If the operand's type or element type cannot be statically resolved well
  enough to classify it, reject it as insufficiently defined. If its type is
  fully resolved but is not a compatible element or supported collection,
  reject it as incompatible. Do not silently coerce either case.

For set ACs, `∪`, `∩`, and `\` use ordinary set union, intersection, and
difference. Duplicates are discarded and order is irrelevant. Set elements
must have a defined equality relation; for this Go-oriented semantics,
require a comparable element type and use Go `==`.

For list ACs:

- `∪` appends the operand sequence to the current list; a singleton operand
  appends one element.
- `∩` retains the longest common prefix of the current list and operand
  sequence. A singleton operand is a one-element sequence.
- `\` removes every occurrence in the current list of each element in the
  operand sequence; a singleton operand removes every occurrence of that
  element. For example, `[1,2,1,3,1,4,1,5] \ [3,1,3]` is `[2,4,5]`.

List element equality is Go `==` under the same comparability requirement.
The source expression's slice or vector order and multiplicity are
preserved when forming its operand sequence.

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

Each argument is a `<go-expression>` and must meet the restrictions in the
BNF. Its identifiers must resolve in the scope of the invocation; the
expression must have no side effects or variable assignments. The complete
argument list must be valid for the model's declared parameters under Go
function-call rules, including the rules for multiple-result calls.

For each incoming value flow, prove that every pointer dereferenced by a
selector is non-nil. If this cannot be established, stop analysis with an
`unknown value flow` error. Every subscripted variable must resolve to the
same base collection kind on all value-flow paths: array, slice, or map.
For each path, establish that array and slice subscripts are in bounds;
array size and slice length may differ between paths. Slice bounds use
length, not capacity. A missing map key, including a lookup in a nil map,
produces the map value type's zero value, as in Go.

Do not silently assume any of these properties. Any resolution failure
stops analysis with an `unknown value flow` error; report `UNKNOWN` when a
proof outcome is required. An expression known to violate a restriction is
not a valid model argument. The same collection kind does not exempt the
expression from Go's type-checking rules.

A call evaluates the model body as a function returning `bool`, with
arguments evaluated and bound to parameters according to Go function-call
semantics. The body is interpreted in the applicable modeled state;
pointer and other reference values retain their ordinary Go behavior.
An otherwise valid expression that is outside the supported symbolic
analysis subset yields `UNKNOWN` if it affects the proof; do not execute
arbitrary project code to evaluate it.

Before making proof progress, statically inspect the abstract-model call
graph for the source file. Each model is a vertex, and every
`<abstract-model-call>` in its body is an edge to the referenced model,
regardless of whether the call is reachable during evaluation. If a
self-call or any other cycle is discovered, report a specification error
and stop immediately; make no further proof progress or proof claim. If a
proof outcome is required, report `UNKNOWN`, not `REFUTED`. If the graph
cannot be checked, do not assume it is acyclic or proceed with the proof.

## Loop-head conditions

For an annotation labeled `loophead`, evaluate its preconditions once per
loop execution, before entering the loop, not again on each iteration.
Evaluate its postconditions once at each loop exit, including an exit
without executing the body and exits via `break`, `return`, or panic.
A `continue` that remains within the loop is not an exit. A loop that does
not exit has no exit evaluation.

These are entry and exit conditions, not implicitly inductive loop
invariants. Proving an exit condition may require a separate invariant
established at entry and preserved on every iteration.

Parsing still requires the source annotation to conform to the productions
in `symbolic-abstraction.bnf`. The naming rules do not relax syntactic
requirements or make a malformed annotation valid.
