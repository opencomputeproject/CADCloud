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

var maxPasswordLength = 24;

function myAccountCol0()
{
	var col0='<div class="dropdown">\
          <div class="dropdown-item" >Your Account</div>\
          <div class="dropdown-item">Tell us about you</div>\
          <div class="dropdown-item">Privacy settings</div>\
        </div>';
        $('#col0').html(col0);
}

function checkPassword(id0, id1)
{
	if ( typeof mylocalStorage[id1] === undefined ) {
		$('#passwordMatch').html("Password don't match");
	}
	else
	{
		if ( mylocalStorage[id0] == mylocalStorage[id1] ) 
		{
			// Hey the password match we shall inform the end user
			$('#label'+id0).css('color','green');
			$('#label'+id1).css('color','green');
		}
		else
		{
			$('#label'+id0).css('color','red');
			$('#label'+id1).css('color','red');
		}
	}
}

function managePasswordEntry(id, storage, focus, onChange, idStorage0)
{
        $(id).keydown(function(e) {
                // e.preventDefault();
                var focus_id = e.target.id;
                switch (e.keyCode) {
                        case 37: // Go Left
                                mylocalStorage['caretPosition'] = getCaretPos($(id).get(0));
                                break;
                case 39: // Go right
                        mylocalStorage['caretPosition'] = getCaretPos($(id).get(0));
                        break;
                case 46: // Delete
                case 8: // Backspace
				// If we have a selection the way we compute the caret is different and we must adapt
				// the password string
				var selection = window.getSelection().toString();
				if ( selection !== undefined && selection.length != 0 ) {
					if ( mylocalStorage['DownMouseX'] > mylocalStorage['UpMouseX'] ) {
						// the caret value is the last character which has been selected on the right
						if ( mylocalStorage['caretPosition']-selection.length != 0 ) {
							// We must get the first part of the string
							mylocalStorage[storage] = mylocalStorage[storage].substring(0,mylocalStorage['caretPosition']-selection.length) +
										  mylocalStorage[storage].substring(mylocalStorage['caretPosition'],
										  mylocalStorage[storage].length);
						} else {
							mylocalStorage[storage] = mylocalStorage[storage].substring(mylocalStorage['caretPosition'],
                                                                                  mylocalStorage[storage].length);
						}
					} else {
							// the caret value is the last character which has been selected on the right
							if ( mylocalStorage['caretPosition']-selection.length != 0 ) {
	                                                        // We must get the first part of the string
	                                                        mylocalStorage[storage] = mylocalStorage[storage].substring(0,mylocalStorage['caretPosition']-selection.length) +
	                                                                                  mylocalStorage[storage].substring(mylocalStorage['caretPosition'],
	                                                                                  mylocalStorage[storage].length);
	                                                } else {
	                                                        mylocalStorage[storage] = mylocalStorage[storage].substring(mylocalStorage['caretPosition'],
	                                                                                  mylocalStorage[storage].length);
							}
						
					}
				}
                                $('#MyPassword').focus();
				
                                mylocalStorage['caretPosition'] = getCaretPos($(id).get(0));
                                if ( mylocalStorage['caretPosition'] != 0 ) {
                                        if ( mylocalStorage['caretPosition'] == mylocalStorage[storage].length ) {
                                                mylocalStorage[storage] = mylocalStorage[storage].substring(0, mylocalStorage['caretPosition']-1);
                                        } else {
                                                mylocalStorage[storage] = mylocalStorage[storage].substring(0,mylocalStorage['caretPosition']-1)+ 
									  mylocalStorage[storage].substring(mylocalStorage['caretPosition']);
                                        }
                                }
				if ( (maxPasswordLength-mylocalStorage[storage].length) < 5 ) {
                                        $(id+'Help').css('color', '#FF0000');
                                 }
                                 else
                                 {
                                        $(id+'Help').css('color', '#6c757d');
                                 }
                                 $(id+'Help').text(mylocalStorage[id+'Help']+(maxPasswordLength-mylocalStorage[storage].length));
				if ( onChange != "" )
				        onChange(storage, idStorage0);
                                break;
                 }
                 // Just in case one day we switch back to a div instead of a span
                 $(id).find("br:last-child").remove();
                 return true;
         });

	

         $(id).click(function(e) {
                mylocalStorage['caretPosition'] = getCaretPos($(id).get(0));
         });

	$(id).mouseup(function(e) {
		mylocalStorage['UpMouseX'] = e.clientX;
         });

	$(id).mousedown(function(e) {
		mylocalStorage['DownMouseX'] = e.clientX;
         });

	$(id).mouseout(function(e) {
                mylocalStorage['caretPosition'] = getCaretPos($(id).get(0));
         });

         $(id).keypress(function(e) {
                 if (e.which) {
                          if ( e.which == 13 ) {
                                $(focus).trigger('focus');
                                return false;
                         }

			 if ( typeof mylocalStorage[id+'Help'] === 'undefined' ) {
                                mylocalStorage[id+'Help'] = $(id+'Help').text();
                         }
			 if ( typeof mylocalStorage[storage]  === 'undefined' ) {
				mylocalStorage[storage]='';
			 }
			 if ( (maxPasswordLength-mylocalStorage[storage].length) >= 0 ) {
				 if ( (maxPasswordLength-mylocalStorage[storage].length) < 5 ) {
					$(id+'Help').css('color', '#FF0000');
				 }
				 else
				 {
					$(id+'Help').css('color', '#6c757d');
				 }
				 $(id+'Help').text(mylocalStorage[id+'Help']+(maxPasswordLength-mylocalStorage[storage].length));

	                         var charStr = String.fromCharCode(e.which);
      	                         var myText = $(id).text();
	                         var position = $(id).selectionStart;
	                         $(id).focus();
	                         mylocalStorage['caretPosition'] = getCaretPos($(id).get(0));
	                         mylocalStorage[storage] = mylocalStorage[storage].substring(0, mylocalStorage['caretPosition']) +
	                         charStr + mylocalStorage[storage].substring(mylocalStorage['caretPosition']);
	                         insertTextAtCursor('X');
	                         mylocalStorage['caretPosition'] = getCaretPos($(id).get(0));
				 if ( onChange != "" )
					 onChange(storage, idStorage0);
	                         return false;
			} else
				return false;
                }
         });

}

