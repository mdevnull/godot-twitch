package main

import (
	"graphics.gd/classdb"
	"graphics.gd/startup"

	"main/node"
)

func main() {
	classdb.Register[node.GodotTwitch]()
	classdb.Register[node.GodotTwitchEmoteStore]()
	startup.Engine()
}
