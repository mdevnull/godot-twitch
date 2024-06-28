# Methods and properties

## Methods

|Name|Description|
|---|---|
|OpenAuthInBrowser|Tries to open the twitch OAuth login site in the users default browser|

## Properties

|Name|Type|Description|
|---|---|---|
|auth_url|String|URI to open to autheenticate with twitch|
|store_token|bool|If true tries to load tokens from disk and stores new tokens to disk|
|is_authed|bool|True if client has been authenticated. This can be true when _ready if store_token is true and valid tokens are stored on disk|
|latest_follower|String|Username of latest follower|
|latest_subscriber|String|Username of latest subscriber|
