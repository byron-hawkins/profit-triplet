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

For example, `<abstract-model-name>` can be instantiated as `PriceBounds`
or `InputSizeBounds`.

Freshness is a contextual constraint, not a property ordinary BNF can enforce:

- When a name is introduced as a declaration, its spelling must be fresh in
  the namespace and scope of that declaration. It must not collide with a
  different declaration already in scope.
- When a name occurs as a reference, it resolves to the matching declaration;
  it is not a new fresh choice. For example, an abstract-model name in an
  invariant or alias refers to the abstract-model declaration with that name.
- Two references to the same declaration must use the same spelling.
- Name freshness does not mean that every occurrence must have a different
  spelling. It applies when a declaration introduces a name; references
  intentionally reuse that spelling.

The applicable namespace and scope are determined by the construct that
introduces the name. Unless a more specific rule says otherwise, an
abstract-model name is unique among abstract-model declarations in the
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
associates it with its body. An occurrence of that name in an invariant
definition or model alias refers to the declaration; it does not introduce
another declaration.

Parsing still requires the source annotation to conform to the productions
in `symbolic-abstraction.bnf`. The naming rules do not relax syntactic
requirements or make a malformed annotation valid.
