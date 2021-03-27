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
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
)

type cacheEntry struct {
	Nickname  string
	Key       string
	SecretKey string
	URI       string
	Port      string
	Buckets   []bucket
	entryMux  sync.Mutex
}

var cache []cacheEntry
var cacheMux sync.Mutex

type bucket struct {
	Name      string
	FileEntry int
	Revision  string
	FileList  []files
}
type files struct {
	File string
}

type minIOEntry struct {
	Username string
	Ports    string
}

var dnsdomain = os.Getenv("DNS_DOMAIN")

// This is getting a user file entry from the cache

func getEntry(Key string) string {
	for _, entry := range cache {
		if entry.Key == Key {
			value, _ := strconv.Atoi(entry.Port)
			value = value + base.MinIOServerBasePort
			result := strconv.Itoa(value)
			return entry.URI + ":" + result
		}
	}
	return ""
}

func getRevision(Key string, bucket string) string {
	for _, entry := range cache {
		if entry.Key == Key {
			for j := range entry.Buckets {
				if entry.Buckets[j].Name == bucket {
					return entry.Buckets[j].Revision
				}
			}
		}
	}
	return ""
}

func getSecretKey(Key string) string {
	for _, entry := range cache {
		if entry.Key == Key {
			return entry.SecretKey + " " + entry.Nickname
		}
	}
	return ""
}

func createBucketRevision(entry cacheEntry, Bucket bucket) {

	port, _ := strconv.Atoi(entry.Port)
	realport := port + 1000 + base.MinIOServerBasePort

	_, err := base.Request("PUT", "http://"+entry.URI+":"+strconv.Itoa(realport)+"/"+Bucket.Name+"r"+Bucket.Revision+"/", "/"+Bucket.Name+"r"+Bucket.Revision+"/", "application/xml", nil, "", entry.Key, entry.SecretKey)

	if err != nil {
		log.Fatal(err)
	}
}

func getnewRevision(entry cacheEntry, BucketName string) string {

	// We must get the end user bucket list
	// and check which one is the latest one
	// The bucket must be empty so it could had been created
	// and not used

	port, _ := strconv.Atoi(entry.Port)
	realport := port + 1000 + base.MinIOServerBasePort

	response, err := base.Request("GET", "http://"+entry.URI+":"+strconv.Itoa(realport)+"/", "/", "application/xml", nil, "", entry.Key, entry.SecretKey)

	if err != nil {
		log.Fatal(err)
	} else {
		// We can move forward
		// We have the list
		// Now is there a bucket with the "r" keyword at the end and a number and the same name than the requested one
		defer response.Body.Close()
		contents, err := ioutil.ReadAll(response.Body)
		if err != nil {
			log.Fatal(err)
		}
		// Got an XML content with the various Buckets created by the end user

		type Bucket struct {
			Name string `xml:"Name"`
		}

		type Content struct {
			XMLName xml.Name `xml:"ListAllMyBucketsResult"`
			Buckets []Bucket `xml:"Buckets>Bucket"`
		}

		XMLcontents := Content{}
		in := bytes.NewReader([]byte(contents))
		_ = xml.NewDecoder(in).Decode(&XMLcontents)

		// Ok Got the bucket list from the user

		prefix := BucketName + "r"
		version := 0
		found := 0
		for i := 0; i < len(XMLcontents.Buckets); i++ {
			if strings.HasPrefix(XMLcontents.Buckets[i].Name, prefix) {
				// must update the value
				versionNUmberString := strings.TrimPrefix(XMLcontents.Buckets[i].Name, prefix)
				newVersion, _ := strconv.Atoi(versionNUmberString)
				if newVersion == version {
					found = 1
				}
				if newVersion > version {
					version = newVersion
				}
			}
		}
		if version == 0 && found == 0 {
			return strconv.Itoa(version)
		}

		response, err := base.Request("GET", "http://"+entry.URI+":"+strconv.Itoa(realport)+"/"+BucketName+"r"+strconv.Itoa(version)+"/", "/"+BucketName+"r"+strconv.Itoa(version)+"/", "application/xml", nil, "", entry.Key, entry.SecretKey)

		if err != nil {
			log.Fatal(err)
		} else {
			contents, _ := ioutil.ReadAll(response.Body)
			// Got the current bucket list content into contents as an XML format. Let's loop through it
			// and issue the delete command
			// and re-issue a get command to check if the directory is empty. As long as it is not
			// We reload the directory entries and we remove the file
			type object struct {
				Name string `xml:"Key"`
			}
			type bucketContent struct {
				XMLName xml.Name `xml:"ListBucketResult"`
				Objects []object `xml:"Contents"`
			}
			XMLbucketcontents := bucketContent{}
			in := bytes.NewReader([]byte(contents))
			_ = xml.NewDecoder(in).Decode(&XMLbucketcontents)
			if len(XMLbucketcontents.Objects) != 0 {
				return strconv.Itoa(version + 1)
			} 
			return strconv.Itoa(version)
		}

	}
	return ""
}

