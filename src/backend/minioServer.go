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
    "os"
    "net/http"
    "strings"
    "log"
    "fmt"
    "io/ioutil"
    "encoding/json"
    "strconv"
    "sync"
    "os/exec"
    "bytes"
    "base"
    "time"
    "net"
    "encoding/xml"
)

type minIOToken struct {
	TokenAuth string
	TokenSecret string
}

type minIOEntry struct {
        Username string
        Ports string
}

var file sync.RWMutex

// This variable contains the storage path name usable for minIO
var minIORoot = os.Getenv("MINIO_ROOT")


func deleteUserContent( username string, r *http.Request) {
        // That stuff need to get the bucket list of a user
        // Then it needs for each bucket to delete each object
        // and remove everybucket
        // I first need to get the user credential
        // As well as it's miniIO server IP / Port
        // get the BucketList
        // get for each bucket the content

	// WARNING: Could do it just by removing the home directory of the user within the minioStorage
	// We need to do that after we stopped the daemon

        var updatedData *base.User

        address := strings.Split(r.Host,":")
        data := base.HTTPGetRequest("http://"+address[0]+":9100"+"/user/"+username+"/userGetInternalInfo")
        updatedData=new(base.User)
        _=json.Unmarshal([]byte(data), updatedData)

	// We need to be sure that the minIO server is up before
	conn, err := net.Dial("tcp", updatedData.Server+":"+updatedData.Ports)
	for err != nil {
		time.Sleep(100 * time.Millisecond)
		conn, err = net.Dial("tcp", updatedData.Server+":"+updatedData.Ports)
	}
	conn.Close()


	response, err := base.Request("GET","http://"+updatedData.Server+":"+updatedData.Ports+"/", "", "application/xml", nil, "", updatedData.TokenAuth, updatedData.TokenSecret)

        if err != nil {
                log.Fatal(err)
        } else {
                defer response.Body.Close()
                contents, err := ioutil.ReadAll(response.Body)
                if err != nil {
                        log.Fatal(err)
                }
		// Got an XML content with the various Buckets created by the end user

                type Bucket struct {
                    Name    string   `xml:"Name"`
                }

		type Content struct {
		    XMLName     xml.Name `xml:"ListAllMyBucketsResult"`
		    Buckets []Bucket `xml:"Buckets>Bucket"`
		}


		XMLcontents := Content{}
                in := bytes.NewReader([]byte(contents))
                _ = xml.NewDecoder(in).Decode(&XMLcontents)

		// Ok Got the bucket list from the user


		for i := 0 ; i < len(XMLcontents.Buckets) ; i++ {

			response, err := base.Request("GET", "http://"+updatedData.Server+":"+updatedData.Ports+"/"+XMLcontents.Buckets[i].Name, "/"+XMLcontents.Buckets[i].Name, "application/xml", nil, "", updatedData.TokenAuth, updatedData.TokenSecret)


		        if err != nil {
		                log.Fatal(err)
		        } else {
				contents,_ := ioutil.ReadAll(response.Body)
				// Got the current bucket list content into contents as an XML format. Let's loop through it
				// and issue the delete command
				// and re-issue a get command to check if the directory is empty. As long as it is not
				// We reload the directory entries and we remove the file
				type object struct {
                		    Name    string   `xml:"Key"`
                		}
                		type bucketContent struct {
		                    XMLName     xml.Name `xml:"ListBucketResult"`
		                    Objects []object `xml:"Contents"`
                		}
				XMLbucketcontents := bucketContent{}
		                in := bytes.NewReader([]byte(contents))
		                _ = xml.NewDecoder(in).Decode(&XMLbucketcontents)
				for len(XMLbucketcontents.Objects) != 0 {
					for j := 0 ; j < len(XMLbucketcontents.Objects) ; j++ {
						// We must delete the object
						_, _ = base.Request("DELETE", "http://"+updatedData.Server+":"+updatedData.Ports+"/"+XMLcontents.Buckets[i].Name+"/"+XMLbucketcontents.Objects[j].Name,
								    "/"+XMLcontents.Buckets[i].Name+"/"+XMLbucketcontents.Objects[j].Name, "application/xml", nil, "", updatedData.TokenAuth, updatedData.TokenSecret)

					}

					response,_ = base.Request("GET", "http://"+updatedData.Server+":"+updatedData.Ports+"/"+XMLcontents.Buckets[i].Name, "/"+XMLcontents.Buckets[i].Name, "application/xml", nil, "",
								updatedData.TokenAuth, updatedData.TokenSecret)

					contents,_ := ioutil.ReadAll(response.Body)
					XMLbucketcontents = bucketContent{}
					in := bytes.NewReader([]byte(contents))
					_ = xml.NewDecoder(in).Decode(&XMLbucketcontents)
				}
				// We can now delete the Bucket
				_, _ = base.Request("DELETE", "http://"+updatedData.Server+":"+updatedData.Ports+"/"+XMLcontents.Buckets[i].Name, "/"+XMLcontents.Buckets[i].Name, "application/xml", nil, "", 
						    updatedData.TokenAuth, updatedData.TokenSecret)

			}

		}	

   		var minIOURI = os.Getenv("MINIO_URI")
	        var minIOTCPPORT = os.Getenv("MINIO_TCPPORT")
		freePort(minIOURI+minIOTCPPORT, username)
	        os.RemoveAll(minIORoot + "/" + username)
        }
        return


}