function myAccountCol1WarningControl()
{
	$('#btnCancel').click( function() {
                        $('#Row1').css('opacity','1');
                        trigger = 0;
                        $('#col3').html('');
                });
        $('#btnConfirmDelete').click( function() {
		// We have a deletion request confirmation
		// Let's send it to the server
		var myJSON;
		var deleteData = $('#customSwitch1').is(":checked");
		myJSON = '{ "CurrentPassword" : "'+ mylocalStorage['MyPasswordDelete'] + '", "DeleteData" : "' + deleteData +'" }';
		Url = '/user/' + mylocalStorage['username'];
	        BuildSignedAuth(Url, 'DELETE' , "application/json", function(authString) {
		$.ajax({
    			type: "DELETE",
                        headers: {
                                "Authorization": "JYP " + mylocalStorage['accessKey'] + ':' + authString['signedString'],
                                "Content-Type" : "application/json",
                                "myDate" : authString['formattedDate']
                        },
			data: myJSON,
	                contentType: 'application/json',
    			url: window.location.origin + '/user/' + mylocalStorage['username'],
			success: function(response){
				// We must disconnect the user 
				// empty the local storage
				// and move back to the home page
				// I must check if the password was good or not, otherwise the account has not been successfully deleted
	                        $('#Row1').css('opacity','1.0');
				if ( response == 'error password' ) {
					form='<center><h1 style="color: #FF0000"> Wrong password - Your account has NOT been deleted </h1>';
	                                form=form+"<h3>Redirecting in 5s<h3>";
       		                        $('#col1').html(form);
                                	$('#col2').html('');
	                                $('#col0').html('');
					$('#col3').html('');
					mylocalStorage['MyPasswordDelete'] = '';
	                                setTimeout(function () {
	                                        myAccount();
	                                }, 5000);
					return;
				}
				for(var propertyName in mylocalStorage) {
	                                delete mylocalStorage[propertyName];
                                }
				form='<center><h1 style="color: #FF0000"> Your account has been successfully deleted </h1>';
                                form=form+"<h3>Redirecting in 5s<h3>";
                                $('#col1').html(form);
                                $('#col2').html('');
                                $('#col0').html('');
				$('#col3').html('');

                                // I must reload the myaccount infrastructure
                                setTimeout(function () {
                                        mainpage();
                                }, 5000);
    			}
        	});
	        });
	});
	managePasswordEntry('#MyPassword','Password','#warningMessage',"","");
}

