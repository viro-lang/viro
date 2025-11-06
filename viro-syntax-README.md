# Viro Syntax Highlighting for Neovim

## Installation

### Option 1: Manual Installation

1. Copy `viro.vim` to your Neovim syntax directory:
   ```bash
   mkdir -p ~/.config/nvim/syntax
   cp viro.vim ~/.config/nvim/syntax/
   ```

2. Add filetype detection to your Neovim config (`~/.config/nvim/init.vim` or `~/.config/nvim/init.lua`):

   **For init.vim:**
   ```vim
   au BufRead,BufNewFile *.viro set filetype=viro
   ```

   **For init.lua:**
   ```lua
   vim.api.nvim_create_autocmd({"BufRead", "BufNewFile"}, {
     pattern = "*.viro",
     command = "set filetype=viro",
   })
   ```

### Option 2: Put in Project Directory

You can also keep the syntax file in your project and load it manually:
```vim
:set syntax=viro
:set runtimepath+=/path/to/viro/repo
```

## What Gets Highlighted

This syntax file focuses on **structural elements** rather than keywords:

- **Comments** (`;...`) - Bright yellow (`#f9e2af`) - bold and highly visible
- **Strings** (`"..."`) - Soft green (`#a6e3a1`) 
- **Set-word** (`name:`) - Blue (`#89b4fa`) - marks variable assignments
- **Set-path** (`obj.field:`) - Blue (`#89b4fa`) - marks nested assignments
- **Get-word** (`:name`) - Mauve (`#cba6f7`) - marks explicit value retrieval
- **Get-path** (`:obj.field`) - Mauve (`#cba6f7`) - marks nested value retrieval
- **Brackets** (`[]`) - Peach (`#fab387`) - block boundaries
- **Parens** (`()`) - Sky cyan (`#89dceb`) - immediate evaluation

All colors use the [Catppuccin Mocha](https://catppuccin.com/palette/) palette for harmonious, easy-on-the-eyes syntax highlighting.

## Customization

### Color Scheme

The syntax file uses **Catppuccin Mocha** colors with explicit definitions:

- **Comments**: Yellow (`#f9e2af`) - bold
- **Strings**: Green (`#a6e3a1`)
- **Set-word/Set-path**: Blue (`#89b4fa`)
- **Get-word/Get-path**: Mauve (`#cba6f7`)
- **Brackets**: Peach (`#fab387`)
- **Parens**: Sky (`#89dceb`)

### Override Colors

If you want to customize the colors, add this to your Neovim config **after** the syntax file loads:

```vim
" Change comments to a different Catppuccin color
hi viroComment ctermfg=Cyan guifg=#89dceb gui=bold

" Use different Catppuccin colors for assignments
hi viroSetWord ctermfg=Magenta guifg=#cba6f7
hi viroSetPath ctermfg=Magenta guifg=#cba6f7

" Change brackets to teal
hi viroBrackets ctermfg=Cyan guifg=#94e2d5

" Change parens to lavender
hi viroParens ctermfg=Blue guifg=#b4befe
```

You can choose from any [Catppuccin Mocha colors](https://catppuccin.com/palette/).

## Philosophy

This highlighting strategy is optimized for **quick visual scanning**:

1. **Comments** are bright → you notice documentation/notes immediately
2. **Assignments** (set-word/set-path) → you see where values are being bound
3. **Explicit gets** (get-word/get-path) → you see where values are explicitly fetched
4. **Brackets/Parens** → you see code structure and evaluation boundaries
5. **Strings** → you see literal data

This avoids the "rainbow soup" problem where every function name gets colored, which doesn't help in a homoiconic language where everything is just a function.

## Testing

Test your syntax highlighting with this sample:

```viro
; This is a comment - should be bright/visible
x: 42                    ; set-word assignment
name: "Alice"            ; string value
obj.field: 100           ; set-path assignment

result: :x               ; get-word retrieval
value: :obj.field        ; get-path retrieval

block: [1 2 3]           ; brackets highlight block structure
calc: (+ 2 3)            ; parens highlight immediate evaluation
```

## Troubleshooting

**Syntax highlighting not working?**

1. Check filetype: `:set filetype?` (should show `viro`)
2. Check syntax is loaded: `:echo b:current_syntax` (should show `viro`)
3. Manually set: `:set filetype=viro`
4. Check runtime path includes syntax directory: `:set runtimepath?`

**Colors look wrong?**

- The syntax file sets explicit Catppuccin Mocha colors that should work with any theme
- If you want different colors, add custom highlights to your config file
- Use `:hi viroComment` to see current highlighting rules

**Very long strings (7000+ chars) not highlighting?**

- This is fixed with `synmaxcol=0` in the syntax file
- If you experience performance issues, you can set a limit: `:set synmaxcol=10000`

## Future Enhancements

Possible additions (if needed):

- Refinements (`--add`, `--multiply`)
- Numbers (integers, decimals, scientific notation)
- Special words (`true`, `false`, `none`)
- Object literal syntax (`#[]`)
- Path expressions without set/get prefixes

Let me know if you'd like any of these added!
