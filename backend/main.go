package main

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/kkdai/youtube/v2"
	"github.com/rs/cors"
	ffmpeg_go "github.com/u2takey/ffmpeg-go"
)

const DEFAULT_PORT int = 7131

const NOT_CACHED int8 = 0
const DOWNLOADING int8 = 1
const CACHED int8 = 2

const FONT_PATH string = "/System/Library/Fonts/Supplemental/Impact.ttf"

var client youtube.Client
var clips map[string]*Clip
var cache map[string]int8

var testClip *Clip

type Caption struct {
	Start float64 `json:"start"`
	End   float64 `json:"end"`
	Text  string  `json:"text"`
}

type CreateVideoRequest struct {
	Type     string `json:"type"`
	VideoUrl string `json:"videoUrl"`
	Save     bool   `json:"save"`
}

type Video struct {
	Id         string              `json:"id"`
	Title      string              `json:"title"`
	Length     int32               `json:"length"`
	Thumbnails []youtube.Thumbnail `json:"thumbnails"`
	Status     int8                `json:"status"`
}

// possibly merge with Clip
type CreateClipRequest struct {
	// Type      string `json:"type"`
	VideoId   string  `json:"videoId"`
	ClipStart float64 `json:"clipStart"`
	ClipEnd   float64 `json:"clipEnd"`
}

type Clip struct {
	Id          string     `json:"id"`
	Title       string     `json:"title"`
	Type        string     `json:"type"`
	VideoId     string     `json:"videoId"`
	VideoLength int32      `json:"videoLength"`
	ClipStart   float64    `json:"clipStart"`
	ClipEnd     float64    `json:"clipEnd"`
	Captions    []*Caption `json:"captions"`
	videoId     string
	dirty       bool
}

// video processing
func getVideo(videoUrl string, save bool) *Video {

	// client.GetVideo works with URL or video ID
	video, err := client.GetVideo(videoUrl)

	if err != nil {
		return nil
	}

	videoLength := getVideoLengthSeconds(video)
	// getCacheStatus()
	status, ok := cache[video.ID]
	if !ok {
		status = NOT_CACHED
	}

	if save && status == NOT_CACHED {
		cache[video.ID] = DOWNLOADING
		go doCacheVideo(video)
	}

	return &Video{
		Id:         video.ID,
		Title:      video.Title,
		Length:     videoLength,
		Thumbnails: video.Thumbnails,
		// we will use status to track if is cached, downloading, or otherwise
		Status: status,
	}
}

func doCacheVideo(video *youtube.Video) {

	videoId := video.ID
	videoPath := fmt.Sprintf("data/%s.mp4", videoId)

	if _, err := os.Stat(videoPath); err == nil {
		cache[videoId] = CACHED
		fmt.Printf("Video %s already exists in cache. Skipping download\n", videoId)
	} else {
		// TODO UI needs some progress on this
		if cache[videoId] == DOWNLOADING {
			return
		}
		fmt.Printf("Video %s does not exist in cache. Downloading to '%s'\n", videoId, videoPath)
		formats := video.Formats.WithAudioChannels()
		stream, _, err := client.GetStream(video, &formats[0])
		if err != nil {
			panic(err)
		}

		// TODO replace
		file, err := os.Create(videoPath)
		if err != nil {
			panic(err)
		}

		defer file.Close()

		_, err = io.Copy(file, stream)
		if err != nil {
			panic(err)
		}

		cache[videoId] = CACHED
		fmt.Println("Download complete")
	}
}

func getVideoLengthSeconds(video *youtube.Video) int32 {
	videoLength := math.Floor(video.Duration.Seconds())
	return int32(videoLength)
}

