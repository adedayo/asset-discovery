package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	discover "github.com/adedayo/asset-discovery/pkg"
)

var (
	mux = http.NewServeMux()
)

func init() {
	addRoutes()
}

func addRoutes() {
	mux.HandleFunc("/", rootHandler)
	mux.HandleFunc("/brands", brandHandler)

}

//ServeAPI serves the analysis service on the specified port
func ServeAPI(config Config) {
	hostPort := ":%d"
	log.Fatal(http.ListenAndServe(fmt.Sprintf(hostPort, config.ApiPort), mux))
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode("Digital asset discovery API is up")
}

func brandHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "only POST method is supported", http.StatusBadRequest)
		return
	}
	var dq DomainQuery
	data, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = json.Unmarshal(data, &dq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	ctx := r.Context()
	r.URL.Query().Get("")
	brands := discover.GetBrands(ctx, dq.Domain, discover.Config{})
	json.NewEncoder(w).Encode(brands)
}

type Config struct {
	ApiPort int
}

type DomainQuery struct {
	Domain string
}
