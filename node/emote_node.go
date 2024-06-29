package node

import (
	"fmt"
	"io"
	"main/lib"
	"net/http"
	"sync"

	"grow.graphics/gd"
)

type GodotTwitchEmoteStore struct {
	gd.Class[GodotTwitchEmoteStore, gd.Node]

	UseDarkTheme gd.Bool `gd:"use_dark_theme"`
	Scale        gd.Int  `gd:"scale"`

	OnEmoteReady gd.SignalAs[func(gd.String)] `gd:"on_emote_ready"`

	emoteQueueLock       *sync.Mutex
	emoteReadyEventQueue []string
	emoteQueueIndex      int
	emoteImageCache      map[string]gd.Object
	emoteByteChan        chan emoteByteResponse

	httpClient *http.Client
}

type emoteByteResponse struct {
	content []byte
	emoteID string
}

func (h *GodotTwitchEmoteStore) Ready(godoCtx gd.Context) {
	h.httpClient = &http.Client{}
	h.emoteReadyEventQueue = make([]string, 0, 10)
	h.emoteQueueLock = &sync.Mutex{}
	h.emoteByteChan = make(chan emoteByteResponse, 1)
	h.emoteImageCache = make(map[string]gd.Object)
}

func (h *GodotTwitchEmoteStore) Process(godoCtx gd.Context, delta gd.Float) {
	if h.emoteQueueIndex > 0 {
		h.emoteQueueLock.Lock()
		defer h.emoteQueueLock.Unlock()

		for i := h.emoteQueueIndex; i > 0; i-- {
			h.OnEmoteReady.Emit(godoCtx.String(h.emoteReadyEventQueue[i-1]))
		}

		h.emoteQueueIndex = 0
	}

	if len(h.emoteByteChan) > 0 {
		emoteResponse := <-h.emoteByteChan

		byteArray := h.Pin().PackedByteArray()
		byteArray.Resize(int64(len(emoteResponse.content)))
		for index, httpByte := range emoteResponse.content {
			byteArray.SetIndex(gd.Int(index), httpByte)
		}

		if !gd.DirAccess.DirExistsAbsolute(gd.DirAccess{}, godoCtx, godoCtx.String("user://emotes")) {
			gd.DirAccess.MakeDirAbsolute(gd.DirAccess{}, godoCtx, godoCtx.String("user://emotes"))
		}

		path := godoCtx.String(fmt.Sprintf("user://emotes/%s", emoteResponse.emoteID))
		accessWriteFa := gd.FileAccess.Open(gd.FileAccess{}, godoCtx, path, gd.FileAccessWrite)
		accessWriteFa.StoreBuffer(byteArray)

		emoteImage := gd.Create(h.Pin(), &gd.Image{})
		emoteImage.LoadPngFromBuffer(byteArray)

		h.emoteImageCache[emoteResponse.emoteID] = emoteImage.AsObject()
		h.addEventQueue(emoteResponse.emoteID)
	}
}

// LoadEmote makes sure the emote is loaded and available and triggers a "on_emote_ready" after.
func (h *GodotTwitchEmoteStore) LoadEmote(godoCtx gd.Context, emoteID gd.String) {
	if _, isCached := h.emoteImageCache[emoteID.String()]; isCached {
		h.OnEmoteReady.Emit(emoteID)
		return
	}

	path := godoCtx.String(fmt.Sprintf("user://emotes/%s", emoteID.String()))

	if gd.FileAccess.FileExists(gd.FileAccess{}, godoCtx, path) {
		emoteFa := gd.FileAccess.Open(gd.FileAccess{}, godoCtx, path, gd.FileAccessRead)
		emoteBuffer := emoteFa.GetBuffer(godoCtx, emoteFa.GetLength())

		emoteImage := gd.Create(h.Pin(), &gd.Image{})
		emoteImage.LoadPngFromBuffer(emoteBuffer)

		h.emoteImageCache[emoteID.String()] = emoteImage.AsObject()
		h.addEventQueue(emoteID.String())
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
			lib.LogErr(godoCtx, fmt.Sprintf("error loading emote: %s", err.Error()))
			return
		}

		httpBytes, err := io.ReadAll(emoteResp.Body)
		if err != nil {
			lib.LogErr(godoCtx, fmt.Sprintf("error loading emote: %s", err.Error()))
			return
		}

		h.emoteByteChan <- emoteByteResponse{
			content: httpBytes,
			emoteID: emoteIdStr,
		}
	}(emoteID.String())
}

// GetEmote loads the image data and returns an gd.Image
func (h *GodotTwitchEmoteStore) GetEmote(godoCtx gd.Context, emoteID gd.String) gd.Object {
	return h.emoteImageCache[emoteID.String()]
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
