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
    "base"
    "net/http"
    "log"
    "sync"
    "fmt"
    "encoding/json"
    "os/exec"
    "io/ioutil"
    "time"
    "bytes"
    "encoding/xml"
    "strings"
    "strconv"
    "crypto/hmac"
    "encoding/base64"
    "crypto/sha1"
    "net/url"
    "bufio"
)

var ProjectTMP = os.Getenv("PROJECT_TEMP")
var ProjectURI = os.Getenv("PROJECT_URI")
var ProjectMinIOPort = os.Getenv("PROJECT_MINIO_TCPPORT")

var file sync.RWMutex

type projectEntry struct {
        Key string
        SecretToken string
}

type freecadEntry struct {
         Nickname string
         Key string
         SecretKey string
         URI string
         Port string
         Bucket string
         Revision string
}

var data projectEntry

func checkAccess(w http.ResponseWriter, r *http.Request, login string) (bool){

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


func createEntry(content string) {
	// So we need to store the data 
	// I must push that to the minio backend
        var dataFreeCAD freecadEntry
	_ = json.Unmarshal([]byte(content), &dataFreeCAD)


	// We can push the entry to the project management system

	 t := time.Now()
         formatted := fmt.Sprintf("%d-%02d-%02d",
                       t.Year(), t.Month(), t.Day())

        fullPath := "/public/"+formatted+"/"+dataFreeCAD.Nickname+"/"+dataFreeCAD.Bucket+"r"+dataFreeCAD.Revision+".json"

        method := "PUT"

        _, _ = base.Request(method, "http://"+ProjectURI+ProjectMinIOPort+fullPath, fullPath, "application/xml", []byte(content), "", data.Key, data.SecretToken)


}

func moveEntry(r *http.Request, content string) {
	// Javascript is sending us '/projects/public/'+Date+'/'+Owner+'/'+Name+'/' into the URL
	// And the Body contains the Revisions numbers
	var revisions []string
	var dates []string
	jsonEntries := strings.Split(content, "\n")
	_=json.Unmarshal([]byte(jsonEntries[0]), &revisions)
	_=json.Unmarshal([]byte(jsonEntries[1]), &dates)
	path := r.URL.Path

	// path is giving us the target
	// So the source is the opposite

	words := strings.Split(path, "/")
	target := words[2]
	source := ""
	private := 1
	switch target {
		case "public":
			source = "private"
			private = 1
		case "private":
			source = "public"
			private = 0
	}	

        contents := getJSONEntry(r.URL.Path, private)

        // I must put the content into the right structure
        var dataFreeCAD freecadEntry
        _ = json.Unmarshal([]byte(contents), &dataFreeCAD)


	for i := 0 ; i < len(revisions) ; i++ {
		contents := getJSONEntry("/"+words[1]+"/"+words[2]+"/"+dates[i]+"/"+words[4]+"/"+words[5]+"/"+revisions[i], private)

	        t := time.Now()
	        formatted := fmt.Sprintf("%d-%02d-%02d",
                       t.Year(), t.Month(), t.Day())

	        fullPath := "/"+target+"/"+formatted+"/"+words[4]+"/"+words[5]+"r"+revisions[i]+".json"

	        method := "PUT"

	        _, _ = base.Request(method, "http://"+ProjectURI+ProjectMinIOPort+fullPath, fullPath, "application/xml", []byte(contents), "", data.Key, data.SecretToken)



		// Now let's delete the source

                t = time.Now()
                formatted = fmt.Sprintf("%d-%02d-%02d",
                       t.Year(), t.Month(), t.Day())

                fullPath = "/"+source+"/"+dates[i]+"/"+words[4]+"/"+words[5]+"r"+revisions[i]+".json"

                method = "DELETE"

	        _, _ = base.Request(method, "http://"+ProjectURI+ProjectMinIOPort+fullPath, fullPath, "application/xml", nil, "", data.Key, data.SecretToken)

        }


}

func getList(username string) string{

	// The default is to return a project list

        fullPath := "/public/"

        method := "GET"

       q := url.Values{}
       q.Add("list-type", "2")
       q.Add("max-keys", "1000")


        // That is a new request so let's do it
        var response *http.Response

	response,_ = base.Request(method, "http://"+ProjectURI+ProjectMinIOPort+fullPath, fullPath, "application/xml", nil, q.Encode(), data.Key, data.SecretToken)


	defer response.Body.Close()
        contents, _ := ioutil.ReadAll(response.Body)
	type Code struct {
                XMLName   xml.Name `xml:"ListBucketResult"`
                Keys []string `xml:"Contents>Key"`
		NextContinuationToken string `xml:"NextContinuationToken"`
		IsTruncated string `xml:"IsTruncated"`
        }

        XMLcontents := Code{}
        in := bytes.NewReader([]byte(contents))
        _ = xml.NewDecoder(in).Decode(&XMLcontents)
	var output string

	type projectEntry struct {
		Owner string
		Name string
		Private int
		Date []string
		Revisions []string
	}

	var projectList []projectEntry
	
	// WARNING: THIS REQUEST CAN BE EXTREMELY SLOW WHEN PROJECTS NUMBER WILL INCREASE


	for {

		for i := 0 ; i < len(XMLcontents.Keys) ; i++ {
			entry := strings.Split(XMLcontents.Keys[i], "/")
			if ( len(entry) == 3 ) {
				if ( username != "" ) {
					if ( entry[1] != username ) {
						continue
					}
				}
				realName := entry[2]
				realName = strings.TrimSuffix(realName, ".json")	
				suffixIndex := strings.LastIndex(realName, "r")
				revision := realName[suffixIndex+1:]
				realName = realName[0:suffixIndex]
				var index int
				index = -1
				// Is it into the array ?
				for j := 0 ; j < len(projectList) ; j++ {
					if projectList[j].Name == realName  && projectList[j].Owner == entry[1] {
						index = j
					}
				}
				if  index == -1 {
					var newprojectEntry projectEntry
					newprojectEntry.Date = append(newprojectEntry.Date, entry[0])
					newprojectEntry.Owner = entry[1]
					newprojectEntry.Name = realName
					newprojectEntry.Private = 0
					newprojectEntry.Revisions=append(newprojectEntry.Revisions,revision)
					projectList = append(projectList,newprojectEntry)
				} else {
					projectList[index].Date = append(projectList[index].Date, entry[0])
					projectList[index].Revisions=append(projectList[index].Revisions,revision)	
				}
	
			}
		}
	
		if ( XMLcontents.IsTruncated == "true" ) {
			// We must pursue and load the remaining part of the object list
		        fullPath = "/public/"
	
		        method = "GET"
	
	
		       q := url.Values{}
		       q.Add("list-type", "2")
		       q.Add("max-keys", "1000")
		       q.Add("continuation-token", XMLcontents.NextContinuationToken)
	
			response,_ = base.Request(method, "http://"+ProjectURI+ProjectMinIOPort+fullPath, fullPath, "application/xml", nil, q.Encode(), data.Key, data.SecretToken)

		        defer response.Body.Close()
		        contents, _ = ioutil.ReadAll(response.Body)
	
		        XMLcontents = Code{}
		        in := bytes.NewReader([]byte(contents))
		        _ = xml.NewDecoder(in).Decode(&XMLcontents)
		} else {
			break
		}

	}


	// Must do the same thing for the private project of the user

	if ( username != "" ) {


	        fullPath = "/private/"

	        method = "GET"

		response,_ = base.Request(method, "http://"+ProjectURI+ProjectMinIOPort+fullPath, fullPath, "application/xml", nil, "",  data.Key, data.SecretToken)


	        defer response.Body.Close()
	        contents, _ = ioutil.ReadAll(response.Body)

	        XMLcontents = Code{}
	        in = bytes.NewReader([]byte(contents))
	        _ = xml.NewDecoder(in).Decode(&XMLcontents)

	        for i := 0 ; i < len(XMLcontents.Keys) ; i++ {
	                entry := strings.Split(XMLcontents.Keys[i], "/")
	                if ( len(entry) == 3 ) {
	                        if ( entry[1] != username ) {
	                                continue
	                        }
	                        realName := entry[2]
	                        realName = strings.TrimSuffix(realName, ".json")
	                        suffixIndex := strings.LastIndex(realName, "r")
	                        revision := realName[suffixIndex+1:]
	                        realName = realName[0:suffixIndex]
	    	                var index int
	                        index = -1
	                        // Is it into the array ?
	                        for j := 0 ; j < len(projectList) ; j++ {
	                                if projectList[j].Name == realName  && projectList[j].Owner == entry[1] {
	                                        index = j
	                                }
	                        }
	                        if  index == -1 {
	                                var newprojectEntry projectEntry
	                                newprojectEntry.Date = append(newprojectEntry.Date, entry[0])
	                                newprojectEntry.Owner = entry[1]
	                                newprojectEntry.Name = realName
					newprojectEntry.Private = 1
	                                newprojectEntry.Revisions=append(newprojectEntry.Revisions,revision)
	                                projectList = append(projectList,newprojectEntry)
	                        } else {
	                                projectList[index].Date = append(projectList[index].Date, entry[0])
	                                projectList[index].Revisions=append(projectList[index].Revisions,revision)
	                        }
	
	                }
	        }

	}
	

        output = "{ \"Entries\" : ["
        for i := 0 ; i < len(projectList) ; i++ {
		output = output + "{"
		output = output + "\"Name\" : \""+ projectList[i].Name +"\" , "
		output = output + "\"Owner\" : \""+ projectList[i].Owner +"\" , "
		output = output + "\"Private\" : \""+ strconv.Itoa(projectList[i].Private) +"\" , "
		output = output + "\"Date\" : [" 
		
		for j := 0 ; j < len(projectList[i].Date) ; j++ {
                        output = output + "\""+ projectList[i].Date[j]+"\""
                        if ( j < len(projectList[i].Date) - 1 ) {
                                output = output + ","
                        }
                }

		output = output + "], "
		output = output + "\"Revisions\" : " 

		output = output + "["
		for j := 0 ; j < len(projectList[i].Revisions) ; j++ {
			output = output + "\""+ projectList[i].Revisions[j]+"\""
			if ( j < len(projectList[i].Revisions) - 1 ) {
				output = output + ","
			}
		}
		output = output + "]"
		output = output + "}" 
		if ( i < len(projectList)-1 ) {
			output = output + ","
		}
        }
        output = output +"]}"

	return output
}

func getJSONEntry(Path string, private int) (string) {
	var keyWords []string
        keyWords = strings.Split(Path, "/")
        targetUrl := keyWords[3] + "/" + keyWords[4] + "/" + keyWords[5] + "r" + keyWords[6] + ".json"

        // We must load the json file to know where the data are stored before forwarding them
        // to the client

	// That stuff works only if the project is public
	// Is it private ?
	fullPath := ""	
	if ( private == 1 ) {
	        fullPath = "/private/" + targetUrl
	} else {
        	fullPath = "/public/" + targetUrl
	}

        method := "GET"

        response, _ := base.Request(method, "http://"+ProjectURI+ProjectMinIOPort+fullPath, fullPath, "application/xml", nil, "", data.Key, data.SecretToken)


        defer response.Body.Close()
        contents, _ := ioutil.ReadAll(response.Body)

	return string(contents)

}

func getMagnet(w http.ResponseWriter, Path string, private int) {

	contents := getJSONEntry(Path, private)
	// I must put the content into the right structure
        var dataFreeCAD freecadEntry
        _ = json.Unmarshal([]byte(contents), &dataFreeCAD)

	// We must get the picture !

        fullPath := "/ctrl/"+dataFreeCAD.Bucket + "r" + dataFreeCAD.Revision + ".png"

        method := "GET"
	realPort,_ := strconv.Atoi(dataFreeCAD.Port)
	realPort = realPort + 1000 + base.MinIOServerBasePort

        response, _ := base.Request(method, "http://"+dataFreeCAD.URI+":"+strconv.Itoa(realPort)+fullPath, fullPath, "application/octet-stream", nil, "", dataFreeCAD.Key, dataFreeCAD.SecretKey)



        defer response.Body.Close()
	content, _ := ioutil.ReadAll(response.Body)
	w.Write([]byte(base64.StdEncoding.EncodeToString(content)))

}

func getAvatar(w http.ResponseWriter, path string, private int) {

	contents := getJSONEntry(path, private)

        // I must put the content into the right structure
        var dataFreeCAD freecadEntry
        _ = json.Unmarshal([]byte(contents), &dataFreeCAD)

        // We must get the picture !

        client := &http.Client{}

        myDate := time.Now().UTC().Format(http.TimeFormat)
        myDate = strings.Replace(myDate, "GMT", "+0000", -1)

        fullPath := "/user/"+dataFreeCAD.Nickname + "/getAvatar" 

        method := "GET"

        realPort,_ := strconv.Atoi(dataFreeCAD.Port)
        realPort = realPort + 1000 + base.MinIOServerBasePort

	URI := os.Getenv("CREDENTIALS_URI")
	TCPPORT := os.Getenv("CREDENTIALS_TCPPORT")

        req, _ := http.NewRequest(method,"http://"+URI+TCPPORT+fullPath, nil)

        // That is a new request so let's do it
        response, _ := client.Do(req)

        defer response.Body.Close()
        content, _ := ioutil.ReadAll(response.Body)
	// I don't need to encode it this is soon made by the users API
        w.Write([]byte(content))

}

type localstring struct {
	Value   string  `xml:"value,attr"`
}

type linklist struct {
	Entry   string  `xml:"value,attr"`
}

type Property struct {
	Name    string   `xml:"name,attr"`
	Strings  []localstring   `xml:"String"`
	LinkList []linklist `xml:"LinkList>Link"`
}

type Content struct {
	Name     string `xml:"name,attr"`
	Properties []Property `xml:"Properties>Property"`
}

type Object struct {
	XMLName     xml.Name `xml:"Document"`
	Objects []Content `xml:"ObjectData>Object"`
}

type Translate struct {
	FreeCADName string
	ModelName string
	Tag int
        hasShape int
}

func getIndex(label string, tree Object) []int {
	returnValue := []int { 0,0 }
	for i:= 0 ; i < len(tree.Objects) ; i++ {
		if ( tree.Objects[i].Name == label ) {
		// We have 2 options either this is a Part or an Assembly
		// If that is an assembly we must look for the Group property
		// If that is a Part we must look for the Label property
			for j:=0 ; j < len(tree.Objects[i].Properties) ; j++  {
	                        if ( tree.Objects[i].Properties[j].Name == "Group" ) {
					returnValue[1] = j
				}
			}
			returnValue[0] = i
			return returnValue
		}
	}
	return returnValue
}

func label(name string, labels []Translate) (string) {
	for i := 0 ; i < len(labels) ; i++ {
		if ( labels[i].FreeCADName == name ) {
			return labels[i].ModelName
		}
	}	
	return ""
}

func getCode(indexes []int, tree Object, labels []Translate) (string) {
	var code string
	code = ""
	// Shall do that if there is a shape attached to the object not based on the front name of the object

	hasShape := 0
        for i := 0 ; i < len(labels) ; i++ {
                if ( labels[i].FreeCADName == tree.Objects[indexes[0]].Name ) {
                        hasShape = labels[i].hasShape
                }
        }


	if ( hasShape == 1 ) {
		return "<li id='" + tree.Objects[indexes[0]].Name+"'>\n" + label(tree.Objects[indexes[0]].Name, labels) + "</li>\n"
	} else {
		code = code + "\n<li>\n" + tree.Objects[indexes[0]].Name + "<ul>\n" 
		for i := 0 ; i < len(tree.Objects[indexes[0]].Properties[indexes[1]].LinkList) ; i++  {
			newindexes := []int { 0,0 }
			newindexes = getIndex(tree.Objects[indexes[0]].Properties[indexes[1]].LinkList[i].Entry, tree)
			code = code + getCode(newindexes, tree, labels) 
		}
		code = code + "\n</ul></li>\n"
	}
	return code
}

func getPlayerCode(w http.ResponseWriter, path string, private int) {

        contents := getJSONEntry(path, private)

        // I must put the content into the right structure
        var dataFreeCAD freecadEntry
        _ = json.Unmarshal([]byte(contents), &dataFreeCAD)

        // We must create the XeoglCode

        client := &http.Client{}

        fullPath := "/html/playerxeogl.html"

        method := "GET"

        req, _ := http.NewRequest(method,"http://"+StorageURI+StorageTCPPORT+fullPath, nil)


        // That is a new request so let's do it
        response, _ := client.Do(req)
        defer response.Body.Close()
        content, _ := ioutil.ReadAll(response.Body)

	// We must know how many parts stands into the part ... based on that we can generate the code properly

        fullPath = "/ctrl/"+dataFreeCAD.Bucket + "r" + dataFreeCAD.Revision + ".json"

        method = "GET"
        realPort,_ := strconv.Atoi(dataFreeCAD.Port)
        realPort = realPort + 1000 + base.MinIOServerBasePort
	var err error
	response, err = base.Request(method, "http://"+dataFreeCAD.URI+":"+strconv.Itoa(realPort)+fullPath, fullPath, "application/octet-stream", nil, "", data.Key, data.SecretToken)


        defer response.Body.Close()
	partNbInt := -1
	if ( err == nil ) {
	        partnb, _ := ioutil.ReadAll(response.Body)
		partNbInt,err = strconv.Atoi(string(partnb))
		if ( err != nil ) {
			partNbInt = -1
		}
	} 

	Private := ""
        keyWords := strings.Split(path, "/")
	if ( private == 0 ) {
		Private="0"
		if ( partNbInt == - 1 ) {
        		fullPath = "/projects/getModel/"+keyWords[3] + "/" + keyWords[4] + "/" + keyWords[5] + "r" + keyWords[6]  + "/" + Private + "/" 
		} else {
	        	fullPath = "/projects/getModel/"+keyWords[3] + "/" + keyWords[4] + "/" + keyWords[5] + "r" + keyWords[6] + "/" + Private + "/" + "p0"
		}
	} else {
		Private="1"
		if ( partNbInt == - 1 ) {
	        	fullPath = "/projects/"+ keyWords[4] +"/getModel/"+keyWords[3] + "/" + keyWords[4] + "/" + keyWords[5] + "r" + keyWords[6] + "/" + Private + "/"
		} else {
	        	fullPath = "/projects/"+ keyWords[4] +"/getModel/"+keyWords[3] + "/" + keyWords[4] + "/" + keyWords[5] + "r" + keyWords[6] + "/" + Private + "/" + "p0"
		}
	}


	contentString := strings.Replace(string(content), "MODELSOURCE", fullPath, 1)

	// We must use the NEWCODE stuff to request additionnal object parts
	code := ""

	if ( partNbInt != -1 ) {
		for i := 1 ; i < (partNbInt +1) ; i++ {
			code = code+"Model.addChild(new xeogl.OBJModel({" 
	               	code = code+"id: \"Part"+strconv.Itoa(i)+"\"," 
			if ( private == 0 ) {
				code = code+"src: \""+"/projects/getModel/"+keyWords[3] + "/" + keyWords[4] + "/" + keyWords[5] + "r" + keyWords[6] + "/" + Private + "/" + "p"+strconv.Itoa(i)+"\""
			} else {
				code = code+"src: \""+"/projects/"+ keyWords[4] +"/getModel/"+keyWords[3] + "/" + keyWords[4] + "/" + keyWords[5] + "r" + keyWords[6] + "/" + Private + "/" + "p"+strconv.Itoa(i)+"\""
			}
	                code = code+" }).on(\"loaded\", function () {"
	                code = code+"        var cameraFlight = new xeogl.CameraFlightAnimation();"
	                code = code+"        cameraFlight.flyTo(Model.scene);"
	                code = code+"        console.log(\"Camera adjusted\");"
			code = code+" }));"
		}
	}
	code = code + "var partNb="+ strconv.Itoa(partNbInt) + ";\n"
	contentString = strings.Replace(contentString, "NEWCODE", code, 1)

	// We must request the Part tree

        fullPath = "/ctrl/"+dataFreeCAD.Bucket + "r" + dataFreeCAD.Revision + ".tree"

        method = "GET"
        realPort,_ = strconv.Atoi(dataFreeCAD.Port)
        realPort = realPort + 1000 + base.MinIOServerBasePort

        response, err = base.Request(method, "http://"+dataFreeCAD.URI+":"+strconv.Itoa(realPort)+fullPath, fullPath, "application/octet-stream", nil, "", data.Key, data.SecretToken)


        defer response.Body.Close()
        nodes := -1
        if ( err == nil ) {
                tree, _ := ioutil.ReadAll(response.Body)
	        reader := bufio.NewReader(bytes.NewReader([]byte(tree)))
       		line, err := reader.ReadString('\n')
                nodes,err = strconv.Atoi(string(line))
                if ( err != nil ) {
                        nodes = -1
                }
        }


	if ( nodes == -1 ) {
		// We must get the xml file

	        fullPath = "/"+dataFreeCAD.Bucket + "r" + dataFreeCAD.Revision + "/Document.xml"

	        method = "GET"
	        realPort,_ = strconv.Atoi(dataFreeCAD.Port)
	        realPort = realPort + 1000 + base.MinIOServerBasePort

	        response, err = base.Request(method, "http://"+dataFreeCAD.URI+":"+strconv.Itoa(realPort)+fullPath, fullPath, "application/octet-stream", nil, "",dataFreeCAD.Key, dataFreeCAD.SecretKey)


	        defer response.Body.Close()
	        if ( err == nil ) {
	                Document, _ := ioutil.ReadAll(response.Body)
		contents := Object{}
		parts := []Translate{}

		// We have to build the tree from the node which are not part of any other nodes
		// To do that we have 2 kinds of nodes. The One which are Parts, and the one which are Assemblies
		// We must list them tne loop over them and tag them by being part of an Assembly or not
		// The one which are not part are the one that we must dump as there leaf will come with them
		
                in := bytes.NewReader(Document)
                _ = xml.NewDecoder(in).Decode(&contents)

		objectIndex := []int { 0, 0 }

		// We must keep only the stuff which have a property name label
		for i:= 0 ; i < len(contents.Objects) ; i++ {
				hasLinks := 0
				hasShape := 0
				var entry Translate
				for j:=0 ; j < len(contents.Objects[i].Properties) ; j++  {
					if ( contents.Objects[i].Properties[j].Name == "Label" ) {
						entry.FreeCADName = contents.Objects[i].Name
						entry.ModelName = contents.Objects[i].Properties[j].Strings[0].Value
						entry.Tag = 0
						entry.hasShape = 0
					}
					if ( contents.Objects[i].Properties[j].Name == "Group" ) {
						hasLinks = 1
						objectIndex[0] = i
                                                objectIndex[1] = j
					}
					if ( contents.Objects[i].Properties[j].Name == "Shape" ) {
						hasShape = 1
						entry.hasShape = 1
					}
				}
				if (  hasLinks == 0  && hasShape == 1) {
					parts = append(parts,entry)
				}
		}

		for i:=0 ; i < len(parts) ; i ++ {
			for j:=0 ; j < len(contents.Objects) ; j++ {
				for k := 0; k < len(contents.Objects[j].Properties) ; k++ {
					if ( contents.Objects[j].Properties[k].Name == "Group" ) {
						for l := 0 ; l < len(contents.Objects[j].Properties[k].LinkList) ; l++ {
							if ( parts[i].FreeCADName == contents.Objects[j].Properties[k].LinkList[l].Entry ) {
								parts[i].Tag = 1
							}
						}
					}
				}
			}	
		}

		code := ""
		if ( len(parts) == 0 ) {
			code = "<ul id='tree1' style='display:none;'><li>" + getCode(objectIndex, contents, parts) + "</li></ul>"
		} else {
			code = "<ul id='tree1' style='display:none;'><li><li>Root</li><ul>"
			switched := 0
			for i:=0 ; i < len(parts) ; i ++ {
				if ( parts[i].Tag == 0 ) {
					objectIndex[0] = i;
					objectIndex[1] = 0;
					switched = 1
					code = code +  getCode(objectIndex, contents, parts) 
				}
			}
			if ( switched == 0 ) {
				code = code +  getCode(objectIndex, contents, parts)
			}
			code = code + "</ul></li></ul>"
		}
	        contentString = strings.Replace(contentString, "TREE", code, 1)

	        }
	}

	// We got the tree
	

        // I don't need to encode it this is soon made by the users API
        w.Write([]byte(contentString))

}


// getModel is used from xeogl OBJ loader, so if it is not connected it can't get access to private data


func getModel(w http.ResponseWriter, path string, private int) {


	// We must get the JSON content to know where to load the requested file from
	// The format is unstandard as it is issued from XEOGL javascript


        keyWords := strings.Split(path, "/")
        targetUrl := keyWords[3] + "/" + keyWords[4] + "/" + keyWords[5] + ".json"

        // We must load the json file to know where the data are stored before forwarding them
        // to the client

	fullPath := ""

	if ( private == 1 ) {
        	fullPath = "/private/" + targetUrl
	} else {
	        fullPath = "/public/" + targetUrl
	}

        method := "GET"

        response, _ := base.Request(method,"http://"+ProjectURI+ProjectMinIOPort+fullPath, fullPath, "application/xml", nil, "",data.Key, data.SecretToken)

        defer response.Body.Close()
        contents, _ := ioutil.ReadAll(response.Body)



        // I must put the content into the right structure
        var dataFreeCAD freecadEntry
        _ = json.Unmarshal([]byte(contents), &dataFreeCAD)

        fullPath = "/ctrl/"+dataFreeCAD.Bucket + "r" + dataFreeCAD.Revision + keyWords[7] + ".obj"

        method = "GET"
        realPort,_ := strconv.Atoi(dataFreeCAD.Port)
        realPort = realPort + 1000 + base.MinIOServerBasePort


	response, _ = base.Request(method,"http://"+dataFreeCAD.URI+":"+strconv.Itoa(realPort)+fullPath, fullPath, "application/octet-stream", nil,"",dataFreeCAD.Key, dataFreeCAD.SecretKey)


        defer response.Body.Close()
        content, _ := ioutil.ReadAll(response.Body)

        w.Write([]byte((content)))

}

func userCallback(w http.ResponseWriter, r *http.Request) {
	command := [...]string{ "getList", "getMagnet", "getAvatar", "getPlayerCode", "getModel" }

	words := strings.Split(r.URL.Path, "/")

	var found int

	found = 0
	if ( len(words) > 2 ) {
		for j := 0 ; j < len(command) ; j++ {
			if ( words[2] == command[j] ) {
				found = 1
			}
		}
	}

        switch r.Method {
                case http.MethodPut:
			// We create the entry only if this is a new FreeCAD file otherwise 
			// that is a move operation which is driven by an Array
			body := string(base.HTTPGetBody(r))
			if ( body[0] == '[' ) {
				moveEntry(r, body)
			} else {
				createEntry(body)
			}
		case http.MethodGet:
			context := strings.Split(r.URL.Path, "/")	
			// Check anonymous call
			// We must do that test only if the retreived data is private
			// If it is not private
			if ( checkAccess(w,r,context[2]) ) {
				keyWords := strings.Split(r.URL.Path, "/")
				keyWords = append(keyWords[:2], keyWords[3:]...)
				path := strings.Join(keyWords[:], "/")
				// The last entry from the command line is the Public / Private switch
				private := keyWords[len(keyWords) - 1 ]
				Private := 1
				if ( private == "0" ) {
					Private = 0
				}
				switch context[3] {
                                        case "getList":
                                                w.Write([]byte(getList(context[2])))
                                        case "getMagnet":
                                                getMagnet(w,path,Private)
                                        case "getAvatar":
                                                getAvatar(w,path,Private)
                                        case "getPlayerCode":
                                                getPlayerCode(w,path,Private)
                                        case "getModel":
                                                getModel(w,path,Private)
                                }
			} else {
				if ( found == 1 ) {
                                	switch context[2] {
	                                        case "getList":
	                                                w.Write([]byte(getList("")))
							return
	                                        case "getMagnet":
	                                                getMagnet(w,r.URL.Path,0)
							return
	                                        case "getAvatar":
	                                                getAvatar(w,r.URL.Path,0)
							return
	                                        case "getPlayerCode":
	                                                getPlayerCode(w,r.URL.Path,0)
							return
	                                        case "getModel":
	                                                getModel(w,r.URL.Path,0)
							return
	                                }
					w.Write(([]byte)("Error access denierd"))
				}
			
			}
                default:
        }
}





var StorageURI = os.Getenv("STORAGE_URI")
var StorageTCPPORT = os.Getenv("STORAGE_TCPPORT")
var minIOURI = os.Getenv("MINIO_URI")
var minIOTCPPORT = os.Getenv("MINIO_TCPPORT")
var projectURI = os.Getenv("PROJECT_URI")


func start_minio() {

	// Must ask to the storage backend if I soon have accessToken or not
	// otherwise I must create new ones
	// and start minio based on them

	content:=base.HTTPGetRequest("http://"+StorageURI+StorageTCPPORT+"/projects/")

	if ( len(content) == 0 ) {

		data.Key = base.GenerateAccountACKLink(20)
		data.SecretToken = base.GenerateAuthToken("mac",40)
		buffer,_ := json.Marshal(data)
		_=base.HTTPPutRequest("http://"+StorageURI+StorageTCPPORT+"/projects/", buffer, "application/json")
	} else {

		_=json.Unmarshal([]byte(content), &data)
	}

	// I must take care of the Key and Token
	os.Setenv("MINIO_ACCESS_KEY", data.Key) 
	os.Setenv("MINIO_SECRET_KEY", data.SecretToken)
	os.Setenv("MINIO_BROWSER", "off")

	s := os.Getenv("PROJECT_MINIO_TCPPORT")

	args := []string { "server", "--address" }
	args = append (args, projectURI+ s)

	path := os.Getenv("PROJECT_TEMP")

	args = append (args, path)
	cmd := exec.Command("minio", args...)
	cmd.Start()
	done := make(chan error, 1)
	go func() {
	done <- cmd.Wait()
	}()

	// We must check if the public and private bucket exist
	// if not then we must create them

        initialBuckets := [...]string {"public", "private"}

	for _,value := range initialBuckets {

	        fullPath := "/"+value+"/"

	        method := "GET"

	        response, err := base.Request(method, "http://"+ProjectURI+ProjectMinIOPort+fullPath, fullPath, "application/xml", nil, "", data.Key, data.SecretToken)

	        if err != nil {
			// Error might be caused by the fact that the daemon is not running yet
			for err != nil {
                                        time.Sleep(1*time.Second)
                                        response,err = base.Request(method, "http://"+ProjectURI+ProjectMinIOPort+fullPath, fullPath, "application/xml", nil, "", data.Key, data.SecretToken)
                        }
	        } 

		init := 0 
		for init != 1 {
	                defer response.Body.Close()
	                contents, _ := ioutil.ReadAll(response.Body)
	  	        // I must parse the output
	                type Code struct {
	                        XMLName   xml.Name `xml:"Error"`
	                        CodeName string `xml:"Code"`
	                }

	                XMLcontents := Code{}
	                in := bytes.NewReader([]byte(contents))
	                _ = xml.NewDecoder(in).Decode(&XMLcontents)
	                if ( XMLcontents.CodeName == "NoSuchBucket" ) {
	                        // We must create the bucket
	                        fullPath := "/"+value+"/"
	                        method := "PUT"
				_,_ = base.Request(method, "http://"+ProjectURI+ProjectMinIOPort+fullPath, fullPath, "application/xml", nil, "", data.Key, data.SecretToken)
				init = 1
	               	} 
			if ( XMLcontents.CodeName == "XMinioServerNotInitialized" ) {
				time.Sleep(1*time.Second)
				response,err = base.Request(method, "http://"+ProjectURI+ProjectMinIOPort+fullPath, fullPath, "application/xml", nil, "", data.Key, data.SecretToken)
			}
		}
	}



}


// This is our very basic metadata server (aka access auth servers)

func main() {
    print("=============================== \n")
    print("| Starting Project backend    |\n")
    print("| (c) 2020 CADCloud           |\n")
    print("| Development version -       |\n")
    print("| Private use only            |\n")
    print("=============================== \n")

    // The project backend data are stored within a minio instances
    // These datas are the metadatas of the end user and we shall be sure that
    // they are resilient to system crash etc
    // We also need to be sure that they scale per server

    start_minio()


    mux := http.NewServeMux()
    var ProjectURI = os.Getenv("PROJECT_URI")
    var ProjectTCPPORT = os.Getenv("PROJECT_TCPPORT")

    mux.HandleFunc("/", userCallback)

    log.Fatal(http.ListenAndServe(ProjectURI+ProjectTCPPORT, mux))
}

