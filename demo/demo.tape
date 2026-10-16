Output demo/demo.gif

Require go
Require zoxide
Require zellij
Require nvim

Set Shell bash
Set Theme "TokyoNight"
Set FontSize 16
Set Width 1200
Set Height 780
Set Padding 20
Set TypingSpeed 90ms

Hide
Type "source demo/setup.sh" Enter
Sleep 3s
Show

Type "tobu" Sleep 400ms Enter
Sleep 1.5s
Down Sleep 400ms
Down Sleep 400ms
Up Sleep 600ms
Type "web" Sleep 1s
Enter
Sleep 3s

Ctrl+b Type "d"
Sleep 1s

Type "tobu" Sleep 400ms Enter
Sleep 2s
Escape
Sleep 500ms

Type "tobu infra" Sleep 400ms Enter
Sleep 3s

Hide
Ctrl+b Type "d"
Sleep 500ms
Type "for s in api web infra; do zellij kill-session $s; done; rm -rf /tmp/tobu-demo" Enter
Sleep 1s