// controller methods
func handleVideo(w http.ResponseWriter, r *http.Request) {

	if r.Method == "POST" {
		decoder := json.NewDecoder(r.Body)
		var body CreateVideoRequest
		_ = decoder.Decode(&body)

		video := getVideo(body.VideoUrl, body.Save)
		if video == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(video)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func handleVideoId(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodGet {
		vars := mux.Vars(r)
		videoId := vars["id"]

		video := getVideo(videoId, false)
		if video == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(video)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func handleVideoIdPreview(w http.ResponseWriter, r *http.Request) {
	// get
	vars := mux.Vars(r)
	videoId := vars["id"]

	previewPath := fmt.Sprintf("data/%s-preview.mp4", videoId)

	if _, err := os.Stat(previewPath); err == nil {
		fmt.Printf("Preview exists in cache. Skipping render\n")
	} else {
		fmt.Println("Preview is dirty or not in cache. Rendering")
		filePath := fmt.Sprintf("data/%s.mp4", videoId)

		ffmpeg_go.
			Input(filePath).
			Filter("fps", ffmpeg_go.Args{"15"}).
			Filter("scale", ffmpeg_go.Args{"240:-2"}).
			Output(previewPath, ffmpeg_go.KwArgs{"map": "0:a"}).
			OverWriteOutput().ErrorToStdOut().Run()

	}

	w.Header().Set("Cache-Control", "no-store")
	http.ServeFile(w, r, previewPath)
}

// POST /clip
// GET 	/clip
func handleClip(w http.ResponseWriter, r *http.Request) {

	fmt.Printf("Received a %s request\n", r.Method)

	if r.Method == http.MethodPost {

		decoder := json.NewDecoder(r.Body)
		var body CreateClipRequest
		err := decoder.Decode(&body)
		if err != nil {
			fmt.Println("Failed to decode CreateClipRequest")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		clip := createClip(&body)
		clips[clip.Id] = clip

		fmt.Println("Created clip with id " + clip.Id)
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(clip)

	} else if r.Method == http.MethodGet {

		clips := getAllClips()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(clips)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func handleClipId(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	clipId := vars["id"]

	var clip *Clip = clips[clipId]

	if r.Method == http.MethodPut {

		if clip == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		decoder := json.NewDecoder(r.Body)
		var body Clip
		_ = decoder.Decode(&body)

		fmt.Println("Changes received for clip id: " + clipId)
		clip.Captions = body.Captions
		clip.dirty = true

	} else if r.Method == http.MethodGet {
		if clip == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
	} else {
		// can probably do this in router
		w.WriteHeader(http.StatusMethodNotAllowed)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(clip)
}

func renderClip(clip *Clip, preview bool) string {

	// TODO need to separate video ID from clip ID here

	var outputPath string = fmt.Sprintf("data/%s.gif", clip.Id)
	if preview {
		outputPath = fmt.Sprintf("data/%s-preview.mp4", clip.Id)
	}

	// TODO dirty preview, dirty non-preview
	if _, err := os.Stat(outputPath); err == nil && !clip.dirty && preview {
		fmt.Printf("Preview exists in cache or is not dirty. Skipping render\n")
	} else {

		fmt.Println("Rendering")

		// TODO move to clip preview method
		videoId := clip.VideoId
		var inputPath string = fmt.Sprintf("data/%s.mp4", videoId)
		if preview {
			inputPath = fmt.Sprintf("data/%s-preview.mp4", videoId)
		}

		clipStart := clip.ClipStart
		clipEnd := clip.ClipEnd

		input_args := ffmpeg_go.KwArgs{"ss": clipStart}

		if clipEnd > clipStart {
			input_args = ffmpeg_go.KwArgs{"ss": clipStart, "t": clipEnd - clipStart}
		}

		stream := ffmpeg_go.
			Input(inputPath, input_args).
			Filter("fps", ffmpeg_go.Args{"15"})

		if preview {
			stream = stream.Filter("scale", ffmpeg_go.Args{"240:-2"})
		}

		fontsize := "48"
		if preview {
			fontsize = "16"
		}

		for _, c := range clip.Captions {
			fmt.Printf("Adding caption '%s' from %f to %f", c.Text, c.Start, c.End)
			stream = stream.Drawtext(c.Text, 0, 0, false, ffmpeg_go.KwArgs{
				"x":           "(w-text_w)/2",
				"y":           "h-th-10",
				"fontcolor":   "white",
				"fontfile":    FONT_PATH,
				"fontsize":    fontsize,
				"borderw":     "2",
				"bordercolor": "black",
				"enable":      fmt.Sprintf("between(t,%f,%f)", c.Start, c.End),
			})
		}
		// drawtext=fontfile=/path/to/font.ttf:text='Stack Overflow':fontcolor=white:fontsize=24:box=1:boxcolor=black@0.5:boxborderw=5:x=(w-text_w)/2:y=(h-text_h)/2:enable='between(t,5,10)'
		if preview {
			stream.
				Output(outputPath, ffmpeg_go.KwArgs{"map": "0:a"}).
				OverWriteOutput().ErrorToStdOut().Run()
		} else {
			stream.
				Output(outputPath).
				OverWriteOutput().ErrorToStdOut().Run()
		}

		clip.dirty = false
	}

	return outputPath
}

func handleClipIdPreview(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodGet {

		vars := mux.Vars(r)
		clipId := vars["id"]

		fmt.Println("Preview requested for clip id: " + clipId)
		previewPath := renderClip(clips[clipId], true)

		w.Header().Set("Cache-Control", "no-store")
		http.ServeFile(w, r, previewPath)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func handleClipIdRender(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodGet {

		vars := mux.Vars(r)
		clipId := vars["id"]

		fmt.Println("Render requested for clip id: " + clipId)
		renderPath := renderClip(clips[clipId], false)

		w.Header().Set("Cache-Control", "no-store")
		http.ServeFile(w, r, renderPath)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func createClip(body *CreateClipRequest) *Clip {

	// get video metadata
	video, err := client.GetVideo(body.VideoId)
	if err != nil {
		// TODO return error to the front end so it doesn't request a preview (or requests teapot preview)
		panic(err)
	}

	title := video.Title
	videoId := video.ID

	// download or get the cached video
	doCacheVideo(video)

	videoLength := getVideoLengthSeconds(video)

	clip := &Clip{
		Id:      uuid.New().String(),
		Title:   title,
		VideoId: body.VideoId,
		// may want to use floats actually for super short videos
		VideoLength: videoLength,
		ClipStart:   body.ClipStart,
		ClipEnd:     body.ClipEnd,
		Captions:    make([]*Caption, 0),
		videoId:     videoId,
		// not explicit ;)
		dirty: true,
	}

	clips[clip.Id] = clip

	// TODO start preview, etc
	// TODO load video, save title, save clip, or error
	// TODO save clip here

	return clip
}

func getAllClips() []*Clip {

	clipsArray := make([]*Clip, 0)
	for _, v := range clips {
		clipsArray = append(clipsArray, v)
	}
	return clipsArray
}

func handleTestPreview(w http.ResponseWriter, r *http.Request) {

	previewPath := renderClip(testClip, true)

	w.Header().Set("Cache-Control", "no-store")
	http.ServeFile(w, r, previewPath)

	// test clip should always be dirty
	testClip.dirty = true
}

func handleTestRender(w http.ResponseWriter, r *http.Request) {

	previewPath := renderClip(testClip, false)

	w.Header().Set("Cache-Control", "no-store")
	http.ServeFile(w, r, previewPath)

	// test clip should always be dirty
	testClip.dirty = true
}

func main() {

	fmt.Println("Starting pkgif backend")
	client = youtube.Client{}
	clips = map[string]*Clip{}
	cache = make(map[string]int8)

	testClip = &Clip{
		Id:          "b6cab676-9952-46bf-969b-de7099627ae8",
		Title:       "Max and Paddys Road to Nowhere Episode 01",
		VideoId:     "cpPeXEh5Wkk",
		VideoLength: 1423,
		ClipStart:   709.007074,
		ClipEnd:     711.429409,
		Captions: []*Caption{
			&Caption{Start: 0.416659, End: 1.425548, Text: "we went from that"},
			&Caption{Start: 1.59158, End: 2.556552, Text: "to that"},
		},
		dirty: true,
	}

	// set up routing
	fmt.Println("Initialising router")
	mux := mux.NewRouter()

	// TODO
	// - POST 	/clip: 					create a clip with optional start/end parameters
	// 									returns uuid of clip state. may also wish to return status of import, etc
	// - GET 	/clip: 					get all clips
	// - GET 	/clip/{id}: 			get specific clip details
	// - PUT 	/clip/{id}: 			update start/end parameters of clip
	// 									app must determine whether or not re-render is required
	// 									may need to offset captions
	// - PUT 	/clip/{id}/captions: 	replace clip captions
	// - GET 	/clip/{id}/preview: 	get clip preview
	//
	// - GET 	/cache:					get details of all videos in the cache
	// - DELETE /cache: 				empty cache of all videos
	// - DELETE /cache/{videoId} 		delete specific video id from cache

	mux.HandleFunc("/video", handleVideo)
	mux.HandleFunc("/video/{id}", handleVideoId)
	mux.HandleFunc("/video/{id}/preview", handleVideoIdPreview)

	mux.HandleFunc("/clip", handleClip)
	mux.HandleFunc("/clip/{id}", handleClipId)
	mux.HandleFunc("/clip/{id}/preview", handleClipIdPreview)
	mux.HandleFunc("/clip/{id}/render", handleClipIdRender)

	mux.HandleFunc("/test/preview", handleTestPreview)
	mux.HandleFunc("/test/render", handleTestRender)

	// register cors handler

	// you can't just modify the options in Default, and trying to pass it
	// 	into New() overrides any defaults
	options := cors.Options{
		AllowedHeaders: []string{"Origin", "Accept", "Content-Type", "X-Requested-With"},
		AllowedMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut},
	}
	// options.AllowedMethods = append(options.AllowedMethods, http.MethodPut)
	// cors := cors.Default()
	cors := cors.New(options)

	handler := cors.Handler(mux)

	// TODO use flags to set port
	fmt.Println("Listening")
	http.ListenAndServe(fmt.Sprintf(":%d", DEFAULT_PORT), handler)
}