// This function is called when the User API has been started

func startMinIOServer(URI string, r *http.Request) {
        var minIOArray [base.MaxMinIOServer]minIOEntry
        var allocatedPort [base.MaxMinIOServer]int
        for  i := 0 ; i < base.MaxMinIOServer ; i++  {
                allocatedPort[i] = 0
                minIOArray[i].Username = ""
                minIOArray[i].Ports = ""
        }
        _, err := os.Stat(minIORoot + "/" + "master"+URI+".json")
        if ( ! os.IsNotExist(err) ) {
                // The config file exist we have to read it and find the first available Port
                b,_ := ioutil.ReadFile(minIORoot + "/" + "master"+URI+".json")
                json.Unmarshal([]byte(b),&minIOArray)

                // Initial check - if the username exist, we try to launch the minIO server
                for  i := 0 ; i < base.MaxMinIOServer ; i++  {
                        if ( minIOArray[i].Username != "" ) {
				// Ok the entry is configured, we must start the minIO server for that user
		                value, _ := strconv.Atoi(minIOArray[i].Ports)
       				realTCPPort := value + base.MinIOServerBasePort
		                s := strconv.Itoa(realTCPPort)
		                sCtrl := strconv.Itoa(realTCPPort+1000)
				address := strings.Split(URI,":")
				credentialsURI:=os.Getenv("CREDENTIALS_URI")
				credentialsPort:=os.Getenv("CREDENTIALS_TCPPORT")
				result:=base.HTTPGetRequest("http://"+credentialsURI+credentialsPort+"/user/"+minIOArray[i].Username+"/userGetInternalInfo")
				// We have to unmarshall the result
                                var userPtr *base.User
                                userPtr=new(base.User)
                                json.Unmarshal([]byte(result),userPtr)

                                os.Setenv("MINIO_ACCESS_KEY", userPtr.TokenAuth)
                                os.Setenv("MINIO_SECRET_KEY", userPtr.TokenSecret)
                                os.Setenv("MINIO_BROWSER", "off")
				command := "minio server --address "+address[0]+":"+ s +" "+ minIORoot + "/" + minIOArray[i].Username
				commandCtrl := "minio server --address "+address[0]+":"+ sCtrl +" "+ minIORoot + "ctrl/" + minIOArray[i].Username

		                // before starting the server we must be checking if it is soon running
		                // to do that we must look for the command into the process table

		                args := []string { "-o", "command" }
		                cmd := exec.Command("ps", args...)
		                var out bytes.Buffer
		                cmd.Stdout = &out
		                cmd.Start()
				cmd.Wait()

		                if ( !strings.Contains(out.String(), command) ) {
		                        // Second parameter shall be a string array
		                        args = []string { "server", "--address" }
		                        args = append (args, address[0]+":" + s)
		                        args = append (args, minIORoot + "/" + minIOArray[i].Username)
		                        cmd := exec.Command("minio", args...)
		                        cmd.Start()
					done := make(chan error, 1)
		                        go func() {
		                            done <- cmd.Wait()
		                        }()
		                }

                                if ( !strings.Contains(out.String(), commandCtrl) ) {
                                        // Second parameter shall be a string array
                                        args = []string { "server", "--address" }
                                        args = append (args, address[0]+":" + sCtrl)
                                        args = append (args, minIORoot + "ctrl/" + minIOArray[i].Username)
                                        cmd := exec.Command("minio", args...)
                                        cmd.Start()
                                        done := make(chan error, 1)
                                        go func() {
                                            done <- cmd.Wait()
                                        }()
                                }

				// We must check if the ctrl bucket is properly created or not
				// if not we must create it

				fullPath := "/ctrl/"
				method := "GET"

				response, err := base.Request(method, "http://"+address[0]+":"+sCtrl+fullPath, fullPath, "application/octet-stream", nil, "",
                                                    userPtr.TokenAuth, userPtr.TokenSecret)
				
                                for err != nil {
                                        time.Sleep(1*time.Second)
                                        response,err = base.Request(method, "http://"+address[0]+":"+sCtrl+fullPath, fullPath, "application/octet-stream", nil, "",
							userPtr.TokenAuth, userPtr.TokenSecret)
                                }
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


			                type Content struct {
			                    XMLName     xml.Name `xml:"Error"`
			                    Buckets string `xml:"Code"`
			                }


			                XMLcontents := Content{}
			                in := bytes.NewReader([]byte(contents))
			                _ = xml.NewDecoder(in).Decode(&XMLcontents)

			                // Ok Got the bucket list from the user

			                if ( XMLcontents.Buckets == "NoSuchBucket" ) {


				                fullPath := "/ctrl/"

				                method := "PUT"

						_,_ = base.Request(method, "http://"+address[0]+":"+sCtrl+fullPath, fullPath, "application/octet-stream", nil, "", userPtr.TokenAuth, userPtr.TokenSecret)

					}
			         }

				// We must inform the cache API that the minIO server for
				// the current user has been started
				// we must POST the IP of the server as well as the port number
				// we just have to Marshal the minIOArray[i] as the IP will be known through the
				// r.host of the calling server
				content, _ := json.Marshal(minIOArray[i])
				var CacheURI = os.Getenv("CACHE_URI")
				var CacheTCPPORT = os.Getenv("CACHE_TCPPORT")
			        base.HTTPPutRequest("http://"+CacheURI+CacheTCPPORT+"/user/"+minIOArray[i].Username,content,"application/json")
                        }
                }
	}
}

