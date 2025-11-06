if exists("b:current_syntax")
  finish
endif

syntax clear

setlocal synmaxcol=0

syn match viroComment ";.*$"

syn region viroString start=+"+ skip=+\\\\\|\\"+ end=+"+ oneline keepend

syn match viroSetWord "\<[a-zA-Z_][a-zA-Z0-9_?!-]*:"
syn match viroSetPath "\<[a-zA-Z_][a-zA-Z0-9_?!-]*\(\.[a-zA-Z_][a-zA-Z0-9_?!-]*\)\+:"

syn match viroGetWord ":[a-zA-Z_][a-zA-Z0-9_?!-]*\>"
syn match viroGetPath ":[a-zA-Z_][a-zA-Z0-9_?!-]*\(\.[a-zA-Z_][a-zA-Z0-9_?!-]*\)\+"

syn match viroBrackets "[\[\]]"
syn match viroParens "[()]"

syn match viroNumber "\<-\?\d\+\>"
syn match viroNumber "\<-\?\d\+\.\d\+\>"
syn match viroNumber "\<-\?\d\+\(\.\d\+\)\?[eE][+-]\?\d\+\>"

hi viroComment ctermfg=Yellow guifg=#f9e2af gui=bold cterm=bold
hi viroString ctermfg=Green guifg=#a6e3a1
hi viroNumber ctermfg=Red guifg=#eba0ac
hi viroSetWord ctermfg=Blue guifg=#89b4fa
hi viroSetPath ctermfg=Blue guifg=#89b4fa
hi viroGetWord ctermfg=Magenta guifg=#cba6f7
hi viroGetPath ctermfg=Magenta guifg=#cba6f7
hi viroBrackets ctermfg=DarkYellow guifg=#fab387
hi viroParens ctermfg=Cyan guifg=#89dceb

syn sync minlines=500
syn sync maxlines=1000

let b:current_syntax = "viro"
