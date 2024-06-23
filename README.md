# Twitch Godot

![image](https://github.com/devnull-twitch/godot-twitch/assets/97809736/f0a94561-c8c5-4d71-8063-fe4b0e569cdc)

This project is somewhere between early alpha and abandoned. 

## Setup and build

Follow the installation for [grow-graphics/gd](https://github.com/grow-graphics/gd) and you should be able to simply run `gd` to build and start the project. This repo comes with a proof of concept godot project but I think the resulting `graphics/library.gdextension` and corresponding DLL, SO or dylib file should work if copied over to any other godot project.

## Testing Webserver backend

You can `go run util/ws_mockserver/main.go` to start a CLI application that starts a websocket server. If you set the "use_debug_ws_server" flag on the godot node it will connect to the local websocket server and you can teest a bunch of events. Note that you will still need a valid client ID and secret for a twitch app and run through the auth process.\
That is mainly because some events will trigger additional API calls. 