function myAccountCol1createControl()
{
        $('#btnUpdate').click( function() {
                // I must post the updated data
		var myJSON='{';
		RWFormArray.forEach(function(element) {
			// We must print the variable content except if this is password entry because in that case the real value is
			// within the global vairable array
			if ( element.includes("Password") == true ) {
				myJSON = myJSON + '"' + element + '":"' + mylocalStorage[element] + '"';
			}
			else
			{
				myJSON = myJSON + '"' + element + '":"' + $('#'+element).text() + '"';
			}
			myJSON = myJSON + ',';
		});
		myJSON = myJSON.substr(0, myJSON.length - 1) + '}';
		// I can push to the server the updated data's
		Url = '/user/' + mylocalStorage['username'] + '/updateAccount';
		BuildSignedAuth(Url, 'PUT' , "application/json", function(authString) {
		$.ajax({
                   url: window.location.origin + '/user/' + mylocalStorage['username'] + '/updateAccount',
                   type: 'PUT',
		   headers: {
                                "Authorization": "JYP " + mylocalStorage['accessKey'] + ':' + authString['signedString'],
                                "Content-Type" : "application/json",
                                "myDate" : authString['formattedDate']
                              },
                   data: myJSON,
                   contentType: 'application/json',
                   success: function(response) {
			if ( response == 'email' || response == 'passwordemail' ) {
				// We must clear the window and close the session
				// inform the end user
				form="<center><h1> Your account has been temporarly deactivated </h1>";
				form=form+"<h2>Please revalidate your email<h2>";
				form=form+"<h3>Redirecting in 5s<h3>";
				$('#col1').html(form);
				$('#col2').html('');
				$('#col0').html('');
				disconnect();
			}
			if ( response == 'password' ) {
				form="<center><h1> Password changed successful </h1>";
				form=form+"<h3>Redirecting in 5s<h3>";
				$('#col1').html(form);
				$('#col2').html('');
				$('#col0').html('');

				// We have to update localstorage data with relevant info
				// This is done through a Post with some parameters
				Parameters = 'username=' + mylocalStorage['username'] + '&password=' + mylocalStorage['NewPassword0'];
				var Url = '/user/'+mylocalStorage['username']+'/getToken' + '?' + Parameters;
				BuildSignedAuth(Url, 'POST' , "application/json", function(authString) {
                		$.ajax({
					url: Url,
					type: 'POST',
					headers: {
                                			"Authorization": "JYP " + mylocalStorage['accessKey'] + ':' + authString['signedString'],
			                                "Content-Type" : "application/json",
			                                "myDate" : authString['formattedDate']
                              		},
	                        	success: function postreturn(data) {
	                                	if ( data != "" ) {
	                               		         // Did we got a JSON result ?
	                               		         // if yes we must store the value into RAM
	                                        	// We are catching up the JSON key and store the data into the global locaStorage
	                                        	try {
								var username = mylocalStorage['username'];
								for(var propertyName in mylocalStorage) {
									delete mylocalStorage[propertyName];
								}
								mylocalStorage['username'] = username;
	                                                	var obj = JSON.parse( data );
	                                                	var myarray = Object.keys(obj);
	                                                	for (let i = 0; i  < myarray.length; i++) {
	                                                         	mylocalStorage[myarray[i]] = obj[myarray[i]];
	                                                	}
       		                                         	// we are logged
               		                                 	logged();
                       		                 	} catch(e) {
                               	                 		// This is not a JSON file
                                                		$("#formAnswer").css('color', 'red');
                                                		$("#formAnswer").text(errorMsg);
                                        		}

                                		}
                                		else
                                		{
							// error handling
       		                         	}
                        		},
                		});
				});
				// we probably need to reset th localStorage
			}
			if ( response == 'error password' ) {
				form='<center><h1 style="color: #FF0000"> Password change error </h1>';
                                form=form + "<h2> There was a password missmatch please retry </h2></center>";
                                form=form+"<h3>Redirecting in 5s<h3>";
                                $('#col1').html(form);
                                $('#col2').html('');
                                $('#col0').html('');
				// I must reload the myaccount infrastructure
				setTimeout(function () {
                			myAccount();
        			}, 5000);
			}
                   }
                });
		});
        });
        var trigger=0;
        $('#btnDelete').click(function() {
                if ( trigger == 0 ) {
                        $('#Row1').css('opacity','0.4');
                        // A user account deletion has been requested
                        // Let's load the "form"
                        var Url = '/html/myAccountWarning.html';
                        jQuery.ajaxSetup({async:false});
                        var jqxhr = $.ajax({
					type: "GET",
					url: Url,
                                        success: function postreturn(data) {
    	                                            $('#col3').html(data);
						    myAccountCol1WarningControl();
                                        	}
					});
                        jQuery.ajaxSetup({async:true});

                        mylocalStorage['Password']='';
			managePasswordEntry('#MyPasswordDelete','MyPasswordDelete','#updatePasswordArea',"", "");
                        trigger=1;
                } else {
                        $('#Row1').css('opacity','1');
                        trigger = 0;
                        $('#col3').html('');
                }
        });
	managePasswordEntry('#CurrentPassword','CurrentPassword','#updatePasswordArea',"","");
	managePasswordEntry('#NewPassword0','NewPassword0','#updatePasswordArea',checkPassword, 'NewPassword1');
	managePasswordEntry('#NewPassword1','NewPassword1','#updatePasswordArea',checkPassword, 'NewPassword0');
}