func addFilesEntry(Key string, BucketName string, content string) string {
	for i, entry := range cache {
		if entry.Key == Key {
			// we must update the file entry structure
			var tmpFileList []files
			var emptyIndexes []int
			var j int
			var recovery int
			json.Unmarshal([]byte(content), &tmpFileList)
			// We need to suppress empty files
			for j = range tmpFileList {
				if tmpFileList[j].File == "" {
					entry.entryMux.Lock()
					emptyIndexes = append(emptyIndexes, j)
					entry.entryMux.Unlock()
				}
			}
			recovery = 0
			for j = range emptyIndexes {
				entry.entryMux.Lock()
				tmpFileList = append(tmpFileList[:(emptyIndexes[j]-recovery)], tmpFileList[(emptyIndexes[j]-recovery)+1:]...)
				entry.entryMux.Unlock()
				recovery = recovery + 1
			}
			// Is there soon a BucketName entry ? If not we must append it
			for j = range cache[i].Buckets {
				if cache[i].Buckets[j].Name == BucketName {
					break
				}
			}
			if j == len(cache[i].Buckets) {
				var tmpBucket bucket
				tmpBucket.Name = BucketName
				entry.entryMux.Lock()
				cache[i].Buckets = append(cache[i].Buckets, tmpBucket)
				entry.entryMux.Unlock()
			}
			entry.entryMux.Lock()
			cache[i].Buckets[j].FileEntry = cache[i].Buckets[j].FileEntry + 1
			cache[i].Buckets[j].FileList = append(cache[i].Buckets[j].FileList, tmpFileList...)
			entry.entryMux.Unlock()
		}
	}
	return ""
}

// This function is called when a GET is made on a bucket
// It is emptying the associated cache file entry which is filled during the PUT operation
// and used to detect the end of a FreeCAD / KiCAD file upload

func deleteFilesEntry(Key string, BucketName string) string {
	for i, entry := range cache {
		if entry.Key == Key {
			for j := range cache[i].Buckets {
				if cache[i].Buckets[j].Name == BucketName {
					entry.entryMux.Lock()
					cache[i].Buckets[j].FileList = nil
					cache[i].Buckets[j].FileEntry = 0
					entry.entryMux.Unlock()
					return ""
				}
			}
			// If it is not found we must load the cache with a new revision
			// It means this is a first access
			var newBucket bucket
			newBucket.Name = BucketName
			newBucket.FileList = nil
			newBucket.FileEntry = 0
			newBucket.Revision = getnewRevision(cache[i], BucketName)

			entry.entryMux.Lock()

			cache[i].Buckets = append(cache[i].Buckets, newBucket)
			createBucketRevision(cache[i], cache[i].Buckets[len(cache[i].Buckets)-1])

			entry.entryMux.Unlock()

			return ""
		}
	}
	return ""
}

func deleteFile(Key string, BucketName string, content string) string {
	for i, entry := range cache {
		if entry.Key == Key {
			for j := range cache[i].Buckets {
				if cache[i].Buckets[j].Name == BucketName {
					cache[i].entryMux.Lock()
					for k := range cache[i].Buckets[j].FileList {
						if cache[i].Buckets[j].FileList[k].File == content {
							cache[i].Buckets[j].FileList[k] = cache[i].Buckets[j].FileList[len(cache[i].Buckets[j].FileList)-1]
							cache[i].Buckets[j].FileList = cache[i].Buckets[j].FileList[:len(cache[i].Buckets[j].FileList)-1]
							break
						}
					}
					if len(cache[i].Buckets[j].FileList) == 0 {
						// If this value is set to 2 it means that the FreeCAD file contains a Gui description and we can
						// postprocess the content
						if cache[i].Buckets[j].FileEntry == 2 {
							currentRevision, _ := strconv.Atoi(cache[i].Buckets[j].Revision)
							currentRevision = currentRevision + 1
							cache[i].Buckets[j].Revision = strconv.Itoa(currentRevision)
							createBucketRevision(cache[i], cache[i].Buckets[j])

							// We can issue a preview build command here
							// that is done by calling the freecad API which is going to handle the queue
							FreeCADURI := os.Getenv("FREECAD_URI")
							FreeCADTCPPORT := os.Getenv("FREECAD_TCPPORT")

							type freecadEntry struct {
								Nickname      string
								Key           string
								SecretKey     string
								URI           string
								Port          string
								Dnsdomain     string
								MasterTCPPort string
								Bucket        string
								Revision      string
							}
							var data freecadEntry
							if dnsdomain != "" {
								data.Dnsdomain = dnsdomain
							} else {
								data.Dnsdomain = "https://127.0.0.1"
							}
							data.MasterTCPPort = "443"
							data.Nickname = cache[i].Nickname
							data.Key = cache[i].Key
							data.SecretKey = cache[i].SecretKey
							data.URI = cache[i].URI
							data.Port = cache[i].Port
							data.Bucket = cache[i].Buckets[j].Name
							data.Revision = strconv.Itoa(currentRevision - 1)
							content, _ := json.Marshal(data)
							base.HTTPPutRequest("http://"+FreeCADURI+FreeCADTCPPORT, content, "application/json")
						}
					}
					cache[i].entryMux.Unlock()
				}
			}
		}
	}
	return ""
}