func freePort(URI string, username string) {
        var minIOArray [base.MaxMinIOServer]minIOEntry
        var allocatedPort [base.MaxMinIOServer]int
        for  i := 0 ; i < base.MaxMinIOServer ; i++  {
                allocatedPort[i] = 0
                minIOArray[i].Username = ""
                minIOArray[i].Ports = ""
        }
        // That stuff must be synced
        file.Lock()
        defer file.Unlock()
        // We must open the Master configuration file
        print(minIORoot + "/" + "master"+URI+".json" + "\n")
        _, err := os.Stat(minIORoot + "/" + "master"+URI+".json")
        if ( ! os.IsNotExist(err) ) {
                // The config file exist we have to read it and find the first available Port
                b,_ := ioutil.ReadFile(minIORoot + "/" + "master"+URI+".json")
                json.Unmarshal([]byte(b),&minIOArray)

                // Initial check - if the username exist, we return the existing port
                for  i := 0 ; i < base.MaxMinIOServer ; i++  {
                        if ( minIOArray[i].Username == username ) {
				minIOArray[i].Username = ""
                                minIOArray[i].Ports = ""
                        }
                }
                // we must Marshall the data and rewrite the file
                content, _ := json.Marshal(minIOArray)
                _ = ioutil.WriteFile(minIORoot + "/" + "master"+URI+".json", []byte(content), os.ModePerm)
	}

}


