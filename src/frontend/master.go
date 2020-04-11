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
    "net/http"
    "net/http/httputil"
    "net/url"
    "net"
    "os"
    "bytes"
    "log"
    "fmt"
    "base"
    "io/ioutil"
    "path"
    "time"
    "strings"
    "strconv"
    "crypto/tls"
    "crypto/hmac"
    "crypto/sha1"
    "encoding/base64"
    "html/template"
    "encoding/xml"
    "encoding/json"
    "golang.org/x/crypto/acme/autocert"
)

var staticAssetsDir = os.Getenv("STATIC_ASSETS_DIR")
var templatesDir = os.Getenv("TEMPLATES_DIR")
var tlsCertPath = os.Getenv("TLS_CERT_PATH")
var tlsKeyPath = os.Getenv("TLS_KEY_PATH")
var DNSDomain = os.Getenv("DNS_DOMAIN")
var certStorage = os.Getenv("CERT_STORAGE")

type Content struct {
    XMLName     xml.Name `xml:"Document"`
    Parts []Part `xml:"ObjectData>Object>Properties>Property>Part"`
}

type GuiContent struct {
    XMLName     xml.Name `xml:"Document"`
    GuiParts []GuiPart `xml:"ViewProviderData>ViewProvider>Properties>Property>ColorList"`
}


type Part struct {
    File    string   `xml:"file,attr"`
} 

type GuiPart struct {
    File    string   `xml:"file,attr"`
}

func checkAccess(w http.ResponseWriter, r *http.Request) (bool){
	var url = r.URL.Path
	var command string
	entries := strings.Split(strings.TrimSpace(url[1:]), "/") 

        var login string

	// The login is always accessible
	if ( len(entries) > 2 ) {
		command = entries[2]
		login = entries[1]
	} 
	switch command {
		case "getToken":
				if ( r.Method == http.MethodGet || r.Method == http.MethodPost ) {
					return true
				} else {
					return false
				}
		case "validateUser":
				return true
		case "resetPassword":
				return true
		case "generatePasswordLnkRst":
				return true
		case "createUser":
				return true
	}
        if ( r.Header.Get("Authorization") != "" ) {
		var method string
		switch r.Method {
			case http.MethodGet:
				method = "GET"
			case http.MethodPut:
				method = "PUT"
			case http.MethodPost:
				method = "POST"
			case http.MethodDelete:
				method = "DELETE"
		}
                // Is this an AWS request ?
                words := strings.Fields(r.Header.Get("Authorization"))
                if ( words[0] == "JYP" ) {
                        // Let's dump the various content
                        keys := strings.Split(words[1],":")
                        // We must retrieve the secret key used for encryption and calculate the header
                        // if everything is ok (aka our computed value match) we are good
                        cacheURI := os.Getenv("CACHE_URI")
                        cacheTCPPORT := os.Getenv("CACHE_TCPPORT")
                        result:=base.HTTPGetRequest("http://"+cacheURI+cacheTCPPORT+"/user/"+keys[0]+"/getSecretKey")

			// I am getting the Secret Key and the Nickname
                        stringToSign := method + "\n\n"+r.Header.Get("Content-Type")+"\n"+r.Header.Get("myDate")+"\n"+r.URL.Path
			datas := strings.Split(result," ")
			secretKey := datas[0]
			nickname := datas[1]
			if ( nickname != login ) {
				return false
			}
                        mac := hmac.New(sha1.New, []byte(secretKey))
                        mac.Write([]byte(stringToSign))
                        expectedMAC := mac.Sum(nil)
                        if ( base64.StdEncoding.EncodeToString(expectedMAC) == keys[1] ) {
				return true
                        }
                }
	}
	return false
}


// neuteredFileSystem is used to prevent directory listing of static assets
type neuteredFileSystem struct {
    fs http.FileSystem
}

func (nfs neuteredFileSystem) Open(path string) (http.File, error) {
    // Check if path exists
    f, err := nfs.fs.Open(path)
    if err != nil {
        return nil, err
    }

    // If path exists, check if is a file or a directory.
    // If is a directory, stop here with an error saying that file
    // does not exist. So user will get a 404 error code for a file/directory
    // that does not exist, and for directories that exist.
    s, err := f.Stat()
    if err != nil {
        return nil, err
    }
    if s.IsDir() {
        return nil, os.ErrNotExist
    }

    // If file exists and the path is not a directory, let's return the file
    return f, nil
}


