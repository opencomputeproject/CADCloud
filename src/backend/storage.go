// MIT License
//
// Copyright (c) 2020 CADCloud
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package main

import (
    "base"
    "os"
    "net/http"
    "strings"
    "log"
    "fmt"
    "io/ioutil"
    "sync"
    "encoding/base64"
)

var storageRoot = os.Getenv("STORAGE_ROOT")
// write operation must be protected by a Mutex
var file sync.RWMutex

// This is getting a user file entry

func getEntry(username string) (string,int) {
	// The first letter of the username is used as a directory entry
	// if the directory exist we check for the usenarme.conf entry into it
	// if it is there we return the content of the file
	print(storageRoot + "/" + string(username[0]) + "\n")
	_, err := os.Stat(storageRoot + "/" + string(username[0]))
	if ( ! os.IsNotExist(err) ) {
		// The directory exist we must now check if the file exist
		_,err := os.Stat( storageRoot + "/" + string(username[0]) + "/" + username )
		if ( ! os.IsNotExist(err) ) {
			// We must return the file content into a string
			b,_ := ioutil.ReadFile(storageRoot + "/" + string(username[0]) + "/" + username)
			return string(b),1
		} else
		{
			return "",0
		}
	} else
	{
		return "",0
	}
}

// This is creating a user file entry

func createEntry(username string, content string) (int) {
	_, err := os.Stat(storageRoot + "/" + string(username[0]))
	file.Lock()
	defer file.Unlock()
	if (  os.IsNotExist(err) ) {
		// we must create the directory which will contain the file
		_ = os.Mkdir(storageRoot + "/" + string(username[0]), os.ModePerm)
	}
	_ = ioutil.WriteFile(storageRoot + "/" + string(username[0]) + "/" + username, []byte(content), os.ModePerm)
	return 1
}

func createImage(username string, content string) (int) {
        _, err := os.Stat(storageRoot + "/" + string(username[0]))
	
        file.Lock()
	defer file.Unlock()
        if (  os.IsNotExist(err) ) {
                // we must create the directory which will contain the file
                _ = os.Mkdir(storageRoot + "/" + string(username[0]), os.ModePerm)
        }
	// We have to remove the "base64, stuff"
	coI := strings.Index(content, ",")
        rawImage := string(content)[coI+1:]
        decodedBody, _ := base64.StdEncoding.DecodeString(rawImage)
	_ = ioutil.WriteFile(storageRoot + "/" + string(username[0]) + "/" + username + ".jpg", []byte(decodedBody), os.ModePerm)
	return 1
}

func getImage(username string) (string) {
	_, err := os.Stat(storageRoot + "/" + string(username[0]))
        file.Lock()
        defer file.Unlock()
        if (  os.IsNotExist(err) ) {
                // we must create the directory which will contain the file
                _ = os.Mkdir(storageRoot + "/" + string(username[0]), os.ModePerm)
		return ""
        }

        _, err = os.Stat(storageRoot + "/" + string(username[0]) + "/" + username + ".jpg")
	if ( os.IsNotExist(err) ) {
		var staticAssetsDir = os.Getenv("STATIC_ASSETS_DIR")
		content,_ := ioutil.ReadFile(staticAssetsDir + "images/forklift.png")
		encodedContent := base64.StdEncoding.EncodeToString(content)
		return encodedContent
	} else {
		content,_ := ioutil.ReadFile(storageRoot + "/" + string(username[0]) + "/" + username + ".jpg")
		encodedContent := base64.StdEncoding.EncodeToString(content)
		return encodedContent
	}
}

func deleteEntry(username string, content string) (int) {
	_, err := os.Stat(storageRoot + "/" + string(username[0]) + "/" + username)
	file.Lock()
	defer file.Unlock()
	if (  !os.IsNotExist(err) ) {
		_=os.Remove(storageRoot + "/" + string(username[0]) + "/" + username)
	}
	_, err = os.Stat(storageRoot + "/" + string(username[0]) + "/" + username + ".jpg")
	if (  !os.IsNotExist(err) ) {
                _=os.Remove(storageRoot + "/" + string(username[0]) + "/" + username + ".jpg")
        }
	return 1
}

func projectCallback(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
		case http.MethodGet:
			file.Lock()
			defer file.Unlock()
			_, err := os.Stat(storageRoot + "/projects.json")
			if (  os.IsNotExist(err) ) {
				return 
			}
			content,_ := ioutil.ReadFile( storageRoot+ "/projects.json")
			w.Write(content)
		case http.MethodPut:
			file.Lock()
			defer file.Unlock()
			_ = ioutil.WriteFile(storageRoot + "/projects.json", base.HTTPGetBody(r), os.ModePerm)
	}

}

func htmlCallback(w http.ResponseWriter, r *http.Request) {
        switch r.Method {
                case http.MethodGet:
                        file.Lock()
                        defer file.Unlock()
                        _, err := os.Stat(storageRoot + r.URL.Path)
                        if (  os.IsNotExist(err) ) {
                                return
                        }
                        content,_ := ioutil.ReadFile( storageRoot+ r.URL.Path)
                        w.Write(content)
        }

}

func userCallback(w http.ResponseWriter, r *http.Request) {
        var username string
	var filecontent string
	var return_value int
	// We must breakdown the words, because username is not always the last word
	path := strings.Split( r.URL.Path, "/" )
        if ( len(path) < 3 ) {
                http.Error(w, "401 Malformed URI", 401)
                return
        }
        username = path[2]
	var command string
	if ( len(path) > 3 ) {
		command = path[3]
	}
        switch r.Method {
                case http.MethodGet:
			// Serve the resource.
			// I must return the content of the user file if it does exist otherwise
			// an error
			switch command {
			case "avatar":
				w.Write([]byte(getImage(username)))
			default:
				filecontent, return_value=getEntry(username)
				if ( return_value != 0) {
					fmt.Fprint(w,filecontent)			
				} else {
					fmt.Fprintf(w,"Error")
				}
			}
                case http.MethodPut:
			print("Got a PUT Request \n")
			// Update an existing record.
			if (r.Header.Get("Content-Type") != "image/jpg" ) {
				createEntry(username,string(base.HTTPGetBody(r)))	
			} else {
				createImage(username,string(base.HTTPGetBody(r)))
			}
		case http.MethodDelete:
			print("Got a Delete Request \n")
			deleteEntry(username,string(base.HTTPGetBody(r)))
                default:
        }
}

func main() {
    print("=============================== \n")
    print("| Starting storage backend    |\n")
    print("| (c) 2019 CADCloud           |\n")
    print("| Development version -       |\n")
    print("| Private use only            |\n")
    print("=============================== \n")

    mux := http.NewServeMux()
    var StorageURI = os.Getenv("STORAGE_URI")
    var StorageTCPPORT = os.Getenv("STORAGE_TCPPORT")

    mux.HandleFunc("/user/", userCallback)
    mux.HandleFunc("/projects/" ,projectCallback)
    mux.HandleFunc("/html/" ,htmlCallback)

    log.Fatal(http.ListenAndServe(StorageURI+StorageTCPPORT, mux))
}

