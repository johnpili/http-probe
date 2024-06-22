package main

import (
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/rs/xid"
	"gopkg.in/yaml.v2"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	socketio "github.com/googollee/go-socket.io"
)

var (
	cookieStore    *sessions.CookieStore
	configuration  *Config
	socketIOServer *socketio.Server

	//go:embed static/*
	staticFS embed.FS

	//go:embed views/*
	viewsFS embed.FS
)

// This will handle the loading of config.yml
func loadConfiguration(c string) {
	f, err := os.Open(c)
	if err != nil {
		log.Fatal(err.Error())
	}

	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(&configuration)
	if err != nil {
		log.Fatal(err.Error())
	}
}

func main() {
	pid := os.Getpid()
	err := os.WriteFile("application.pid", []byte(strconv.Itoa(pid)), 0666)
	if err != nil {
		log.Fatal(err)
	}

	var configLocation string
	flag.StringVar(&configLocation, "config", "config.yml", "Set the location of configuration file")
	flag.Parse()
	loadConfiguration(configLocation)

	cookieStore = sessions.NewCookieStore(securecookie.GenerateRandomKey(32))

	socketIOServer = setupSocketIO()
	go socketIOServer.Serve()
	defer socketIOServer.Close()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /", dashboard)
	mux.HandleFunc("GET /{id}", dashboard)

	mux.HandleFunc("GET /send/{id}", receiver)
	mux.HandleFunc("POST /send/{id}", receiver)
	mux.HandleFunc("PUT /send/{id}", receiver)
	mux.HandleFunc("DELETE /send/{id}", receiver)

	mux.Handle("GET /static/", http.FileServerFS(staticFS))

	mux.Handle("GET /socket.io/", socketIOServer)
	mux.Handle("POST /socket.io/", socketIOServer)

	httpServer := &http.Server{
		Addr:              fmt.Sprintf(":%d", configuration.HTTP.Port),
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       120 * time.Second,
		WriteTimeout:      120 * time.Second,
	}

	log.Printf("Server running at http://localhost:%d/\n", configuration.HTTP.Port)
	log.Fatal(httpServer.ListenAndServe())
}

func setupSocketIO() *socketio.Server {
	server := socketio.NewServer(nil)

	server.OnConnect("/", func(s socketio.Conn) error {
		s.SetContext("")
		url := s.URL()
		rooms := url.Query()["r"]
		if len(rooms) > 0 {
			server.JoinRoom("/", rooms[0], s)
			log.Println("connected:", s.ID())
		}
		return nil
	})
	server.OnError("/", func(s socketio.Conn, e error) {
		log.Println("error", e)
	})
	server.OnDisconnect("/", func(s socketio.Conn, reason string) {
		log.Println("closed", reason)
	})
	return server
}

func dashboard(w http.ResponseWriter, r *http.Request) {
	session, err := cookieStore.Get(r, configuration.System.CookieName)
	if err != nil {
		generatedID := generateID(session, w, r)
		http.Redirect(w, r, fmt.Sprintf("/%s", generatedID), 302)
		return
	}

	id := r.PathValue("id")
	if len(id) == 0 {
		if id, ok := session.Values["id"].(string); ok {
			http.Redirect(w, r, fmt.Sprintf("/%s", id), 302)
			return
		}
		generatedID := generateID(session, w, r)
		http.Redirect(w, r, fmt.Sprintf("/%s", generatedID), 302)
		return
	}

	extractedID := ""
	if e, ok := session.Values["id"].(string); ok {
		extractedID = e
	}

	if len(extractedID) == 0 {
		http.Redirect(w, r, fmt.Sprintf("/"), 302)
		return
	}

	if strings.Compare(id, extractedID) != 0 { // Need to check if someone is using an ID that don't belong to their session
		http.Redirect(w, r, fmt.Sprintf("/%s", extractedID), 302)
		return
	}

	data := make(map[string]interface{})
	data["room"] = id

	page := NewPage()
	page.Title = "An HTTP probe tool with web interface | probe.johnpili.com"
	page.SetData(data)
	page.RenderPage(w, r, "views/base.html", "views/index.html")
}

func receiver(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if len(id) == 0 {
		return
	}

	rooms := socketIOServer.Rooms("/")
	for _, room := range rooms {
		if room == id {
			var sb strings.Builder
			fmt.Fprintf(&sb, "Method: %s\n\n", r.Method)

			var sbHeaders strings.Builder
			fmt.Fprintf(&sb, "Headers:\n\n")
			for name, headers := range r.Header {
				for _, h := range headers {
					fmt.Fprintf(&sb, "%v: %v\n", name, h)
					fmt.Fprintf(&sbHeaders, "%v: %v\n", name, h)
				}
			}
			fmt.Fprintf(&sb, "\n")

			username, password, ok := r.BasicAuth()
			if ok {
				fmt.Fprintf(&sb, "Basic Auth:\n\n")
				fmt.Fprintf(&sb, "Username: %s\nPassword: %s\n\n", username, password)
			}

			var sbBody strings.Builder
			bodyBuffer, _ := io.ReadAll(http.MaxBytesReader(w, r.Body, 1024))
			if len(bodyBuffer) > 0 {
				fmt.Fprintf(&sb, "Body:\n\n")
				isJSON := false
				contentType := r.Header.Get("Content-Type")
				if len(contentType) == 0 {
					if strings.ToLower(contentType) == "application/json" {
						isJSON = true
					}
				}

				if isJSON {
					prettyJSON, _ := json.MarshalIndent(bodyBuffer, "", "    ")
					fmt.Fprintf(&sb, "%v\n", string(prettyJSON))
					fmt.Fprintf(&sbBody, "%v\n", string(prettyJSON))
				} else {
					fmt.Fprintf(&sb, "%v\n", string(bodyBuffer))
					fmt.Fprintf(&sbBody, "%v\n", string(bodyBuffer))
				}
			}

			ackReference := xid.New().String()
			ackTimestamp := time.Now()

			socketIOServer.BroadcastToRoom("/", id, fmt.Sprintf("dump@%s", id), r.Method, ackReference, ackTimestamp, sbHeaders.String(), sbBody.String(), sb.String())

			respondStatusCodeWithJSON(w, http.StatusOK, ResponseTicket{
				AckReference: ackReference,
				AckTimestamp: ackTimestamp,
			})
			return
		}
	}

	w.WriteHeader(http.StatusNotFound)
}

func generateID(session *sessions.Session, w http.ResponseWriter, r *http.Request) string {
	generatedID := generateGUID()
	session.Values["id"] = generatedID
	session.Options.MaxAge = 0
	_ = session.Save(r, w)
	return generatedID
}
