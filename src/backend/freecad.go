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
	"bufio"
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

var freeCADTemplate = os.Getenv("FREECAD_TEMPLATE")
var freeCADBinary = os.Getenv("FREECAD_BINARY")
var freeCADTMP = os.Getenv("FREECAD_TEMP")
var pipeName = "_Worker_pipe_"
var outputPipe *os.File

var file sync.RWMutex

type freecadEntry struct {
	Nickname      string
	Key           string
	SecretKey     string
	URI           string
	Port          string
	DNSDomain     string
	MasterTCPPort string
	Bucket        string
	Revision      string
}

func createEntry(content string) int {
	var data freecadEntry
	json.Unmarshal([]byte(content), &data)

	// Let's read the template file
	templateLocation := os.Getenv("FREECAD_TEMPLATE")
	file, err := os.Open(templateLocation + "templatePreview.py")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	b, err := ioutil.ReadAll(file)
	var fileContent string
	var port int
	if data.MasterTCPPort != "" {
		port, _ = strconv.Atoi(data.MasterTCPPort)
	} else {
		port, _ = strconv.Atoi(data.Port)
		port = port + 1000 + base.MinIOServerBasePort
	}

	fileContent = string(b)
	fileContent = strings.Replace(fileContent, "NAME", data.Bucket+"r"+data.Revision, -1)
	fileContent = strings.Replace(fileContent, "SECRETKEY", data.SecretKey, -1)
	fileContent = strings.Replace(fileContent, "KEY", data.Key, -1)
	if data.DNSDomain != "" {
		fileContent = strings.Replace(fileContent, "URI", data.DNSDomain, -1)
		fileContent = strings.Replace(fileContent, "BUCKET", data.Bucket, -1)
	} else {
		fileContent = strings.Replace(fileContent, "URI", data.URI, -1)
		fileContent = strings.Replace(fileContent, "BUCKET", data.Bucket+"r"+data.Revision, -1)
	}
	fileContent = strings.Replace(fileContent, "PORT", strconv.Itoa(port), -1)
	fileContent = strings.Replace(fileContent, "FILE_PATH", freeCADTMP+data.Key+data.Bucket+"r"+data.Revision+".png", -1)
	fileContent = strings.Replace(fileContent, "OBJ_PATH", freeCADTMP+data.Key+data.Bucket+"r"+data.Revision+".obj", -1)

	// FileContent contains the script to be executed by FreeCAD it will generate a file preview and a Wavefront OBJ format
	// this file shall be put into the ctrlr0 bucket from the user account
	// We must pass it to FreeCAD but this is a scarce ressource and that shall be done properly
	// If we got a system crash or whatever

	// We are storing each FreeCAD script to the file system and use a named pipe to process the file
	// when the file has been processed then the python script is destroyed

	// The script must contains all the required info to allow a recovery This include the data contained in content parameter which is the first line of the script
	// as a comment

	script := "# " + content + "\n" + fileContent

	// I must output the script
	f, _ := os.Create(freeCADTMP + data.Key + data.Bucket + "r" + data.Revision + ".py")
	_, _ = f.Write([]byte(script))
	f.Sync()
	// We can inform our worker thread that the script is there
	f.Close()
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	outputPipe.WriteString(fmt.Sprintf(freeCADTMP + data.Key + data.Bucket + "r" + data.Revision + ".py" + "\n"))
	return 1
}

func userCallback(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPut:
		createEntry(string(base.HTTPGetBody(r)))
	default:
	}
}

func objShift(line string, vBase int, vnBase int) string {
	var val1, val2, val3, val4, val5, val6 int
	fmt.Sscanf(line, "f %d//%d %d//%d %d//%d\n", &val1, &val2, &val3, &val4, &val5, &val6)
	return fmt.Sprintf("f %d//%d %d//%d %d//%d\n", val1-vBase, val2-vnBase, val3-vBase, val4-vnBase, val5-vBase, val6-vnBase)
}

func objShiftBuf(tmpBuffer []byte, vOffset int, vnOffset int) []byte {
	reader := bufio.NewReader(bytes.NewReader(tmpBuffer))
	line, err := reader.ReadString('\n')
	buffer := []byte{}
	for err == nil {
		if strings.HasPrefix(line, "f ") {
			line = objShift(line, vOffset, vnOffset)
		}
		buffer = append(buffer, []byte(line)...)
		line, err = reader.ReadString('\n')
	}
	return buffer
}

