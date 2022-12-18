Output demo/ijq.gif
Set Height 1080
Set Width 1920

Hide
Type "make ijq && clear"
Enter
Type "curl https://dummyjson.com/comments | ./ijq"
Enter
Sleep 2
Show

Sleep 2

Type@0.4 "comments[] | select(.user.id < 60)"

Sleep 2

Type@0.4 " | {id, body}"

Sleep 3

Up

Sleep 1

Down @0.2 10

Sleep 1

Tab

Sleep 1

Down@0.2 20

Sleep 2

Up@0.2 10

Sleep 6
