# Quickstart: Parse Dialect

## 1. Basic string validation

```
data: "person@example.org"
parse data [copy local to "@" skip copy domain to end]
```
- Returns `true` and binds `local` = "person", `domain` = "example.org".
- Add `--case` to enforce case-sensitive comparisons.

## 2. Alternation and repetition

```
date?: fn [value] [
    parse value [
        4 digit "-" 2 digit "-" 2 digit
    ]
]
```
- `digit` is a charset rule defined via `digit: charset ["0123456789"]`.
- `some`/`any` provide open-ended repetition: `parse value [some digit]`.

## 3. Block DSL parsing

```
rule: [
    some [set word word! set val integer! | into rule]
]
script: [foo 10 bar 20 [baz 30]]
parse script rule
```
- Ensures every word is followed by an integer or nested rule.
- Recursive `into rule` handles nested blocks.

## 4. Capturing and collecting

```
data: "a=1&b=2&c=3"
pairs: make block! 0
parse data [
    collect [
        some [
            copy key to "=" skip copy val to ["&" | end]
            keep reduce [key val]
            opt skip
        ]
    ]
]
```
- `collect/keep` accumulates `["a" "1" "b" "2" "c" "3"]` without mutating `data`.

## 5. Mutating while parsing

```
parse data [some ["foo" change "foo" "bar"]]
```
- Rewrites every `foo` to `bar`. Copy-on-write protects other references.

## 6. Debugging

- Enable tracing with `trace --on` ... `trace --off` around `parse` to inspect rule progression.
- Syntax errors raise `Script error: Syntax (200)` with `near` pointing at the failing input segment.