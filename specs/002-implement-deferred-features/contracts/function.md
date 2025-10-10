# Function Contracts (Eval/EvalArgs)

## Specyfikacja kontroli ewaluacji argumentów funkcji

### 1. Parametry funkcji użytkownika

Każdy parametr w specyfikacji funkcji użytkownika (`fn [...] [...]`) może mieć flagę `Eval`:
- `Eval=true` (domyślnie): argument jest ewaluowany przed przekazaniem do funkcji
- `Eval=false`: argument przekazywany "raw" (np. blok, lit-word, refinement)

#### Przykład:
```rebol
fn [a 'flag b] [ ... ]
```
- `a` → ewaluowany
- `'flag` → nieewaluowany (lit-word)
- `b` → ewaluowany

### 2. Parametry funkcji natywnych

Natywne funkcje mogą określić tablicę `EvalArgs`:
- `EvalArgs[i]=true` → argument ewaluowany
- `EvalArgs[i]=false` → argument przekazywany raw

#### Przykład:
```go
registerNative("when", NativeInfo{
  ...
  EvalArgs: []bool{true, false}, // condition ewaluowany, block raw
})
```

### 3. Refinements

Refinements (`/flag`, `/option`) mogą być przekazywane jako lit-wordy (`'flag`) z `Eval=false`.
- Błąd, jeśli refinement jest lit-wordem i nie jest ostatni w specyfikacji.

### 4. Testy kontraktowe

- Funkcje muszą być testowane na:
  - Przekazywanie bloków, lit-wordów, refinements bez ewaluacji
  - Błędy typów/refinements
  - Kompatybilność wsteczną

### 5. Przykłady

#### Funkcja użytkownika z blokiem raw
```rebol
fn [block] [ ... ]        ; domyślnie ewaluowany
fn ['block] [ ... ]       ; przekazywany raw
```

#### Funkcja natywna z blokiem raw
```go
registerNative("if", NativeInfo{
  ...
  EvalArgs: []bool{true, false, false}, // condition ewaluowany, oba bloki raw
})
```

#### Refinement
```rebol
fn [a 'refinement] [ ... ]
```

### 6. Kompatybilność

- Stare funkcje bez `Eval`/`EvalArgs` zachowują się jak wcześniej (wszystkie argumenty ewaluowane).
- Nowe funkcje mogą precyzyjnie kontrolować ewaluację każdego argumentu.

### 7. Migracja

- Dodaj pole `Eval` do specyfikacji parametrów funkcji użytkownika.
- Dodaj `EvalArgs` do rejestracji natywnych funkcji.
- Przetestuj przypadki edge-case (blok, lit-word, refinement, typy).

---

Więcej przykładów i testów: `test/contract/function_eval_test.go`, `docs/repl-usage.md`, `MIGRATION.md`.
