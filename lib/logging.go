package lib

import (
	"fmt"

	"graphics.gd/classdb/Engine"
)

func LogWarn(msg string) {
	fullMsg := fmt.Sprintf("[color=orange][b]GodotTwitch Warning[/b]: %s[/color]", msg)
	Engine.PrintRich(fullMsg)
}

func LogErr(msg string) {
	fullMsg := fmt.Sprintf("[color=red][b]GodotTwitch Error[/b]: %s[/color]", msg)
	Engine.PrintRich(fullMsg)
}

func LogInfo(msg string) {
	fullMsg := fmt.Sprintf("[color=purple][b]GodotTwitch[/b]: %s[/color]", msg)
	Engine.PrintRich(fullMsg)
}
