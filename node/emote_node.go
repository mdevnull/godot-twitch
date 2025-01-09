package node

import (
	"fmt"
	"io"
	"main/lib"
	"net/http"
	"sync"

	"graphics.gd/classdb"
	"graphics.gd/classdb/DirAccess"
	"graphics.gd/classdb/FileAccess"
	"graphics.gd/classdb/Image"
	"graphics.gd/classdb/Node"
	"graphics.gd/variant/Float"
	"graphics.gd/variant/Signal"
)

type GodotTwitchEmoteStore struct {
	classdb.Extension[GodotTwitchEmoteStore, Node.Instance] `gd:"GodotTwitchEmoteStore"`

	UseDarkTheme bool `gd:"use_dark_theme"`
	Scale        int  `gd:"scale"`

	OnEmoteReady Signal.Solo[string] `gd:"on_emote_ready"`

	emoteQueueLock       *sync.Mutex
	emoteReadyEventQueue []string
	emoteQueueIndex      int
	emoteImageCache      map[string]Image.Instance
	emoteByteChan        chan emoteByteResponse

	httpClient *http.Client
}

type emoteByteResponse struct {
	content []byte
	emoteID string
}

func (h *GodotTwitchEmoteStore) Ready() {
	h.httpClient = &http.Client{}
	h.emoteReadyEventQueue = make([]string, 0, 10)
	h.emoteQueueLock = &sync.Mutex{}
	h.emoteByteChan = make(chan emoteByteResponse, 1)
	h.emoteImageCache = make(map[string]Image.Instance)
}

func (h *GodotTwitchEmoteStore) Process(delta Float.X) {
	if h.emoteQueueIndex > 0 {
		h.emoteQueueLock.Lock()
		defer h.emoteQueueLock.Unlock()

		for i := h.emoteQueueIndex; i > 0; i-- {
			h.OnEmoteReady.Emit(h.emoteReadyEventQueue[i-1])
		}

		h.emoteQueueIndex = 0
	}

	if len(h.emoteByteChan) > 0 {
		emoteResponse := <-h.emoteByteChan
		if !DirAccess.DirExistsAbsolute("user://emotes") {
			DirAccess.MakeDirAbsolute("user://emotes")
		}

		path := fmt.Sprintf("user://emotes/%s", emoteResponse.emoteID)
		accessWriteFa := FileAccess.Instance(FileAccess.Open(path, FileAccess.Write))
		accessWriteFa.StoreBuffer(emoteResponse.content)

		emoteImage := Image.New()
		emoteImage.LoadPngFromBuffer(emoteResponse.content)

		h.emoteImageCache[emoteResponse.emoteID] = emoteImage
		h.addEventQueue(emoteResponse.emoteID)
	}
}

// LoadEmote makes sure the emote is loaded and available and triggers a "on_emote_ready" after.
func (h *GodotTwitchEmoteStore) LoadEmote(emoteID string) {
	if _, isCached := h.emoteImageCache[emoteID]; isCached {
		h.OnEmoteReady.Emit(emoteID)
		return
	}

	path := fmt.Sprintf("user://emotes/%s", emoteID)

	if FileAccess.FileExists(path) {
		emoteFa := FileAccess.Instance(FileAccess.Open(path, FileAccess.Read))
		emoteBuffer := emoteFa.GetBuffer(emoteFa.GetLength())

		emoteImage := Image.New()
		emoteImage.LoadPngFromBuffer(emoteBuffer)

		h.emoteImageCache[emoteID] = emoteImage
		h.addEventQueue(emoteID)
		return
	}

	// async check on disk and maybe load from web
	go func(emoteIdStr string) {
		theme := "light"
		if h.UseDarkTheme {
			theme = "dark"
		}
		emoteResp, err := h.httpClient.Get(fmt.Sprintf(
			"https://static-cdn.jtvnw.net/emoticons/v2/%s/static/%s/%d.0",
			emoteIdStr,
			theme,
			int64(h.Scale),
		))
		if err != nil {
			lib.LogErr(fmt.Sprintf("error loading emote: %s", err.Error()))
			return
		}

		httpBytes, err := io.ReadAll(emoteResp.Body)
		if err != nil {
			lib.LogErr(fmt.Sprintf("error loading emote: %s", err.Error()))
			return
		}

		h.emoteByteChan <- emoteByteResponse{
			content: httpBytes,
			emoteID: emoteIdStr,
		}
	}(emoteID)
}

// GetEmote loads the image data and returns an gd.Image
func (h *GodotTwitchEmoteStore) GetEmote(emoteID string) Image.Instance {
	return h.emoteImageCache[emoteID]
}

func (h *GodotTwitchEmoteStore) addEventQueue(emoteID string) {
	h.emoteQueueLock.Lock()
	defer h.emoteQueueLock.Unlock()

	if len(h.emoteReadyEventQueue) <= h.emoteQueueIndex {
		h.emoteReadyEventQueue = append(h.emoteReadyEventQueue, emoteID)
	} else {
		h.emoteReadyEventQueue[h.emoteQueueIndex] = emoteID
	}

	h.emoteQueueIndex += 1
}
