extends CanvasLayer

@onready var twitch_node: GodotTwitch = %GodotTwitch

func _on_open_auh():
	twitch_node.OpenAuthInBrowser()