func user(w http.ResponseWriter, r *http.Request) {
	if ( !checkAccess(w, r)  ) {
		w.Write([]byte("Access denied"))
		return
	}

	// parse the url
	u,_ := url.Parse("http://"+r.Host)
        host, _, _ := net.SplitHostPort(u.Host)
	url, _ := url.Parse("http://"+host+":9100")

	// create the reverse proxy
	proxy := httputil.NewSingleHostReverseProxy(url)

	// Update the headers to allow for SSL redirection
	r.URL.Host = "http://"+r.Host+":9100"

	r.Header.Set("X-Forwarded-Host", r.Header.Get("Host"))

	// Note that ServeHttp is non blocking and uses a go routine under the hood
	proxy.ServeHTTP(w , r)
}

func projects(w http.ResponseWriter, r *http.Request) {
        // parse the url
        projectURI := os.Getenv("PROJECT_URI")
        projectTCPPORT := os.Getenv("PROJECT_TCPPORT")

        url, _ := url.Parse("http://"+projectURI+projectTCPPORT)

        // create the reverse proxy
        proxy := httputil.NewSingleHostReverseProxy(url)

        // Update the headers to allow for SSL redirection
        r.URL.Host = "http://"+projectURI+projectTCPPORT

        r.Header.Set("X-Forwarded-Host", r.Header.Get("Host"))

        // Note that ServeHttp is non blocking and uses a go routine under the hood
        proxy.ServeHTTP(w , r)
}

func acl(w http.ResponseWriter, r *http.Request) {
	http.Redirect(
        	w, r,
	        "https://"+r.Host+r.URL.String(),
	        http.StatusMovedPermanently,
    	)
}

func ShiftPath(p string) (head, tail string) {
    p = path.Clean("/" + p)
    i := strings.Index(p[1:], "/") + 1
    if i <= 0 {
        return p[1:], "/"
    }
    return p[1:i], p[i:]
}