var RWFormArray = [];

function myAccountCol1()
{
        var form='<h1> Welcome back ' + mylocalStorage['username'] + '</h1>';
        form = form + '<form id="myAccount"border-radius:3px; width:90%;">'
        // We have to create a Form where the user can see all the data that we have about him
        var Url = '/user/'+mylocalStorage['username']+'/userGetInfo';
        jQuery.ajaxSetup({async:false});
	BuildSignedAuth(Url, 'GET' , "application/json", function(authString) {
        $.ajax({
		     type: "GET",
		     url: Url, 
		     headers: {
    				"Authorization": "JYP " + mylocalStorage['accessKey'] + ':' + authString['signedString'],
				"Content-Type" : "application/json",
				"myDate" : authString['formattedDate']
  		              },
                     success: function postreturn(data) {
                                if ( data != "" ) {
                                        // Did we got a JSON result ?
                                        // if yes we must store the value into RAM
                                        // We are catching up the JSON key and store the data into the global locaStorage
                                        try {
                                                var label;
                                                var obj = JSON.parse( data );
                                                var myarray = Object.keys(obj);
                                                for (let i = 0; i  < myarray.length; i++) {
                                                        // The last 2 characters are giving us the RW informations
                                                        // There shall be also a Label ;)
                                                        if ( myarray[i].substr(-2) != "RW" ) {
                                                                if ( myarray[i].substr(-5) != "LABEL" ) {
                                                                        form = form + '<div class="form-group" style="margin-bottom:0rem">'
                                                                        form =  form + '<label for="'+myarray[i]+'"class="col-sm-2 col-form-label" style="font-weight:bold;\
                                                                                padding-left:0px">'+ myarray[i]+':</label>\n';
                                                                        if (( obj[myarray[i]+"LABEL"] != "" ) && ( typeof obj[myarray[i]+"LABEL"] != 'undefined'))  {
                                                                                label = obj[myarray[i]+"LABEL"];
                                                                        } else {
                                                                                label = "";
                                                                        }
                                                                        if ( obj[myarray[i]+"RW"] == "1" ) {
                                                                                form = form + '<div contenteditable="true" class="form-control" id="'+myarray[i]+
                                                                                '" style=" padding:.0rem 0.75rem;display:inline;border:0px; border-bottom:2px solid;\
                                                                                background-color:#ffffff" aria-describedby="'+myarray[i]+'Help">'+ obj[myarray[i]]+'</div>';
                                                                                form = form + '<small id="'+myarray[i]+'Help" class="form-text text-muted">'+label+'</small>';
										RWFormArray.push(myarray[i]);
                                                                        } else {
                                                                                form = form + '<div contenteditable="false" class="form-control" id="'+myarray[i]+
                                                                                '" style="font-style:italic; color: #888888; padding:.0rem 0.75rem;display:inline;border:0px;\
                                                                                border-bottom:2px solid; background-color:#ffffff" aria-describedby="'+myarray[i]+'Help">'+
                                                                                obj[myarray[i]]+'</div>';
                                                                                form = form + '<small id="'+myarray[i]+'Help" class="form-text text-muted">'+label+'</small>';
                                                                        }
                                                                        form = form + '</div>'
                                                                }
                                                        }

                                                }

						// Must add an update Password opportunity
						$.ajax({
							type: "GET",
							url: '/html/updatePassword.html',
						        success: function getUpdatePassword(data) {
									form = form + data;
									form = form + '</center>'
               			                              		form = form+'</form></div> <button type="button" class="btn btn-success" id="btnUpdate">Update</button>\
		               		                                       <button type="button" class="btn btn-danger pull-right" id="btnDelete">Delete Account</button>';
		                       		                         $('#col1').html(form);
									var currentList=document.querySelectorAll('[id*="Password"]');
									for (var element of currentList) {
										if ( element.tagName == "SPAN" ) {
											if ( element.contentEditable == "true" ) {
												RWFormArray.push(element.id);
											}
										}
									}
			                                           myAccountCol1createControl();
								}
						        });
						// We must add the various imported div to the RWFormArray
                                                // we are logged
                                        } catch(e) {
                                                /*
                                                        // This is not a JSON file
                                                        $("#formAnswer").css('color', 'red');
                                                        ("#formAnswer").text(errorMsg);
                                                */
                                        }

                                }
                                else
                                {
                                    /*  $("#formAnswer").css('color', 'green');
                                        $("#formAnswer").text(successMsg);
                                        $("#btn1").hide();
                                    */
                                }
                        },
//Catch failure
	                error: function (xhr, ajaxOptions, thrownError) {
       		                 $("#formAnswer").css('color', 'red');
               		         $("#formAnswer").text("Auth Error");
                	}
                });
	});

}

