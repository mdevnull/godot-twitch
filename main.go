package main

import (
	"grow.graphics/gd"
	"grow.graphics/gd/gdextension"

	"main/node"
)

func main() {
	godot, ok := gdextension.Link()
	if !ok {
		panic("Unable to link to godot")
	}
	gd.Register[node.GodotTwitch](godot)
}