func home(w http.ResponseWriter, r *http.Request) {

	var b []byte
	var err error
	head,_ := ShiftPath( r.URL.Path)	


	if ( r.Header.Get("Authorization") != "" ) {
		// Is this an AWS request ?
		words := strings.Fields(r.Header.Get("Authorization"))
		if ( words[0] == "AWS" ) {
			var content []byte
			// we must reroute the request and let minIO answer to it
			// but first we must get the user account info 
			// based on the keyAccess

			// We must detect a pattern 
			// Bucket update is
			// GetBucketContent - CreateBucket or directly PUT files
			// We must catch up the Document.xml and DocumentGui.xml as to
			// parse there content and identify the full bucket content
			switch r.Method {
				case http.MethodGet:
					// Is this a bucket content directory request
				        path := strings.Split( r.URL.Path, "/" )
					if ( len(path) == 3 ) {
						// If yes we must empty the cache file
						keys := strings.Split(words[1],":")
						cacheURI := os.Getenv("CACHE_URI")
						cacheTCPPORT := os.Getenv("CACHE_TCPPORT")
						base.HTTPDeleteRequest("http://"+cacheURI+cacheTCPPORT+"/user/"+keys[0]+"/"+path[1])	
				        }
					break;
				default:

			}

			// This code is looking for the TCP port and the URL of the minio server corresponding
			// to the user defined with its public key
			// We excute there the original request

			keys := strings.Split(words[1],":")
			cacheURI := os.Getenv("CACHE_URI")
			cacheTCPPORT := os.Getenv("CACHE_TCPPORT")
			result:=base.HTTPGetRequest("http://"+cacheURI+cacheTCPPORT+"/user/"+keys[0]+"/getEntry")
		        url, _ := url.Parse("http://"+result)
			urlport := url.Port()
			urlhost := url.Hostname()
		        proxy := httputil.NewSingleHostReverseProxy(url)
		        r.URL.Host = "http://"+result
		        r.Header.Set("X-Forwarded-Host", r.Header.Get("Host"))

			// Must keep a copy of the body

			body := base.HTTPGetBody(r)

			realport,_ := strconv.Atoi(urlport)

			// I need to forward the same request to the ctrl part
			// as to have a full copy of the content
			// To do this I need to recompute the header as we will be changing the bucket name (could look like crazy)

                        myDate := time.Now().UTC().Format(http.TimeFormat)
                        myDate = strings.Replace(myDate, "GMT", "+0000", -1)
			var method string	
			switch r.Method {
				case http.MethodGet:
					method="GET"
				case http.MethodPut:
					method="PUT"
				case http.MethodDelete:
					method="DELETE"
			}

			// I must do these request only for PUT and DELETE operations. Get are handled by the original bucket
			if ( method == "PUT" || method == "DELETE" ) {

				// Must get the Revision first
				bucket :=  strings.Split(r.URL.Path,"/")
				result=base.HTTPGetRequest("http://"+cacheURI+cacheTCPPORT+"/user/"+keys[0]+"/getRevision/"+bucket[1])
	
				client := &http.Client{}

				// The Revision shall be added to the bucket name
				
				filePath  := strings.Split(r.URL.Path,"/")
				var fileName string
				fileName = ""
				if ( len(filePath) > 2 ) {
					fileName = filePath[2]
				}

				bucketName := filePath[1] +"r"+result
				var fullPath string
				fullPath = "/"+bucketName+"/"
				if ( fileName != "" ) {
					fullPath = fullPath + fileName
				}

			        proxyReq, _ := http.NewRequest(method,"http://"+urlhost+":"+strconv.Itoa(realport + 1000)+fullPath, bytes.NewReader(body))

	                        stringToSign := method + "\n\n"+r.Header.Get("Content-Type")+"\n"+myDate+"\n"+fullPath
			
		                words := strings.Fields(r.Header.Get("Authorization"))
	                        keys = strings.Split(words[1],":")
	                        // We must retrieve the secret key used for encryption and calculate the header
	                        // if everything is ok (aka our computed value match) we are good
	                        cacheURI = os.Getenv("CACHE_URI")
	                        cacheTCPPORT = os.Getenv("CACHE_TCPPORT")
	                        result=base.HTTPGetRequest("http://"+cacheURI+cacheTCPPORT+"/user/"+keys[0]+"/getSecretKey")

				datas := strings.Split(result, " ")

	                        mac := hmac.New(sha1.New, []byte(datas[0]))
	                        mac.Write([]byte(stringToSign))
	                        expectedMAC := mac.Sum(nil)
	                        signature:=base64.StdEncoding.EncodeToString(expectedMAC)


	                        proxyReq.Header.Set("Authorization","AWS "+keys[0]+":"+signature)
	                        proxyReq.Header.Set("Date", myDate)
				proxyReq.Header.Set("Content-Type", r.Header.Get("Content-Type"))

				// That is a new request so let's do it
				client.Do(proxyReq)

				// We have to update the cache after the write operation
				if ( r.Method == http.MethodPut ) {
                                        // WARNING: This stuff shall be executed only if the minio request is legitimate otherwise there will be
                                        // a security hole
                                        path := strings.Split( r.URL.Path, "/" )
                                        if ( path[2] == "Document.xml" ) {
                                                contents := Content{}
                                                in := bytes.NewReader([]byte(base.HTTPGetBody(r)))
                                                _ = xml.NewDecoder(in).Decode(&contents)
                                                content, _ = json.Marshal(contents.Parts)
                                        }
                                        if ( path[2] == "GuiDocument.xml" ) {
                                                contents := GuiContent{}
                                                in := bytes.NewReader([]byte(base.HTTPGetBody(r)))
                                                _ = xml.NewDecoder(in).Decode(&contents)
                                                content, _ = json.Marshal(contents.GuiParts)
                                        }
                                        // if content is not empty we have to push it to the cache server which will
                                        // be acting as a file tracker
                                        // each time a file is updated through a write operation we can remove it from the cache

                                        // if the cache becomes empty then we can trigger a post processing operation on the read only directory

                                        keys := strings.Split(words[1],":")
                                        cacheURI := os.Getenv("CACHE_URI")
                                        cacheTCPPORT := os.Getenv("CACHE_TCPPORT")

                                        if ( len(content) > 0 ) {
                                                // let's add all files
                                                _=base.HTTPPutRequest("http://"+cacheURI+cacheTCPPORT+"/user/"+keys[0]+"/"+path[1]+"/FilesUpdate",content,"application/json")
                                        } else {
                                                _=base.HTTPPutRequest("http://"+cacheURI+cacheTCPPORT+"/user/"+keys[0]+"/"+path[1]+"/File",[]byte(path[2]),"application/json")
                                        }
				}



			}

			// We must proxy serve
                        proxy.ServeHTTP(w , r)

			return
		}
		if ( words[0] == "AWS4-HMAC-SHA256" ) {
                        // The secrect key is within the first word but it needs to be processed
                        find_keys:=strings.Replace(words[1],"Credential=","",-1)
                        arguments:= strings.Split(find_keys,"/")
                        cacheURI := os.Getenv("CACHE_URI")
                        cacheTCPPORT := os.Getenv("CACHE_TCPPORT")
                        result:=base.HTTPGetRequest("http://"+cacheURI+cacheTCPPORT+"/user/"+arguments[0]+"/getEntry")
                        url, _ := url.Parse("http://"+result)
                        proxy := httputil.NewSingleHostReverseProxy(url)
                        r.URL.Host = "http://"+result
                        r.Header.Set("X-Forwarded-Host", r.Header.Get("Host"))
                        proxy.ServeHTTP(w , r)
                }
		if ( words[0] == "JYP" ) {
			switch ( head ) {
				case "js":
		                        serveJavascript( w,r)
		                case "css":
		                        serveCSS(w,r)
		                case "images":
		                        serveImages(w,r)
				case "user":
					user(w,r)
				case "projects":
					projects(w,r)
			}
		}
                return
	}	
	switch ( head ) {
		case "js":
			serveJavascript( w,r)
		case "css":
			serveCSS(w,r)
		case "images":
			serveImages(w,r)
		case "user":
			user(w,r)
		case "projects":
			projects(w,r)
		default:
			if ( head == "" ) {
				b, err = ioutil.ReadFile(staticAssetsDir+"homepage.html") // just pass the file name
				// this is a potential template file we need to replace the http field
				// by the calling r.Host
				t := template.New("my template")
				buf := &bytes.Buffer{}
				t.Parse(string(b))
				t.Execute(buf, r.Host)
				fmt.Fprintf(w, buf.String())
			} else {
				b, err = ioutil.ReadFile(staticAssetsDir+r.URL.Path) // just pass the file name
				w.Write(b)
			}
	
	}
    	if err != nil {
        	fmt.Print(err)
    	}
}

