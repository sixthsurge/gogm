module gogm/playbot

go 1.22.5

replace gogm/chess => ../chess

replace gogm/chessgui => ../chessgui

replace gogm/botv1 => ../botv1

require (
	gogm/botv1 v0.0.0-00010101000000-000000000000
	gogm/chessgui v0.0.0-00010101000000-000000000000
)

require (
	github.com/veandco/go-sdl2 v0.4.40 // indirect
	gogm/chess v0.0.0-00010101000000-000000000000 // indirect
)