func getNewPort(URI string, username string) (string) {
	var minIOArray [base.MaxMinIOServer]minIOEntry
	var allocatedPort [base.MaxMinIOServer]int
	for  i := 0 ; i < base.MaxMinIOServer ; i++  {
		allocatedPort[i] = 0
		minIOArray[i].Username = ""
		minIOArray[i].Ports = ""
	}
	// That stuff must be synced
	file.Lock()
	defer file.Unlock()
	// We must open the Master configuration file
	print(minIORoot + "/" + "master"+URI+".json" + "\n")
        _, err := os.Stat(minIORoot + "/" + "master"+URI+".json")
        if ( ! os.IsNotExist(err) ) {
                // The config file exist we have to read it and find the first available Port
                b,_ := ioutil.ReadFile(minIORoot + "/" + "master"+URI+".json")
                json.Unmarshal([]byte(b),&minIOArray)

		// Initial check - if the username exist, we return the existing port
		for  i := 0 ; i < base.MaxMinIOServer ; i++  {
                        if ( minIOArray[i].Username == username ) {
				return minIOArray[i].Ports
			}
		}

		for  i := 0 ; i < base.MaxMinIOServer ; i++  {
			if ( minIOArray[i].Username != "" ) {
				value,_ := strconv.Atoi(minIOArray[i].Ports)
				allocatedPort[value]=1
			}
        	}

		// we must find an available port
		// The test shouldn't be != 1 as it is initialized to 0
		availablePort := -1
		for i := 0 ; i < base.MaxMinIOServer ; i++  {
			if ( allocatedPort[i] != 1 ) {
				availablePort = i
				break
			} 
		}
		if ( availablePort == -1 ) {
			// No Port available we must return an error
			return "error"
		}
		// We found a Port
		// we can create the entry into the array and save it as a JSON structure
		i:=0
		for minIOArray[i].Username != "" {
			i++
		}
		minIOArray[i].Username = username
		minIOArray[i].Ports = strconv.Itoa(availablePort)

		// we must Marshall the data and rewrite the file
	        content, _ := json.Marshal(minIOArray)
	        _ = ioutil.WriteFile(minIORoot + "/" + "master"+URI+".json", []byte(content), os.ModePerm)
		return string(strconv.Itoa(availablePort))

        } else
        {
		// the Port will be 9400 and we must create the entry into the configuration file
		minIOArray[0].Username = username
                minIOArray[0].Ports = "00"
		content, _ := json.Marshal(minIOArray)
                _ = ioutil.WriteFile(minIORoot + "/" + "master"+URI+".json", []byte(content), os.ModePerm)
                return "0"
        }	
	return ""
}

