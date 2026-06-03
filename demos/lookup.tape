Output demos/lookup.gif

Require synthient

Set Shell "zsh"
Set Theme {"name":"Aura","black":"#110f18","red":"#ff6767","green":"#61ffca","yellow":"#ffca85","blue":"#a277ff","magenta":"#a277ff","cyan":"#61ffca","white":"#edecee","brightBlack":"#4d4d4d","brightRed":"#ffca85","brightGreen":"#a277ff","brightYellow":"#ffca85","brightBlue":"#a277ff","brightMagenta":"#a277ff","brightCyan":"#61ffca","brightWhite":"#edecee","background":"#15141b","foreground":"#edecee","cursor":"#a277ff","selection":"#a277ff"}
Set TypingSpeed 28ms
Set FontSize 18
Set Width 1320
Set Height 900
Set Padding 22
Set Framerate 30
Set PlaybackSpeed 1.0
Set LineHeight 1.12
Set WindowBar Colorful

Type "synthient lookup 8.8.8.8"

Enter

Sleep 3s

Type "synthient lookup 8.8.8.8 1.1.1.1 --format csv"

Enter

Sleep 3s

Type "synthient lookup domain example.com"

Enter

Sleep 3s
