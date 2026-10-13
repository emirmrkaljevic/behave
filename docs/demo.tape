# Regenerate with: vhs docs/demo.tape  (needs behave, jq and gojq on PATH)
Output docs/demo.gif
Set Shell "bash"
Set FontSize 15
Set Width 1200
Set Height 720
Set Padding 16
Set Theme "Catppuccin Mocha"
Set TypingSpeed 35ms

Hide
Type "unset NO_COLOR; cd $(mktemp -d) && export PS1='$ ' && clear"
Enter
Show

Type "# 1. record how jq behaves"
Enter
Type "behave probe jq --arg . --arg .a --arg keys --arg 'map(.+1)' --arg '.[]'"
Sleep 300ms
Enter
Sleep 2s
Type "# 2. check gojq (a Go rewrite of jq) against it"
Enter
Type "behave diff behave-jq --impl gojq --ignore-stderr --skip-tag meta --max-failures 1"
Sleep 300ms
Enter
Sleep 7s