func createMinIOServer(username string, URL string, accessToken string, secretToken string) (string) {
	// We shall have as input parameters the AccessToken and SecretToken which 
	// will be used to spin off the minIO server configuration
	// First I must look for an available port

	var entry base.MinIOServer

	availablePort := getNewPort(URL, username)
	
	if ( availablePort != "" ) {
	
		// I got a port
		// We must compute the real port
		value, _ := strconv.Atoi(availablePort)
		realTCPPort := value + base.MinIOServerBasePort
		// I must retrieve the User AccessToken and SecretToken
		// then I can spawn a minIO Server for that user
		os.Setenv("MINIO_ACCESS_KEY", accessToken)
		os.Setenv("MINIO_SECRET_KEY", secretToken)
		os.Setenv("MINIO_BROWSER", "off")
		// we must create the user directory
		os.Mkdir(minIORoot + "/" + username, os.ModePerm)
		// we must create the ctrl area
		os.Mkdir(minIORoot + "ctrl/" + username, os.ModePerm)
		s := strconv.Itoa(realTCPPort)
		// This is the ctrl TCP Port
		sCtrl := strconv.Itoa(realTCPPort+1000)

                address := strings.Split(URL,":")	

		command := "minio server --address "+ address[0] +":"+ s +" "+ minIORoot + "/" + username
		commandCtrl := "minio server --address "+ address[0] +":"+ sCtrl +" "+ minIORoot + "ctrl/" + username

		// before starting the server we must be checking if it is soon running
	        // to do that we must look for the command into the process table

		args := []string { "-o", "command" }
		cmd := exec.Command("ps", args...)
		var out bytes.Buffer
		cmd.Stdout = &out
		cmd.Start()
		cmd.Wait()

		// We start the User MiniIO daemon

		if ( !strings.Contains(out.String(), command) ) {
			// Second parameter shall be a string array
			args = []string { "server", "--address" }
			args = append (args, address[0] + ":" + s)
			args = append (args, minIORoot + "/" + username)
			cmd := exec.Command("minio", args...)
			cmd.Start()
			done := make(chan error, 1)
			go func() {
			    done <- cmd.Wait()
			}()
		}

		// We start the Ctrl MiniIO daemon

                if ( !strings.Contains(out.String(), commandCtrl) ) {
                        // Second parameter shall be a string array
                        args = []string { "server", "--address" }
                        args = append (args, address[0] + ":" + sCtrl)
                        args = append (args, minIORoot + "ctrl/" + username)
                        cmd := exec.Command("minio", args...)
                        cmd.Start()
                        done := make(chan error, 1)
                        go func() {
                            done <- cmd.Wait()
                        }()
                }


		// we must create the ctrl bucket which is used to store previews etc ...
		// We must be sure that the minio server is started ...

                fullPath := "/ctrl/"

                method := "PUT"

                response,err := base.Request(method, "http://"+address[0]+":"+sCtrl+fullPath, fullPath, "application/octet-stream", nil, "", accessToken, secretToken)

		for err != nil {
                         time.Sleep(1*time.Second)
                         response,err = base.Request(method, "http://"+address[0]+":"+sCtrl+fullPath, fullPath, "application/octet-stream", nil, "", accessToken, secretToken)
                }

		init := 0
		// We can move forward
		for init != 1 {
			defer response.Body.Close()
			contents, err := ioutil.ReadAll(response.Body)
			if err != nil {
				log.Fatal(err)
			}
			// Got an XML content with the various Buckets created by the end user
			type Content struct {
				XMLName     xml.Name `xml:"Error"`
				Code string `xml:"Code"`
			}
			XMLcontents := Content{}
			in := bytes.NewReader([]byte(contents))
			err = xml.NewDecoder(in).Decode(&XMLcontents)
			if ( err == nil ) {
				if ( XMLcontents.Code == "XMinioServerNotInitialized" ) {
					time.Sleep(1*time.Second)
					response,err = base.Request(method, "http://"+address[0]+":"+sCtrl+fullPath, fullPath, "application/octet-stream", nil, "", accessToken, secretToken)
				} 
			} else {
				init = 1
			}
		}

		// We need to return some information like
		// the minIO IP address, the TCPPORT
		// as to properly configure the reverse proxy and route the traffic to it
		// We also need to implement the user loopback as to configure the reverse proxy
		entry.Port = s
		// we must split the URL as it does contain the port
		entry.URI = address[0]
		data,_ := json.Marshal(entry)


		type minIOEntry struct {
		        Username string
		        Ports string
		}

		var localEntry minIOEntry

		localEntry.Username = username
		localEntry.Ports =  availablePort

                content, _ := json.Marshal(localEntry)
                var CacheURI = os.Getenv("CACHE_URI")
                var CacheTCPPORT = os.Getenv("CACHE_TCPPORT")
                base.HTTPPutRequest("http://"+CacheURI+CacheTCPPORT+"/user/"+username,content,"application/json")


		return string(data)
	} else {
		return "error"
	}
	return "error"
}

