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
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

var storageURI = os.Getenv("STORAGE_URI")
var storageTCPPORT = os.Getenv("STORAGE_TCPPORT")
var minIOURI = os.Getenv("MINIO_URI")
var minIOTCPPORT = os.Getenv("MINIO_TCPPORT")

// Upercase is mandatory for JSON library parsing

type userPublic struct {
	Nickname         string
	NicknameRW       string
	NicknameLABEL    string
	TokenType        string
	TokenTypeRW      string
	TokenAuth        string
	TokenAuthRW      string
	TokenSecret      string
	TokenSecretLABEL string
	TokenSecretRW    string
	CreationDate     string
	CreationDateRW   string
	Lastlogin        string
	LastloginRW      string
	Email            string
	EmailRW          string
	EmailLABEL       string
}

func userExist(username string) bool {
	// We must call the storage backend with the username
	var result string
	// that must be an http request instead of a vejmarie
	result = base.HTTPGetRequest("http://" + storageURI + storageTCPPORT + "/user/" + username)
	if result == "Error" {
		return false
	}
	return true
}

func userGetInfo(nickname string) *userPublic {
	// We must call the storage backend service to get access to the resource
	// We could have a bucket / fileid approach which could be translated into flat file
	// or database management
	var tempvalue *base.User
	var returnvalue *userPublic
	var result string
	if userExist(nickname) {
		result = base.HTTPGetRequest("http://" + storageURI + storageTCPPORT + "/user/" + nickname)
		tempvalue = new(base.User)
		json.Unmarshal([]byte(result), tempvalue)
		returnvalue = new(userPublic)
		returnvalue.Nickname = tempvalue.Nickname
		returnvalue.NicknameRW = "0"
		returnvalue.NicknameLABEL = "This is your unique identifier. It will appeared within your publications and used to refer you as author. It is visible to any other users."
		returnvalue.TokenType = tempvalue.TokenType
		returnvalue.TokenTypeRW = "0"
		returnvalue.TokenAuth = tempvalue.TokenAuth
		returnvalue.TokenAuthRW = "0"
		returnvalue.TokenSecret = tempvalue.TokenSecret
		returnvalue.TokenSecretLABEL = "TokenType, TokenAuth and TokenSecret are private values that you shouldn't share with anybody. They are automatically assigned to you as to provide you unique authentication capabilities to JustYour.Parts services. Use them to connect you to the service through FreeCAD or an Amazon s3 compliant client. Please refer to our end user documentation for further informations."
		returnvalue.TokenSecretRW = "0"
		returnvalue.CreationDate = tempvalue.CreationDate
		returnvalue.CreationDateRW = "0"
		returnvalue.Lastlogin = tempvalue.Lastlogin
		returnvalue.LastloginRW = "0"
		returnvalue.Email = tempvalue.Email
		returnvalue.EmailLABEL = "Your primary email address. It won't be shared with anybody except if you explicitely activate that feature into the privacy field. Warning your email address must be verified each time you change it. During that process your account is disabled and can't be recovered without contacting us."
		returnvalue.EmailRW = "1"
	}

	return returnvalue
}

func userGetInternalInfo(nickname string) *base.User {
	// We must call the storage backend service to get access to the resource
	// We could have a bucket / fileid approach which could be translated into flat file
	// or database management
	var returnvalue *base.User
	var result string
	if userExist(nickname) {
		result = base.HTTPGetRequest("http://" + storageURI + storageTCPPORT + "/user/" + nickname)
		returnvalue = new(base.User)
		json.Unmarshal([]byte(result), returnvalue)
	}
	return returnvalue
}

