package lib

import (
	"fmt"

	"grow.graphics/gd"
)

func LogWarn(godoCtx gd.Context, msg string) {
	fullMsg := fmt.Sprintf("[color=orange][b]GodotTwitch Warning[/b]: %s[/color]", msg)
	godoCtx.PrintRich(godoCtx.Variant(godoCtx.String(fullMsg)))
}

func LogErr(godoCtx gd.Context, msg string) {
	fullMsg := fmt.Sprintf("[color=red][b]GodotTwitch Error[/b]: %s[/color]", msg)
	godoCtx.PrintRich(godoCtx.Variant(godoCtx.String(fullMsg)))
}

func LogInfo(godoCtx gd.Context, msg string) {
	fullMsg := fmt.Sprintf("[color=purple][b]GodotTwitch[/b]: %s[/color]", msg)
	godoCtx.PrintRich(godoCtx.Variant(godoCtx.String(fullMsg)))
}
