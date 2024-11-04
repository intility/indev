Set Shell zsh
Set Theme {  "name": "custom",  "black": "#26233a",  "red": "#eb6f92",  "green": "#9ccfd8",  "yellow": "#f6c177",  "blue": "#31748f",  "magenta": "#c4a7e7",  "cyan": "#ebbcba",  "white": "#e0def4",  "brightBlack": "#6e6a86",  "brightRed": "#eb6f92",  "brightGreen": "#9ccfd8",  "brightYellow": "#f6c177",  "brightBlue": "#31748f",  "brightMagenta": "#c4a7e7",  "brightCyan": "#ebbcba",  "brightWhite": "#e0def4",  "background": "#071627",  "foreground": "#e0def4",  "cursor": "#e0def4",  "selection": "#191724" }
Set FontSize 14
Set Width 750
Set Height 500
Set Padding 20
Set WindowBar Colorful
Set TypingSpeed 0.1
Sleep 2s
Type "indev -h"
Enter
Sleep 3s
Type "indev cluster create -n my-cluster --nodes 2 --preset minimal"
Enter
Sleep 2s
Type "indev cluster list"
Enter
Sleep 6s
