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

package base
import (
    "net/http"
    "log"
    "io/ioutil"
    "fmt"
    "math/rand"
    "golang.org/x/crypto/bcrypt"
    "time"
    "bytes"
    "net/smtp"
    "net/mail"
    "net"
    "crypto/tls"
    "crypto/hmac"
    "encoding/base64"
    "crypto/sha1"
    "strings"
    "os"
)

type MinIOServer struct {
        URI string
        Port string
}

const MaxMinIOServer = 100
const MinIOServerBasePort = 9400

type User struct {
        Nickname string
        Password string
        TokenType string
        TokenAuth string
        TokenSecret string
        CreationDate string
        Lastlogin string
        Email string
        Active int
        ValidationString string
        Ports string
        Server string
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789+/")
var simpleLetters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
var randInit = 0

func randAlphaSlashPlus(n int) string {
    if ( randInit == 0 ) {
	rand.Seed(time.Now().UnixNano())
    }
    b := make([]rune, n)
    for i := range b {
        b[i] = letters[rand.Intn(len(letters))]
    }
    return string(b)
}

func randAlpha(n int) string {
    if ( randInit == 0 ) {
        rand.Seed(time.Now().UnixNano())
    }
    b := make([]rune, n)
    for i := range b {
        b[i] = simpleLetters[rand.Intn(len(simpleLetters))]
    }
    return string(b)
}

func GenerateAccountACKLink(length int) (string) {
	return randAlpha(length)
}

func GenerateAuthToken(TokenType string, length int) (string) {
	return randAlphaSlashPlus(length)
}

func HashPassword(password string) (string, error) {
    bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
    return err == nil
}

// Send some email

var smtpServer = os.Getenv("SMTP_SERVER") // example: smtp.google.com:587
var smtpAccount =  os.Getenv("SMTP_ACCOUNT") 
var smtpPassword = os.Getenv("SMTP_PASSWORD")

func SendEmail(email string, subject string, validationString string) {
    servername := smtpServer
    host, _, _ := net.SplitHostPort(servername)
    from := mail.Address{"", smtpAccount+"@"+host}
    to   := mail.Address{"", email}
    subj := subject
    body := validationString

    // Setup headers
    headers := make(map[string]string)
    headers["From"] = from.String()
    headers["To"] = to.String()
    headers["Subject"] = subj

    // Setup message
    message := ""
    for k,v := range headers {
        message += fmt.Sprintf("%s: %s\r\n", k, v)
    }
    message += "\r\n" + body

    // Connect to the SMTP Server

    auth := smtp.PlainAuth("",smtpAccount, smtpPassword, host)

    // TLS config
    tlsconfig := &tls.Config {
        InsecureSkipVerify: true,
        ServerName: host,
    }


    // uncomment the following line to use a pure SSL connection without STARTTLS

    //conn, err := tls.Dial("tcp", servername, tlsconfig)
    conn, err := smtp.Dial( servername)
    if err != nil {
        log.Panic(err)
    }

    // comment that line to use SSL connection

    conn.StartTLS(tlsconfig)

    // Auth
    if err = conn.Auth(auth); err != nil {
        log.Panic(err)
    }

    // To && From
    if err = conn.Mail(from.Address); err != nil {
        log.Panic(err)
    }

    if err = conn.Rcpt(to.Address); err != nil {
        log.Panic(err)
    }

    // Data
    w, err := conn.Data()
    if err != nil {
        log.Panic(err)
    }

    _, err = w.Write([]byte(message))
    if err != nil {
        log.Panic(err)
    }

    err = w.Close()
    if err != nil {
        log.Panic(err)
    }

    conn.Quit()
	if err != nil {
		log.Printf("smtp error: %s", err)
	}
	
}


func Request(method string, Uri string, Path string, Data string, content []byte, query string, Key string, SecretKey string) ( *http.Response, error) {

	client := &http.Client{}

        myDate := time.Now().UTC().Format(http.TimeFormat)
        myDate = strings.Replace(myDate, "GMT", "+0000", -1)
	var req *http.Request
	if ( content != nil ) {
        	req, _ = http.NewRequest(method,Uri, bytes.NewReader(content))
	} else {
		req, _ = http.NewRequest(method,Uri, nil)
	}

       stringToSign := method + "\n\n"+Data+"\n"+myDate+"\n"+Path

       mac := hmac.New(sha1.New, []byte(SecretKey))
       mac.Write([]byte(stringToSign))
       expectedMAC := mac.Sum(nil)
       signature:=base64.StdEncoding.EncodeToString(expectedMAC)

       req.Header.Set("Authorization","AWS "+Key+":"+signature)
       req.Header.Set("Date", myDate)
       req.Header.Set("Content-Type", Data)
       if ( len(content) > 0 ) {
       		req.ContentLength = int64(len(content))
       }

       req.URL.RawQuery = query


        // That is a new request so let's do it
        var response *http.Response
	var err error
        response, err  = client.Do(req)
	return response, err

}


// Some HTTP request
// Get request to the storage backend

func HTTPGetRequest(request string) (string){
        resp, err := http.Get(request)
        if err != nil {
                log.Fatalln(err)
        }

        body, err := ioutil.ReadAll(resp.Body)
        if err != nil {
                log.Fatalln(err)
        }
	return(string(body))
}

func HTTPDeleteRequest(request string){
	client := &http.Client{}
	content := []byte{0}
	httprequest, err := http.NewRequest("DELETE", request, bytes.NewReader(content))
        httprequest.ContentLength = 0
        response, err := client.Do(httprequest)
        if err != nil {
                log.Fatal(err)
        } else {
                defer response.Body.Close()
                _, err := ioutil.ReadAll(response.Body)
                if err != nil {
                        log.Fatal(err)
                }
        }
}

// Put request to the storage backend

func HTTPPutRequest(request string, content []byte, contentType string) (string){
	print("Running a PUT Request \n")
	client := &http.Client{}
	httprequest, err := http.NewRequest("PUT", request, bytes.NewReader(content))
	httprequest.Header.Set("Content-Type", contentType)
	httprequest.ContentLength = int64(len(content))
	response, err := client.Do(httprequest)
	if err != nil {
		log.Fatal(err)
	} else {
		defer response.Body.Close()
		contents, err := ioutil.ReadAll(response.Body)
		if err != nil {
			log.Fatal(err)
		}
		return string(contents)
	}
	return ""
}


func HTTPGetBody(r *http.Request) ([]byte) {
	buf, _ := ioutil.ReadAll(r.Body)
	rdr1 := ioutil.NopCloser(bytes.NewBuffer(buf))
	rdr2 := ioutil.NopCloser(bytes.NewBuffer(buf))
	b := new(bytes.Buffer)
	b.ReadFrom(rdr1)
	r.Body = rdr2
	return(b.Bytes())
}