func updateAccount(username string, w http.ResponseWriter, r *http.Request) bool {
	var updatedData *base.User
	var serverReturn string
	serverReturn = ""
	type accountUpdate struct {
		Email           string
		CurrentPassword string
		NewPassword0    string
		NewPassword1    string
	}
	exist := userExist(username)
	if !exist {
		fmt.Fprint(w, "Error")
		return false
	}
	updatedData = userGetInternalInfo(username)
	var getJSON = base.HTTPGetBody(r)
	var newData accountUpdate

	// We have to unMarshal the body to update the data

	_ = json.Unmarshal(getJSON, &newData)

	// So now let's run some comparaison
	if updatedData.Active == 0 {
		http.Error(w, "401 User not activated Please check email", 401)
		return false
	}

	if newData.CurrentPassword != "undefined" {
		if !base.CheckPasswordHash(newData.CurrentPassword, updatedData.Password) {
			w.Write([]byte("error password"))
			return false
		}
		// we are good to update the password and log off the user
		// but only if the size is bigger than 0 !
		if newData.NewPassword0 != "undefined" {
			updatedData.Password, _ = base.HashPassword(newData.NewPassword0)
			b, _ := json.Marshal(updatedData)
			base.HTTPPutRequest("http://"+storageURI+storageTCPPORT+"/user/"+updatedData.Nickname, b, "application/json")
			serverReturn = serverReturn + "password"
		}
	}

	// If the email address are different
	if updatedData.Email != newData.Email {
		// We must put the account into an inactive mode as long as the new email has not been validated
		// We must renew the email check account
		updatedData.Email = newData.Email
		updatedData.Active = 0
		// we change the Validation string and send the email
		updatedData.ValidationString = base.GenerateAccountACKLink(24)
		b, _ := json.Marshal(updatedData)
		base.HTTPPutRequest("http://"+storageURI+storageTCPPORT+"/user/"+updatedData.Nickname, b, "application/json")
		base.SendEmail(updatedData.Email, "Account activation - Action required",
			"Please click the following link as to validate your account https://"+
				r.Host+"/user/"+updatedData.Nickname+"/validateUser/"+updatedData.ValidationString)
		updatedData = nil
		serverReturn = serverReturn + "email"
	}

	// If the Password is modified we must validate that the previous password has been properly typed in
	w.Write([]byte(serverReturn))
	return true

}

func createUser(username string, w http.ResponseWriter, r *http.Request) bool {
	var updatedData *base.User
	exist := userExist(username)
	if exist {
		fmt.Fprint(w, "Error")
		return false
	}

	updatedData = new(base.User)
	updatedData.Nickname = username
	updatedData.Email = r.FormValue("email")

	// this is a creation
	updatedData.TokenAuth = base.GenerateAccountACKLink(20)
	updatedData.TokenSecret = base.GenerateAuthToken("mac", 40)
	updatedData.TokenType = "mac"
	updatedData.CreationDate = string(time.Now().Format(time.RFC1123Z))
	updatedData.Password, _ = base.HashPassword(r.FormValue("password"))
	updatedData.Lastlogin = ""
	updatedData.Active = 0
	updatedData.ValidationString = base.GenerateAccountACKLink(24)
	b, _ := json.Marshal(updatedData)
	base.HTTPPutRequest("http://"+storageURI+storageTCPPORT+"/user/"+updatedData.Nickname, b, "application/json")
	base.SendEmail(updatedData.Email, "Account activation - Action required",
		"Please click the following link as to validate your account https://"+
			r.Host+"/user/"+updatedData.Nickname+"/validateUser/"+updatedData.ValidationString)
	updatedData = nil
	return true

}

func updateAvatar(username string, w http.ResponseWriter, r *http.Request) bool {
	// We must store the body content within the avatar file of the end user
	exist := userExist(username)
	if !exist {
		fmt.Fprint(w, "Error")
		return false
	}
	base.HTTPPutRequest("http://"+storageURI+storageTCPPORT+"/user/"+username, base.HTTPGetBody(r), "image/jpg")
	return true
}

func getAvatar(username string, w *http.ResponseWriter) {
	exist := userExist(username)
	if !exist {
		fmt.Fprint(*w, "Error")
		return
	}
	(*w).Write([]byte(base.HTTPGetRequest("http://" + storageURI + storageTCPPORT + "/user/" + username + "/avatar")))
}

func sendPasswordResetLink(username string, w http.ResponseWriter, r *http.Request) bool {
	var updatedData *base.User
	exist := userExist(username)
	if !exist {
		fmt.Fprint(w, "Error")
		return false
	}
	updatedData = userGetInternalInfo(username)
	updatedData.ValidationString = base.GenerateAccountACKLink(24)
	// The user can't be active as long as we do not have reset the password
	updatedData.Active = 0
	b, _ := json.Marshal(updatedData)
	base.HTTPPutRequest("http://"+storageURI+storageTCPPORT+"/user/"+updatedData.Nickname, b, "application/json")
	base.SendEmail(updatedData.Email, "Account password reset - Action required",
		"Please click the following link as to update  your password https://"+
			r.Host+"/user/"+updatedData.Nickname+"/resetPassword/"+updatedData.ValidationString)
	updatedData = nil
	return true

}

