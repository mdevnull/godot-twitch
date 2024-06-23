extends Node2D

class_name EventBase

@export var anim_process: float = 0.0

func _ready():
  var vp_width = get_viewport().get_visible_rect().size.x
  position.x = -vp_width

func _process(_delta):
  var vp_width = get_viewport().get_visible_rect().size.x 
  var current_x = -vp_width + ((vp_width * 2) * anim_process)
  position.x = current_x