func stopServer( username string, r *http.Request) {
	mycontent,_ := ioutil.ReadAll(r.Body)
	var myuser base.User
        json.Unmarshal([]byte(mycontent),&myuser)
	serverIP := myuser.Server
        TCPPort := myuser.Ports

	// We must check if the daemon is running
	command := "minio server --address "+ serverIP + ":"+ TCPPort +" "+ minIORoot + "/" + username
	Port,_ := strconv.Atoi(TCPPort)
	CtrlPort := strconv.Itoa(Port+1000)
	commandCtrl := "minio server --address "+ serverIP + ":"+CtrlPort +" "+ minIORoot + "ctrl/" + username

        args := []string { "-o", "pid,command" }
        cmd := exec.Command("ps", args...)
        var out bytes.Buffer
        cmd.Stdout = &out
        cmd.Start()
        cmd.Wait()
	// Must find the PID
        stringArray := strings.Split(out.String(),"\n")
        for i := 0 ; i < len(stringArray)-1 ; i++ {
		localCommand := strings.SplitN(strings.TrimSpace(stringArray[i])," ",2)
		if ( localCommand[1] == command ) {
			pid := localCommand[0]
			// We must issue a SIGINT to that PID to stop it
			args := []string { "-SIGINT", pid }
			cmd = exec.Command("kill", args...)
			var out bytes.Buffer
		        cmd.Stdout = &out
		        cmd.Start()
		        cmd.Wait()
		}
		if ( localCommand[1] == commandCtrl ) {
                        pid := localCommand[0]
                        // We must issue a SIGINT to that PID to stop it
                        args := []string { "-SIGINT", pid }
                        cmd = exec.Command("kill", args...)
                        var out bytes.Buffer
                        cmd.Stdout = &out
                        cmd.Start()
                        cmd.Wait()
                }
	}
}


func userCallback(w http.ResponseWriter, r *http.Request) {
	var url = r.URL.Path
        var username string
	var command string
	var myminioToekn minIOToken
	// Is there a command ?
	entries := strings.Split(url[1:], "/")
        // The login is always accessible
        if ( len(entries) > 2 ) {
                command = entries[2]
        } else {
		command = ""
        }
        username = entries[1]
        switch r.Method {
                case http.MethodGet:
			// Get is providing the TCPPORT of the user
			// and the IP address of the minIO server attached to it
			// normally this is a static allocation
			// but ...
                case http.MethodPut:
			switch command {
				case "stopServer":
					stopServer(username, r)
				default:
					// The content is within the body
					mycontent,_ := ioutil.ReadAll(r.Body)
		                	json.Unmarshal([]byte(mycontent),&myminioToekn)
		
					accessToken := myminioToekn.TokenAuth
					secretToken := myminioToekn.TokenSecret

					// request for a new entry
					// the parameters are the access key and the secret key
					// this is safe to get them as parameter as we are not running
					// on a public network

		                        // Update an existing record.

					// WARNING the r.host shall be replaced by an ALLOCATION ALGORITHM
					// to determine on which server we can allocate the storage for the user

					var minIOURI = os.Getenv("MINIO_URI")
    					var minIOTCPPORT = os.Getenv("MINIO_TCPPORT")
		                        userParameters := createMinIOServer(username, minIOURI+minIOTCPPORT, accessToken, secretToken )
					fmt.Fprint(w, userParameters)
				}
                case http.MethodDelete:
				deleteUserContent(username, r)
                default:
        }
}

func startCallback(w http.ResponseWriter, r *http.Request) {
    var minIOURI = os.Getenv("MINIO_URI")
    var minIOTCPPORT = os.Getenv("MINIO_TCPPORT")
    startMinIOServer(minIOURI+minIOTCPPORT, r )
}


func main() {
    print("================================================ \n")
    print("| Starting minIO storage allocation backend    |\n")
    print("| (c) 2019 CADCloud                            |\n")
    print("| Development version -                        |\n")
    print("| Private use only                             |\n")
    print("================================================ \n")

    mux := http.NewServeMux()
    var minIOURI = os.Getenv("MINIO_URI")
    var minIOTCPPORT = os.Getenv("MINIO_TCPPORT")

    if _, err := os.Stat(minIORoot); os.IsNotExist(err) {
    	os.MkdirAll(minIORoot, os.ModePerm)
    }
    // We have to start configured server and report there existence to the 
    // main caching server

    mux.HandleFunc("/user/", userCallback)
    mux.HandleFunc("/start/", startCallback)

    log.Fatal(http.ListenAndServe(minIOURI+minIOTCPPORT, mux))
}