// This is creating a user file entry

func createEntry(username string, content string, URI string) int {
	// The data are Marshalled
	// I must get the data info from the user and build a lookup table
	var ptr *minIOEntry
	var entry cacheEntry
	ptr = new(minIOEntry)
	json.Unmarshal([]byte(content), ptr)

	// We must get the user entry to get the Key !
	credentialsURI := os.Getenv("CREDENTIALS_URI")
	credentialsPort := os.Getenv("CREDENTIALS_TCPPORT")
	result := base.HTTPGetRequest("http://" + credentialsURI + credentialsPort + "/user/" + username + "/userGetInternalInfo")

	// We have to unmarshall the result
	var userPtr *base.User
	userPtr = new(base.User)
	json.Unmarshal([]byte(result), userPtr)

	entry.Key = userPtr.TokenAuth
	entry.SecretKey = userPtr.TokenSecret
	entry.Nickname = userPtr.Nickname
	u, _ := url.Parse("http://" + URI)
	entry.URI = u.Hostname()
	entry.Port = ptr.Ports
	cacheMux.Lock()
	cache = append(cache, entry)
	cacheMux.Unlock()
	return 1
}

func deleteEntry(username string, content string) int {
	return 1
}

func userCallback(w http.ResponseWriter, r *http.Request) {
	var username string
	path := strings.Split(r.URL.Path, "/")
	switch r.Method {
	case http.MethodGet:
		// Serve the resource.
		// I must return the content of the user file if it does exist otherwise
		// an error
		Key := path[2]
		function := path[3]
		switch function {
		case "getEntry":
			content := getEntry(Key)
			if content != "" {
				fmt.Fprint(w, content)
			} else {
				fmt.Fprintf(w, "Error")
			}
		case "getSecretKey":
			content := getSecretKey(Key)
			if content != "" {
				fmt.Fprint(w, content)
			} else {
				fmt.Fprintf(w, "Error")
			}
		case "getRevision":
			content := getRevision(Key, path[4])
			w.Write([]byte(content))
		default:
		}
	case http.MethodPut:
		if len(path) == 3 {
			// Update an existing record.
			username = path[2]
			createEntry(username, string(base.HTTPGetBody(r)), r.Host)
		}
		if len(path) == 5 {
			if path[4] == "FilesUpdate" {
				addFilesEntry(path[2], path[3], string(base.HTTPGetBody(r)))
			}
			if path[4] == "File" {
				deleteFile(path[2], path[3], string(base.HTTPGetBody(r)))
			}
		}
	case http.MethodDelete:
		if len(path) == 4 {
			deleteFilesEntry(path[2], path[3])
		} else {
			// NOT IMPLEMENTED YET
			deleteEntry(username, string(base.HTTPGetBody(r)))
		}
	default:
	}
}

func main() {
	print("=============================== \n")
	print("| Starting cacheServer backend|\n")
	print("| (c) 2019 CADCloud           |\n")
	print("| Development version -       |\n")
	print("| Private use only            |\n")
	print("=============================== \n")

	mux := http.NewServeMux()
	var CacheURI = os.Getenv("CACHE_URI")
	var CacheTCPPORT = os.Getenv("CACHE_TCPPORT")

	mux.HandleFunc("/user/", userCallback)

	log.Fatal(http.ListenAndServe(CacheURI+CacheTCPPORT, mux))
}