func resetPassword(username string, w http.ResponseWriter, r *http.Request) bool {
	var updatedData *base.User
	exist := userExist(username)
	if !exist {
		fmt.Fprint(w, "Error")
		return false
	}
	updatedData = userGetInternalInfo(username)
	if updatedData.ValidationString != r.FormValue("validation") {
		fmt.Fprint(w, "Error")
		return false
	}
	updatedData.ValidationString = ""
	updatedData.Password, _ = base.HashPassword(r.FormValue("password"))
	updatedData.Active = 1
	b, _ := json.Marshal(updatedData)
	base.HTTPPutRequest("http://"+storageURI+storageTCPPORT+"/user/"+updatedData.Nickname, b, "application/json")
	return true
}

func validateUser(username string, validationstring string) bool {
	var updatedData *base.User
	// We  must check if the user exist
	exist := userExist(username)
	if !exist {
		return false
	}
	// We must read the user data and update the content of it
	updatedData = userGetInternalInfo(username)
	// We must check that the validation string is a match
	if updatedData.ValidationString != validationstring {
		return false
	}
	updatedData.Active = 1

	b, _ := json.Marshal(updatedData)
	// we must now initiate the minIO server for that user !
	data := base.HTTPPutRequest("http://"+minIOURI+minIOTCPPORT+"/user/"+updatedData.Nickname, b, "application/json")
	var entry *base.MinIOServer
	entry = new(base.MinIOServer)
	json.Unmarshal([]byte(data), entry)
	updatedData.Ports = entry.Port
	updatedData.Server = entry.URI

	// We write back the data
	c, _ := json.Marshal(updatedData)
	base.HTTPPutRequest("http://"+storageURI+storageTCPPORT+"/user/"+updatedData.Nickname, c, "application/json")

	// And return positively
	return true
}

func deleteUser(username string, w http.ResponseWriter, r *http.Request) bool {
	// We delete the user by a direct call to the storage subsystem
	var updatedData *base.User
	// I am receiving the password within the http body of the delete request
	type accountDelete struct {
		CurrentPassword string
		DeleteData      string
	}
	var newData accountDelete
	var getJSON = base.HTTPGetBody(r)
	_ = json.Unmarshal(getJSON, &newData)
	if newData.DeleteData == "true" {
	} else {
	}
	updatedData = userGetInternalInfo(username)
	// if the received password is not the one of the end user we can't erase it's account
	// might be a browser hack
	if !base.CheckPasswordHash(newData.CurrentPassword, updatedData.Password) {
		w.Write([]byte("error password"))
		return false
	}

	if newData.DeleteData == "true" {
		// We need to delete user content
		// We also need to stop the miniIO service
		c, _ := json.Marshal(updatedData)
		base.HTTPDeleteRequest("http://" + minIOURI + minIOTCPPORT + "/user/" + updatedData.Nickname)
		// Let's stop the server
		_ = base.HTTPPutRequest("http://"+minIOURI+minIOTCPPORT+"/user/"+updatedData.Nickname+"/stopServer", c, "application/json")
		base.HTTPDeleteRequest("http://" + storageURI + storageTCPPORT + "/user/" + username)
		// We must invalidate the user cache content
	} else {
		// Just need to disable the account by unactivating it
		// It could be recovered by resetting the password
		// I need to stop the miniIO service it is useless and shouldn't answer by the way
		updatedData.Active = 0
		c, _ := json.Marshal(updatedData)
		base.HTTPPutRequest("http://"+storageURI+storageTCPPORT+"/user/"+updatedData.Nickname, c, "application/json")
		// we must now stop the minIO server for that user !
		//	        _=base.HTTPPutRequest("http://"+minIOURI+minIOTCPPORT+"/user/"+updatedData.Nickname+"/stopServer",c,"application/json")
		// We must invalidate the user cache content
	}

	// And return positively
	return true
}

