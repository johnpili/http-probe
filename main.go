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
	"github.com/psi-incontrol/go-sprocket/sprocket"
	"github.com/rs/xid"
	"gopkg.in/yaml.v2"

	socketio "github.com/googollee/go-socket.io"
)

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
	log.Println("| HTTP Probe                                         |")
	log.Println("| Author: John Pili                                  |")
	log.Println("------------------------------------------------------")
	loadConfiguration(configLocation)

	envCookieKey := os.Getenv("ENV_HTTP_PROBE_COOKIE_KEY")
	if len(envCookieKey) > 0 {
		configuration.System.CookieKey = envCookieKey
	}

	if len(configuration.System.CookieKey) <= 0 {
		log.Fatalln("Missing cookie_key, please set the key value in the config.yml")
	}

	cookieKey := configuration.System.CookieKey
	cookieStore = sessions.NewCookieStore([]byte(cookieKey))

	socketIOServer = setupSocketIO()
	go socketIOServer.Serve()
	defer socketIOServer.Close()

	viewBox := rice.MustFindBox("views")
	staticBox := rice.MustFindBox("static")

	controllersHub := controllers.New(viewBox, nil, cookieStore, &configuration)

	//#region SINGLE BINARY
	staticFileServer := http.StripPrefix("/static/", http.FileServer(staticBox.HTTPBox()))
	//#endregion

	router := bone.New()
	router.Get("/static/", staticFileServer)
	router.Get("/socket.io/", socketIOServer)
	router.Post("/socket.io/", socketIOServer)
	router.HandleFunc("/send/:id", httpDump)
	controllersHub.BindRequestMapping(router)

	// CODE FROM https://medium.com/@mossila/running-go-behind-iis-ce1a610116df
	port := strconv.Itoa(configuration.HTTP.Port)
	if os.Getenv("ASPNETCORE_PORT") != "" { // get enviroment variable that set by ACNM
		port = os.Getenv("ASPNETCORE_PORT")
	}

	httpServer := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  120 * time.Second,
		WriteTimeout: 120 * time.Second,
	}

	if configuration.HTTP.IsTLS {
		log.Printf("Server running at https://localhost:%s/\n", port)
		log.Fatal(httpServer.ListenAndServeTLS(configuration.HTTP.ServerCert, configuration.HTTP.ServerKey))
		return
	}
	log.Printf("Server running at http://localhost:%s/\n", port)
	log.Fatal(httpServer.ListenAndServe())
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

	sprocket.RespondOkayJSON(w, models.ResponseTicket{
		AckReference: ackReference,
		AckTimestamp: ackTimestamp,
	})
	return
}
