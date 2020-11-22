package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	gpio "github.com/warthog618/gpio"

	"github.com/gorilla/handlers"
)

const STATUS_FIRE_OFF = 0
const STATUS_FIRE_ON = 1

var fireState int = STATUS_FIRE_OFF
var gpioLock = &sync.Mutex{}
var pin *gpio.Pin

type FireResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	State   string `json:"state"`
}

func getFireState() int {
	gpioLock.Lock()
	defer gpioLock.Unlock()
	return fireState
}

func fireToggle() int {
	if getFireState() == STATUS_FIRE_OFF {
		setFireOn()
		return fireState
	}
	setFireOff()
	return fireState
}

func setFireOn() {
	gpioLock.Lock()
	defer gpioLock.Unlock()
	pin.High()
	log.Printf("toggled fire to ON")
	fireState = STATUS_FIRE_ON
}

func setFireOff() {
	gpioLock.Lock()
	defer gpioLock.Unlock()
	pin.Low()
	log.Printf("toggled fire to OFF")
	fireState = STATUS_FIRE_OFF
}

func formatfireState() string {
	if fireState == STATUS_FIRE_OFF {
		return "off"
	}
	return "on"
}

func main() {

	gpio_err := gpio.Open()
	if gpio_err != nil {
		panic(gpio_err)
	}

	pin = gpio.NewPin(4)
	pin.Output()

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(getHomePage()))
	})

	mux.HandleFunc("/fire/on", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		setFireOn()
		res, _ := json.Marshal(&FireResponse{Success: true, Message: "Fire On", State: formatfireState()})
		w.Write(res)
	})

	mux.HandleFunc("/fire/off", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		setFireOff()
		res, _ := json.Marshal(&FireResponse{Success: true, Message: "Fire Off", State: formatfireState()})
		w.Write(res)
	})

	mux.HandleFunc("/fire/status", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		res, _ := json.Marshal(&FireResponse{Success: true, Message: "", State: formatfireState()})
		w.Write(res)
	})

	mux.HandleFunc("/fire/toggle", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fireToggle()
		res, _ := json.Marshal(&FireResponse{Success: true, Message: "Fire Toggled", State: formatfireState()})
		w.Write(res)
	})

	server := &http.Server{
		Addr:         ":80",
		Handler:      handlers.LoggingHandler(os.Stdout, mux),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	server.ListenAndServe()
}

func getHomePage() string {
	return fmt.Sprintf(`<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Strict//EN" "http://www.w3.org/TR/xhtml1/DTD/xhtml1-strict.dtd">
<html xmlns="http://www.w3.org/1999/xhtml">
<head>
<title>Fire Control</title>
<meta name="HandheldFriendly" content="True"/>
<meta name="MobileOptimized" content="320"/>
<meta name="viewport" content="width=device-width, initial-scale=1"/>

<script language="javascript">

	var page_is_hidden = false;

	function sendReq(uri) {

		// Avoid sending requests when the page is hidden
		if (page_is_hidden) return;

		xhr = new XMLHttpRequest();
		xhr.onreadystatechange = function() {
			if (this.readyState == 4 && this.status == 200) {
				var info = JSON.parse(this.responseText);
				showResponse(info);
			}
		};
		xhr.open("GET", uri, true);
		xhr.send();		
	}

	function showResponse(info) {
		document.getElementById("success").innerHTML = "Success: " + info.success;
		if (info.message != "") {
			document.getElementById("message").innerHTML = "Message: " + info.message;
		} else {
			document.getElementById("message").innerHTML = "";
		}
		// document.getElementById("state").innerHTML = "State: " + info.state;
		document.getElementById("fire_state").innerHTML = "Fire is " + info.state;
	}

	function fireToggle() {
		sendReq("/fire/toggle");
		return false;
	}
	
	function fireOn() {
		sendReq("/fire/on");
		return false;
	}
	
	function fireOff() {
		sendReq("/fire/off");
		return false;
	}
	
	function updateFireStatus() {
		// Update the status fields
		sendReq("/fire/status");
		return false;
	}

	// Set the name of the hidden property and the change event for visibility
	var hidden, visibilityChange; 
	if (typeof document.hidden !== "undefined") { // Opera 12.10 and Firefox 18 and later support 
	  hidden = "hidden";
	  visibilityChange = "visibilitychange";
	} else if (typeof document.msHidden !== "undefined") {
	  hidden = "msHidden";
	  visibilityChange = "msvisibilitychange";
	} else if (typeof document.webkitHidden !== "undefined") {
	  hidden = "webkitHidden";
	  visibilityChange = "webkitvisibilitychange";
	}

	function handleVisibilityChange() {
	  if (document[hidden]) {
	    page_is_hidden = true;
	  } else {
	    page_is_hidden = false;
	  }
	}

	// Warn if the browser doesn't support addEventListener or the Page Visibility API
	if (typeof document.addEventListener === "undefined" || hidden === undefined) {
	  console.log("This demo requires a browser, such as Google Chrome or Firefox, that supports the Page Visibility API.");
	} else {
	  // Handle page visibility change   
	  document.addEventListener(visibilityChange, handleVisibilityChange, false);
	}

</script>

<style type="text/css">
body {
	background: white;
	color: #222;
	font-family: 'Trebuchet MS', 'Tahoma', 'Arial', 'Helvetica';
  }

.button {
    background-color: #e7e7e7;
    border: none;
    color: white;
    padding: 10px 20px;
    text-align: center;
    text-decoration: none;
    display: inline-block;
    font-size: 16px;
}

#btnOn {
	background-color: #4CAF50;
}

#btnKeep {
	background-color: #f44336;
}

#btnOff {
	background-color: #444444;
}

#response {
	padding: 10px;
	font-size: 10px;
	text-align; center;
	color: #ccc;
	visibility: hidden;
}

#fire_controls {
	margin-bottom: 10px;
	text-align: center;
	padding: 5px;
}

#fire_controls2 {
	margin-top: 10px;
	text-align: center;
	padding: 5px;
}

#fire {
	text-align: center;
}

</style>

<body>
	<div id="fire">

		<div id="fire_controls">
			  <h2 id="fire_state">Fire is %s</h2>
			  <button class="button" id="btnOn" onclick="fireToggle()">Toggle FIRE</button>
		</div>

		<div id="response">
			  <span id="success"></span>
			  <span id="message"></span>
			  <span id="state"></span>
		</div>

	</div>

<script language="javascript">
	setInterval(function(){ updateFireStatus() }, 1000);
</script>

<div id="fire_controls2">
<button class="button" id="btnKeep" onclick="fireOn()">Fire ON</button>
<button class="button" id="btnOff" onclick="fireOff()">Fire OFF</button>
</div>

</body>
</html>
    `, formatfireState())
}