func userCallback(w http.ResponseWriter, r *http.Request) {
	var username, command string

	path := strings.Split(r.URL.Path, "/")
	if len(path) < 3 {
		http.Error(w, "401 Malformed URI", 401)
		return
	}
	username = path[2]
	if len(path) >= 4 {
		command = path[3]
	}
	switch r.Method {
	case http.MethodGet:
		switch command {
		case "validateUser":
			// got a validation link ....
			// we have to accept user activation
			// First check if the account exist
			// if yes we must get the data, compare the link and if a match
			// activate the user allowing a call to the API to get the connection token
			if !validateUser(username, path[4]) {
				http.Error(w, "401 Validation string error", 401)
			} else {
				http.Redirect(
					w, r,
					"https://"+r.Host+"/?loginValidated=1",
					http.StatusMovedPermanently,
				)
			}
		case "resetPassword":
			// We have to validate the user, then display the right return page
			if !validateUser(username, path[4]) {
				http.Error(w, "401 Validation string error", 401)
			} else {
				print("REDIRECTION")
				http.Redirect(
					w, r,
					"https://"+r.Host+"/?resetPassword=1&username="+username+"&validation="+path[4],
					http.StatusMovedPermanently,
				)
			}
		case "userGetInternalInfo":
			var result *base.User
			// Serve the resource.
			result = userGetInternalInfo(username)
			b, _ := json.Marshal(*result)
			fmt.Fprint(w, string(b))
		case "userGetInfo":
			var result *userPublic
			// Serve the resource.
			result = userGetInfo(username)
			b, _ := json.Marshal(*result)
			fmt.Fprint(w, string(b))

		case "getAvatar":
			getAvatar(username, &w)
		default:
			var result *base.User
			// Serve the resource.
			result = userGetInternalInfo(username)
			b, _ := json.Marshal(*result)
			fmt.Fprint(w, string(b))
		}
	case http.MethodPut:
		// Update an existing record.
		switch command {
		case "updateAvatar":
			updateAvatar(username, w, r)
		case "updateAccount":
			updateAccount(username, w, r)
		default:
		}
	case http.MethodPost:
		// Ok I am getting there the various parameters to log a user
		switch command {
		case "getToken":
			// We must get the user info and validate the password sent
			// if the user doesn't have any API Token
			// we have to generate it !
			// if the user doesn't exist we need to deny the request
			password := r.FormValue("password")
			var result *base.User
			result = userGetInternalInfo(username)
			if !base.CheckPasswordHash(password, result.Password) {
				http.Error(w, "401 Password error", 401)
				return
			}
			if result.Active == 0 {
				http.Error(w, "401 User not activated Please check email", 401)
				return
			}
			// We have the right password !
			// So, we need to send the secret and access token
			// as the end user could login the to the API
			// and load the right page !
			returnValue := " { \"accessKey\" : \"" + result.TokenAuth +
				"\", \"secretKey\" : \"" + result.TokenSecret + "\" }"
			result.Lastlogin = string(time.Now().Format(time.RFC1123Z))
			b, _ := json.Marshal(result)
			base.HTTPPutRequest("http://"+storageURI+storageTCPPORT+"/user/"+result.Nickname, b, "application/json")
			fmt.Fprintf(w, string(returnValue))
		case "createUser":
			createUser(username, w, r)
		case "generatePasswordLnkRst":
			sendPasswordResetLink(username, w, r)
		case "resetPassword":
			resetPassword(username, w, r)
		default:
			http.Error(w, "401 Unknown user command\n", 401)

		}
	case http.MethodDelete:
		// Remove the record.
		deleteUser(username, w, r)
	default:
	}
}

func main() {
	// http to https redirection
	print("=============================== \n")
	print("| Starting user credentials  |\n")
	print("| (c) 2019 CADCloud           |\n")
	print("| Development version -       |\n")
	print("| Private use only            |\n")
	print("=============================== \n")

	var wg sync.WaitGroup
	mux := http.NewServeMux()
	var CredentialURI = os.Getenv("CREDENTIALS_TCPPORT")
	print("Attaching to " + CredentialURI + "\n")
	// Serve one page site dynamic pages
	mux.HandleFunc("/user/", userCallback)
	wg.Add(1)
	go http.ListenAndServe(CredentialURI, mux)
	// we can release the minIOServer API
	var minIOServerURI = os.Getenv("MINIO_URI")
	var miniIOTCPPORT = os.Getenv("MINIO_TCPPORT")
	base.HTTPPutRequest("http://"+minIOServerURI+miniIOTCPPORT+"/start/", []byte{0}, "")
	wg.Wait()
}
