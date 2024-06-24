extends AnimationPlayer

@onready var twitch_node: GodotTwitch = %GodotTwitch
@onready var follow_node: EventBase = get_node("../FollowEvent")
@onready var sub_node: EventBase = get_node("../SubEvent")
@onready var gift_node: EventBase = get_node("../GiftEvent")
@onready var raid_node: Node2D = get_node("../RaidEvent")

func _ready():
  twitch_node.on_follow.connect(_on_godot_twitch_follow)
  twitch_node.on_subscribtion.connect(_on_godot_twitch_subscribtion)
  twitch_node.on_sub_gift.connect(_on_godot_twitch_giftsub)
  twitch_node.on_raid.connect(_on_godot_twitch_raid)
  twitch_node.on_poll_begin.connect(_on_godot_twitch_poll_begin)
  twitch_node.on_poll_end.connect(_on_godot_twitch_poll_end)

func _on_godot_twitch_follow(username: String):
  var name_label = follow_node.get_node("ContentContainer/HBoxContainer/NameLabel") as Label
  name_label.text = username
  queue("run_follow")

func _on_godot_twitch_subscribtion(username: String, months: int, tier: int):
  var name_label = sub_node.get_node("ContentContainer/HBoxContainer/NameLabel") as Label
  var months_label = sub_node.get_node("ContentContainer/HBoxContainer/MonthsLabel") as Label
  var tier_label = sub_node.get_node("ContentContainer/HBoxContainer/TierLabel") as Label
  name_label.text = username
  months_label.text = "%d" % months
  tier_label.text = "%d" % tier
  queue("run_sub")

func _on_godot_twitch_giftsub(username: String, amount: int, tier: int, total: int):
  var name_label = sub_node.get_node("ContentContainer/HBoxContainer/NameLabel") as Label
  var gift_amount_label = sub_node.get_node("ContentContainer/HBoxContainer/GiftAmountLabel") as Label
  var tier_label = sub_node.get_node("ContentContainer/HBoxContainer/TierLabel") as Label
  var total_label = sub_node.get_node("ContentContainer/HBoxContainer/TotalAmountLabel") as Label
  name_label.text = username
  gift_amount_label.text = "%d" % amount
  tier_label.text = "%d" % tier
  total_label.text = "%d" % total
  queue("run_gift")

func _on_godot_twitch_raid(username: String, profile_url: String, viewer_count: int):
  var name_label = raid_node.get_node("ContentContainer/VBoxContainer/NameLabel") as Label
  var viewer_count_node = raid_node.get_node("ContentContainer/VBoxContainer/HBoxContainer/ViewerCount") as Label
  var profile_texture_rect = raid_node.get_node("ContentContainer/VBoxContainer/TextureRect") as TextureRect
  name_label.text = username
  viewer_count_node.text = "%d" % viewer_count
  change_img_from_url(profile_url, profile_texture_rect)
  raid_node.visible = true
  await get_tree().create_timer(5.0).timeout
  raid_node.visible = false

func _on_godot_twitch_poll_begin(_title: String, _lockTimeUnix: int, choices: Array[Dictionary]):
  print(choices)

func _on_godot_twitch_poll_end(_title: String, choices: Array[Dictionary]):
  print(choices)

func change_img_from_url(img_url: String, texture_rect: TextureRect):
  # Create an HTTP request node and connect its completion signal.
  var http_request = HTTPRequest.new()
  add_child(http_request)
  http_request.connect("request_completed", self._http_request_completed.bind(texture_rect))

  # Perform the HTTP request. The URL below returns a PNG image as of writing.
  var error = http_request.request(img_url)
  if error != OK:
    push_error("An error occurred in the HTTP request.")

# Called when the HTTP request is completed.
func _http_request_completed(result, _response_code, _headers, body, texture_rect):
  if result != HTTPRequest.RESULT_SUCCESS:
    push_error("Image couldn't be downloaded. Try a different image.")

  var image = Image.new()
  var error = image.load_png_from_buffer(body)
  if error != OK:
    push_error("Couldn't load the image.")
    return

  var texture = ImageTexture.create_from_image(image)

  texture_rect.texture = texture