func serveJavascript(w http.ResponseWriter, r *http.Request) {
// we must serve the homepage.html content
        b, err := ioutil.ReadFile(staticAssetsDir+r.URL.Path) // just pass the file name
        if err != nil {
                fmt.Print(err)
        }
	w.Header().Set("Content-Type", "application/javascript")
	w.Write(b)
}

func serveCSS(w http.ResponseWriter, r *http.Request) {
// we must serve the homepage.html content
        b, err := ioutil.ReadFile(staticAssetsDir+r.URL.Path) // just pass the file name
        if err != nil {
                fmt.Print(err)
        }
        w.Header().Set("Content-Type", "text/css")
	w.Write(b)
}

func serveImages(w http.ResponseWriter, r *http.Request) {
// we must serve the homepage.html content
        b, err := ioutil.ReadFile(staticAssetsDir+r.URL.Path) // just pass the file name
        if err != nil {
                fmt.Print(err)
        }
        w.Header().Set("Content-Type", "image/png")
        w.Write(b)
}

// httpsRedirect redirects http requests to https
func httpsRedirect(w http.ResponseWriter, r *http.Request) {
    http.Redirect(
        w, r,
        "https://"+r.Host+r.URL.String(),
        http.StatusMovedPermanently,
    )
}

func main() {
    // http to https redirection
    print("=============================== \n")
    print("| Starting frontend           |\n")
    print("| (c) 2019 CADCloud team      |\n")
    print("| Development version -       |\n")
    print("| Private use only            |\n")
    print("=============================== \n")


    // before starting the services we must request for miniIO servers start / check / status


    // Serve static files while preventing directory listing
    
    mux := http.NewServeMux()

    // Highest priority must be set to the signed request

    // We are implementing our own router because we must redirect storage request
    // to the minio server

    mux.HandleFunc("/",home)


    if ( DNSDomain != "" ) {
	// if DNS_DOMAIN is set then we run in a production environment
	// we must get the directory where the certificates will be stored

	certManager := autocert.Manager{
		Prompt: autocert.AcceptTOS,
		Cache:  autocert.DirCache(certStorage),
		HostPolicy: autocert.HostWhitelist(DNSDomain),
    	}

	server := &http.Server{
		Addr:    ":443",
		Handler: mux,
                ReadTimeout:  600 * time.Second,
                WriteTimeout: 600 * time.Second,
                IdleTimeout:  120 * time.Second,
		TLSConfig: &tls.Config{
			GetCertificate: certManager.GetCertificate,
		},
    	}

        go func() {
        h := certManager.HTTPHandler(nil)
                log.Fatal(http.ListenAndServe(":http", h))
        }()

	server.ListenAndServeTLS("", "")
    } else {
	go http.ListenAndServe(":80", http.HandlerFunc(httpsRedirect))
    // Launch TLS server
   	log.Fatal(http.ListenAndServeTLS(":443", tlsCertPath, tlsKeyPath, mux))
   }
}
