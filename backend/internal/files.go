package internal

import (
	"fmt"
	"net/http"
	"os"
)

type DirItem struct {
	Name string `json:"name"`
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
