package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	rice "github.com/GeertJohan/go.rice"
	"github.com/go-zoo/bone"
	"github.com/gorilla/sessions"
	"github.com/johnpili/http-probe/controllers"
	"github.com/johnpili/http-probe/models"
	"github.com/rs/xid"
	"gopkg.in/yaml.v2"

	socketio "github.com/googollee/go-socket.io"
)

type responseTicket struct {
	Reference string    `json:"reference"`
	DateTime  time.Time `json:"DateTime"`
}

// Configurations / Settings
var (
	configuration  models.Config
	cookieStore    *sessions.CookieStore
	socketIOServer *socketio.Server
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
	err := ioutil.WriteFile("application.pid", []byte(strconv.Itoa(pid)), 0666)
	if err != nil {
		log.Fatal(err)
	}

	var configLocation string
	flag.StringVar(&configLocation, "config", "config.yml", "Set the location of configuration file")
	flag.Parse()

	log.Println("------------------------------------------------------")
	log.Println("| HTTP-CAT                                           |")
	log.Println("| Author: John Pili                                  |")
	log.Println("------------------------------------------------------")

	cookieKey := os.Getenv(configuration.System.EnvCookieKey)

	loadConfiguration(configLocation)
	cookieStore = sessions.NewCookieStore([]byte(cookieKey))

	socketIOServer = setupSocketIO()
	go socketIOServer.Serve()
	defer socketIOServer.Close()

	viewBox := rice.MustFindBox("views")
	staticBox := rice.MustFindBox("static")
	router := bone.New()
	controllersHub := controllers.New(viewBox, nil, cookieStore, &configuration)
	//#region SINGLE BINARY
	staticFileServer := http.StripPrefix("/static/", http.FileServer(staticBox.HTTPBox()))
	//#endregion

	router.Handle("/socket.io/", socketIOServer)
	router.HandleFunc("/", controllersHub.PageController.DashboardHandler)
	router.HandleFunc("/:id", controllersHub.PageController.DashboardHandler)
	router.HandleFunc("/:id/in", httpDump)
	router.Handle("/static/", staticFileServer)

	// CODE FROM https://medium.com/@mossila/running-go-behind-iis-ce1a610116df
	port := strconv.Itoa(configuration.HTTP.Port)
	if os.Getenv("ASPNETCORE_PORT") != "" { // get enviroment variable that set by ACNM
		port = os.Getenv("ASPNETCORE_PORT")
	}

	if configuration.HTTP.IsTLS {
		log.Printf("Server running at https://localhost:%s/\n", port)
		log.Fatal(http.ListenAndServeTLS(":"+port, configuration.HTTP.ServerCert, configuration.HTTP.ServerKey, router)) // Start HTTP Server
		return
	}
	log.Printf("Server running at http://localhost:%s/\n", port)
	log.Fatal(http.ListenAndServe(":"+port, router)) // Start HTTP Server
}

func setupSocketIO() *socketio.Server {
	server, err := socketio.NewServer(nil)

	if err != nil {
		log.Fatal(err)
	}

	server.OnConnect("/", func(s socketio.Conn) error {
		s.SetContext("")

		url := s.URL()
		rooms := url.Query()["r"]
		if len(rooms) > 0 {
			//log.Println(rooms[0])
			server.JoinRoom("/", rooms[0], s)
			fmt.Println("connected:", s.ID())
		}
		return nil
	})

	server.OnEvent("/", "notice", func(s socketio.Conn, msg string) {
		//fmt.Println("notice:", msg)
		//log.Println(s.Namespace())
		//log.Println(s.Rooms())
		//server.BroadcastToRoom("/", "ginvera", "dump", msg)
	})

	/*server.OnEvent("/dump", "msg", func(s socketio.Conn, method  headers string, body string, raw string) string {
		s.SetContext("")
		//s.SetContext(raw)
		return "recv " + raw
	})*/
	server.OnEvent("/", "bye", func(s socketio.Conn) string {
		last := s.Context().(string)
		s.Emit("bye", last)
		s.Close()
		return last
	})
	server.OnError("/", func(s socketio.Conn, e error) {
		//fmt.Println("meet error:", e)
	})
	server.OnDisconnect("/", func(s socketio.Conn, reason string) {
		log.Println("closed", reason)
	})
	return server
}

func httpDump(w http.ResponseWriter, r *http.Request) {

	id := bone.GetValue(r, "id")
	if len(id) == 0 {
		return
	}

	if configuration.Simulator.EnableDelay {
		if configuration.Simulator.DelayType == "RANDOM" {
			rand.Seed(time.Now().UnixNano())
			delta := rand.Intn(configuration.Simulator.DelaySec + 1)
			time.Sleep(time.Duration(delta) * time.Second)
		} else if configuration.Simulator.DelayType == "FIXED" {
			rand.Seed(time.Now().UnixNano())
			delta := configuration.Simulator.DelaySec
			time.Sleep(time.Duration(delta) * time.Second)
		}
	}

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

	bodyBuffer, _ := ioutil.ReadAll(r.Body)
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

	socketIOServer.BroadcastToRoom("/", id, "dump", r.Method, ackReference, ackTimestamp, sbHeaders.String(), sbBody.String(), sb.String())

	respondWithJSON(w, responseTicket{
		Reference: ackReference,
		DateTime:  ackTimestamp,
	})
	return
}

func respondWithJSON(w http.ResponseWriter, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		log.Println(err.Error())
	}
	w.Header().Set("Content-Type", "application/json;charset=UTF-8")
	w.WriteHeader(200)
	w.Write(response)
}
