package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/kkdai/youtube/v2"
	"github.com/rs/cors"
	ffmpeg_go "github.com/u2takey/ffmpeg-go"
)

type Caption struct {
	Start float64 `json:"start"`
	End   float64 `json:"end"`
	Text  string  `json:"text"`
}

type CreateSessionRequest struct {
	Type     string `json:"type"`
	VideoUrl string `json:"videoUrl"`
}

type Session struct {
	Id        string     `json:"id"`
	Type      string     `json:"type"`
	VideoUrl  string     `json:"videoId"`
	ClipStart float64    `json:"clipStart"`
	ClipEnd   float64    `json:"clipEnd"`
	Captions  []*Caption `json:"captions"`
}

type DirItem struct {
	Name string `json:"name"`
}

func save_full_video(video_id string) {

	// yt-dlp https://www.youtube.com/watch\?v\=dQw4w9WgXcQ -o full
	// 	we'll need to know the output format
	// ffmpeg -i full.webm -filter_complex "fps=30,scale=240:-2" -b:a 96k full-preview.mp4

	fmt.Println("Downloading video...")
	client := youtube.Client{}

	video, err := client.GetVideo(video_id)
	if err != nil {
		panic(err)
	}

	formats := video.Formats.WithAudioChannels()
	stream, _, err := client.GetStream(video, &formats[0])
	if err != nil {
		panic(err)
	}

	// TODO replace
	file, err := os.Create("data/full.mp4")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	_, err = io.Copy(file, stream)
	if err != nil {
		panic(err)
	}
}

func save_preview() {
	// -filter_complex "fps=30,scale=240:-2"
	fmt.Println("Saving preview...")
	ffmpeg_go.
		Input("data/full.mp4").
		Filter("fps", ffmpeg_go.Args{"30"}).
		Filter("scale", ffmpeg_go.Args{"240:-2"}).
		Output("data/preview.mp4", ffmpeg_go.KwArgs{"map": "0:a"}).
		OverWriteOutput().ErrorToStdOut().Run()
}

// func get_preview() {

// }

func handleRoot(w http.ResponseWriter, r *http.Request) {

	fmt.Println("Received a request")
	if r.Method == "POST" {
		decoder := json.NewDecoder(r.Body)
		var body CreateSessionRequest
		err := decoder.Decode(&body)
		if err != nil {
			fmt.Println("Failed to decode CreateSessionRequest")
			w.WriteHeader(http.StatusInternalServerError)
		}

		session := &Session{
			Id:        uuid.New().String(),
			Type:      body.Type,
			VideoUrl:  body.VideoUrl,
			ClipStart: 0,
			// TODO figure out clip end
			ClipEnd:  100,
			Captions: make([]*Caption, 0),
		}
		// TODO load video, save title, save session, or error
		// TODO save session here
		fmt.Println("Created session with id " + session.Id)
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(session)
	}
}

func handleList(w http.ResponseWriter, r *http.Request) {

	fmt.Println("received a list request")
	fmt.Println("Reading files from " + os.Getenv("HOME"))
	files, _ := os.ReadDir(os.Getenv("HOME"))

	dirItems := make([]*DirItem, 0)
	for i := 0; i < len(files); i++ {
		name := files[i].Name()

		if []rune(name)[0] == rune('.') {
			// do nothing
		} else {
			dirItems = append(dirItems, &DirItem{Name: name})
		}
	}

	// bytes, _ := json.Marshal(dirItems)
	// jsons := string(bytes)

	// write back headers, status and json
	w.Header().Set("Content-Type", "application/json")
	// w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)

}

func handlePreview(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Cache-Control", "no-store")

	videoID := "dQw4w9WgXcQ"

	if r.Method == "POST" {
		// TODO create session (GUID?)
		session := &Session{
			Id:        uuid.New().String(),
			Type:      "blah",
			VideoUrl:  "",
			ClipStart: 0,
			ClipEnd:   100,
			Captions:  make([]*Caption, 0),
		}

		fmt.Println("New session: " + session.Id)
		save_full_video(videoID)
		save_preview()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(session)
	} else if r.Method == "PUT" {
		save_preview()
	} else {
		// will 404 if file not found
		http.ServeFile(w, r, "data/preview.mp4")
	}
}

func handleSessionPreview(w http.ResponseWriter, r *http.Request) {
	args := strings.Split(r.URL.Path, "/")
	fmt.Println("not really no" + args[0])
}

func handleNothing(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}

func main() {
	fmt.Println("Starting pkgif backend")

	// set up routing
	mux := http.NewServeMux()
	mux.HandleFunc("/", handleRoot)
	mux.HandleFunc("/list", handleList)
	mux.HandleFunc("/preview", handlePreview)
	mux.HandleFunc("/nothing", handleNothing)
	mux.HandleFunc("/session/", handleSessionPreview)

	// register cors handler
	handler := cors.Default().Handler(mux)

	// TODO use flags to set port
	http.ListenAndServe(":7131", handler)
}