function myAccountCol2() {
        avatar='<center><div style="font-weight:bold; text-decoration:underline">Your profile picture</div>' +
            '<div class="avatar-upload">'+
                '<div class="avatar-edit">'+
                    '<input type="file" id="imageUpload" accept=".png, .jpg, .jpeg" />'+
                    '<label for="imageUpload"></label>'+
                '</div>'+
                '<div class="avatar-preview">'+
                '    <div id="imagePreview" style="">'+
                '    </div>'+
                '</div>'+
            '</div>'+
        '</div></center>';
        $('#col2').html(avatar);
        jQuery.ajaxSetup({async:true});

        $('#Password').prop('disabled', true);
        $('#Password').css('font-style', 'italic');

        $('#Nickname').prop('contenteditable', false);
        $('#TokenType').prop('disabled', true);
        $('#TokenAuth').prop('disabled', true);
        $('#TokenSecret').prop('disabled', true);
        $('#CreationDate').prop('disabled', true);
        $('#Lastlogin').prop('disabled', true);
        $('#Active').prop('disabled', true);
        $('#ValidationString').prop('disabled', true);

        loadCSS("css/avatar.css");
        loadJS("js/avatar.js");
}

myAccountCol0();
myAccountCol1();
myAccountCol2(); // Let's display the Avatar