func worker() {
	file, err := os.OpenFile(freeCADTMP+pipeName, os.O_RDONLY, 0600)
	if err != nil {
		log.Fatal("Open named pipe file error:", err)
	}
	defer file.Close()
	reader := bufio.NewReader(file)
	for {
		// We are reading line per line
		// the pipe contains the name of the python file to process
		line, err := reader.ReadBytes('\n')
		if err == nil {
			// Second parameter shall be a string array
			args := []string{"-P", freeCADTMP}
			inputPath := string(line)
			args = append(args, inputPath[:len(inputPath)-1])
			cmd := exec.Command(freeCADBinary, args...)
			cmd.Start()
			cmd.Wait()

			// Normally everything went smoothly
			// so we can push back the data to the ctrl part of the user who updated his bucket
			// The first line of the file contains the relevant data to push back the various output data to the
			// ctrl bucket

			// We must determine if the data is public or private
			// To do this we are issuing a pseudo getList and look for previous revisions

			var dataInFile freecadEntry

			file, err := os.Open(inputPath[:len(inputPath)-1])
			if err != nil {
				log.Fatal(err)
			}
			defer file.Close()

			reader := bufio.NewReader(file)

			firstLine, err := reader.ReadString('\n')
			firstLine = strings.TrimPrefix(firstLine, "# ")

			json.Unmarshal([]byte(firstLine), &dataInFile)

			// We have to postprocess the OBJ file as to avoid that the transfer becomes to fat

			maxSize := 16 * 1024 * 1024

			fileOBJ, _ := os.Open(freeCADTMP + dataInFile.Key + dataInFile.Bucket + "r" + dataInFile.Revision + ".obj")
			defer fileOBJ.Close()

			reader = bufio.NewReader(fileOBJ)

			line, err := reader.ReadString('\n')

			buffer := []byte{}
			tmpBuffer := []byte{}
			var mtllib string
			partnb := 0

			vBase := 0
			vnBase := 0
			fBase := 0

			vOffset := 0
			vnOffset := 0

			vOBJOffset := 0
			vnOBJOffset := 0

			for err == nil {
				if len(mtllib) == 0 {
					if strings.HasPrefix(line, "mtllib") {
						mtllib = line
					}
				}
				if strings.HasPrefix(line, "v ") {
					// That is a vector
					vOBJOffset = vOBJOffset + 1
				}
				if strings.HasPrefix(line, "vn ") {
					// That is a vector normal
					vnOBJOffset = vnOBJOffset + 1
				}
				if strings.HasPrefix(line, "f ") {
					// That is a face which is defined by an three coordinates points v//vn both globally indexed
					fBase = fBase + 1
					// We must shift the values
					line = objShift(line, vBase, vnBase)
				}

				// We can detect an new entry
				if strings.HasPrefix(line, "o ") {
					// We start recording the new object
					if len(tmpBuffer) != 0 {
						// We must store the temporary buffer into the main buffer if the size is ok
						if (len(buffer) + len(tmpBuffer)) < maxSize {
							buffer = append(buffer, tmpBuffer[:]...)
							tmpBuffer = []byte{}
							tmpBuffer = append(tmpBuffer, []byte(line)...)
							vOffset = vOBJOffset
							vnOffset = vnOBJOffset
						} else {
							// we need to purge the buffer and create a new file entry
							partnbString := strconv.Itoa(partnb)
							fullName := "/ctrl/" + dataInFile.Bucket + "r" + dataInFile.Revision + "p" + partnbString + ".obj"
							realport, _ := strconv.Atoi(dataInFile.Port)
							realport = realport + 1000 + base.MinIOServerBasePort
							fullPath := fullName
							method := "PUT"

							_, _ = base.Request(method, "http://"+dataInFile.URI+":"+strconv.Itoa(realport)+fullPath, fullPath, "application/octet-stream", buffer, "", dataInFile.Key, dataInFile.SecretKey)

							// The Buffer has been dumped, we must update the Base offsets
							previousvBase := vBase
							previousvnBase := vnBase
							vBase = vOffset
							vnBase = vnOffset

							// We must realign the tmpBuffer with the vOffset and vnOffset on there face model
							// as we dumped the buffer and they are going to become the new entry into a new obj file
							// tmpBuffer = objShiftBuf(tmpBuffer, vOffset , vnOffset )
							tmpBuffer = objShiftBuf(tmpBuffer, vOffset-previousvBase, vnOffset-previousvnBase)

							vOffset = vOBJOffset
							vnOffset = vnOBJOffset

							partnb = partnb + 1
							buffer = []byte{}
							buffer = append([]byte(mtllib), tmpBuffer[:]...)
							tmpBuffer = []byte{}
							tmpBuffer = append(tmpBuffer, []byte(line)...)
						}
					} else {
						// We have to init the buffer with the Material library
						tmpBuffer = []byte(mtllib)
						tmpBuffer = append(tmpBuffer, []byte(line)...)
					}
				} else {
					tmpBuffer = append(tmpBuffer, []byte(line)...)
				}
				line, err = reader.ReadString('\n')
			}

			// We must dump the tail

			if (len(buffer) + len(tmpBuffer)) < maxSize {
				buffer = append(buffer, tmpBuffer[:]...)
			} else {
				partnbString := strconv.Itoa(partnb)
				fullName := "/ctrl/" + dataInFile.Bucket + "r" + dataInFile.Revision + "p" + partnbString + ".obj"
				realport, _ := strconv.Atoi(dataInFile.Port)
				realport = realport + 1000 + base.MinIOServerBasePort
				fullPath := fullName
				method := "PUT"

				_, _ = base.Request(method, "http://"+dataInFile.URI+":"+strconv.Itoa(realport)+fullPath, fullPath, "application/octet-stream", buffer, "", dataInFile.Key, dataInFile.SecretKey)

				buffer = append([]byte(mtllib), tmpBuffer[:]...)
				partnb = partnb + 1
			}
			if len(buffer) > 0 {
				partnbString := strconv.Itoa(partnb)
				fullName := "/ctrl/" + dataInFile.Bucket + "r" + dataInFile.Revision + "p" + partnbString + ".obj"
				realport, _ := strconv.Atoi(dataInFile.Port)
				realport = realport + 1000 + base.MinIOServerBasePort
				fullPath := fullName
				method := "PUT"

				_, _ = base.Request(method, "http://"+dataInFile.URI+":"+strconv.Itoa(realport)+fullPath, fullPath, "application/octet-stream", buffer, "", dataInFile.Key, dataInFile.SecretKey)

			}

			// We must now save how many parts are present within the split OBJ

			fullName := "/ctrl/" + dataInFile.Bucket + "r" + dataInFile.Revision + ".json"
			realport, _ := strconv.Atoi(dataInFile.Port)
			realport = realport + 1000 + base.MinIOServerBasePort
			fullPath := fullName
			method := "PUT"

			_, _ = base.Request(method, "http://"+dataInFile.URI+":"+strconv.Itoa(realport)+fullPath, fullPath, "application/octet-stream", []byte(strconv.Itoa(partnb)), "", dataInFile.Key, dataInFile.SecretKey)

			fileList := [3]string{freeCADTMP + dataInFile.Key + dataInFile.Bucket + "r" + dataInFile.Revision + ".png",
				freeCADTMP + dataInFile.Key + dataInFile.Bucket + "r" + dataInFile.Revision + ".obj",
				freeCADTMP + dataInFile.Key + dataInFile.Bucket + "r" + dataInFile.Revision + ".mtl"}

			fileTarget := [3]string{"/ctrl/" + dataInFile.Bucket + "r" + dataInFile.Revision + ".png",
				"/ctrl/" + dataInFile.Bucket + "r" + dataInFile.Revision + ".obj",
				"/ctrl/" + dataInFile.Bucket + "r" + dataInFile.Revision + ".mtl"}

			for index := range fileList {

				client := &http.Client{}

				realport, _ := strconv.Atoi(dataInFile.Port)
				realport = realport + 1000 + base.MinIOServerBasePort
				myDate := time.Now().UTC().Format(http.TimeFormat)
				myDate = strings.Replace(myDate, "GMT", "+0000", -1)

				fullPath := fileTarget[index]
				fileToSend, _ := os.Open(fileList[index])

				info, _ := fileToSend.Stat()

				defer fileToSend.Close()
				method := "PUT"

				req, _ := http.NewRequest(method, "http://"+dataInFile.URI+":"+strconv.Itoa(realport)+fullPath, bufio.NewReader(fileToSend))

				stringToSign := method + "\n\n" + "application/octet-stream" + "\n" + myDate + "\n" + fullPath

				mac := hmac.New(sha1.New, []byte(dataInFile.SecretKey))
				mac.Write([]byte(stringToSign))
				expectedMAC := mac.Sum(nil)
				signature := base64.StdEncoding.EncodeToString(expectedMAC)

				req.Header.Set("Authorization", "AWS "+dataInFile.Key+":"+signature)
				req.Header.Set("Date", myDate)
				req.Header.Set("Content-Type", "application/octet-stream")
				req.ContentLength = info.Size()

				// That is a new request so let's do it
				client.Do(req)

				os.Remove(fileList[index])
			}
			// We must push the data to the Project management backend
			// which is storing a list of accessible project and credential to do so

			projectURI := os.Getenv("PROJECT_URI")
			projectTCPPORT := os.Getenv("PROJECT_TCPPORT")
			content, _ := json.Marshal(dataInFile)
			base.HTTPPutRequest("http://"+projectURI+projectTCPPORT+"/", content, "application/json")

			os.Remove(freeCADTMP + dataInFile.Key + dataInFile.Bucket + "r" + dataInFile.Revision + ".py")
			os.Remove(freeCADTMP + dataInFile.Key + dataInFile.Bucket + "r" + dataInFile.Revision + ".pyc")
		}
	}

}

func main() {
	print("=============================== \n")
	print("| Starting freecad backend    |\n")
	print("| (c) 2020 CADCloud           |\n")
	print("| Development version -       |\n")
	print("| Private use only            |\n")
	print("=============================== \n")
	os.Remove(freeCADTMP + pipeName)
	err := syscall.Mkfifo(freeCADTMP+pipeName, 0600)
	if err != nil {
		log.Fatal("Make named pipe file error:", err)
	}

	go worker()

	// we can open the writer

	outputPipe, _ = os.OpenFile(freeCADTMP+pipeName, os.O_WRONLY, 0600)

	defer outputPipe.Close()

	mux := http.NewServeMux()
	var freeCADURI = os.Getenv("FREECAD_URI")
	var freeCADTCPPORT = os.Getenv("FREECAD_TCPPORT")

	mux.HandleFunc("/", userCallback)

	log.Fatal(http.ListenAndServe(freeCADURI+freeCADTCPPORT, mux))
}
