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


var mylocalStorage = {};
window.mylocalStorage = mylocalStorage;

function clearDocument(){
	$(document.body).empty();
}

function loadHTML(filename){
	jQuery.ajaxSetup({async:false});
        jQuery.get(filename, function(data, status){
                $(document.body).append(data);
        });
        jQuery.ajaxSetup({async:true});
}

function getHTML(filename){
        jQuery.ajaxSetup({async:false});
        jQuery.get(filename, function(data, status){
        	jQuery.ajaxSetup({async:true});
		return(data);
        });
}

function loadCSS(filename){
	jQuery.ajaxSetup({async:false});
        jQuery.get(filename, function(data, status){
	$("<style>").prop("type", "text/css").html(data).appendTo("head");
//                $(document.body).append(data);
        });
        jQuery.ajaxSetup({async:true});
}

function loadJS(filename){
        jQuery.ajaxSetup({async:false});
        jQuery.getScript(filename, function(data, textStatus, jqxhr) {
                });
        jQuery.ajaxSetup({async:true});
}

var getUrlParameter = function getUrlParameter(sParam) {
    var sPageURL = window.location.search.substring(1),
        sURLVariables = sPageURL.split('&'),
        sParameterName,
        i;

    for (i = 0; i < sURLVariables.length; i++) {
        sParameterName = sURLVariables[i].split('=');

        if (sParameterName[0] === sParam) {
            return sParameterName[1] === undefined ? true : decodeURIComponent(sParameterName[1]);
        }
    }
};

function logged()
{
	mainpage();
}

function disconnect()
{
	delete mylocalStorage['accessKey'];
	delete mylocalStorage['secretKey'];
	delete mylocalStorage['username'];
	// Wait 5s and redirect to mainpage
	setTimeout(function () {
		mainpage();
    	}, 5000);
}

function insertTextAtCursor(text) {
    var sel, range, textNode;
    if (window.getSelection) {
        sel = window.getSelection();
        if (sel.getRangeAt && sel.rangeCount) {
            range = sel.getRangeAt(0).cloneRange();
            range.deleteContents();
            textNode = document.createTextNode(text);
            range.insertNode(textNode);
            range.setStart(textNode, textNode.length);
            range.setEnd(textNode, textNode.length);
            sel.removeAllRanges();
            sel.addRange(range);
        }
    } else if (document.selection && document.selection.createRange) {
        range = document.selection.createRange();
        range.pasteHTML(text);
    }
}


function myAccount()
{
	clearDocument();
	loadHTML("navbar.html");
        loadJS("js/navbar.js");
	navbarHover();
	loginBtn();

	// We must put in place the layout here and allow various entries to be available
	// The first one is personal settings but others might be coming up
	var layout = '<div class="container-fluid"><div class="row" id="Row1">\
			<div class="col" style="width:10%" id="col0"></div>\
			<div closs="col" style="width:60%" id="col1"></div>\
			<div class="col" style="width:10%" id="col2"></div></div>\
			<div class="row"><div class="col" style="width:100%" id="col3"></div></div>';
        $(document.body).append(layout);

	loadJS("js/myaccount.js");
}

function myProjects()
{
        clearDocument();
        loadHTML("navbar.html");
        loadJS("js/navbar.js");
        navbarHover();
        loginBtn();

        // We must put in place the layout here and allow various entries to be available
        // The first one is personal settings but others might be coming up
        var layout = '	<div class="container-fluid" style="padding-left:2px;"><div class="row" id="Title">\
			<div class="col offset col0" style="width:20%; position:fixed; overflow-y:scroll; height:100%;" id="title0"></div>\
			<h2><div closs="col " style="width:80%; right:0px; position:fixed;" id="title1"></div></h2>\
			</div>\
			<div class="container-fluid" style="padding-left:2px;"><div class="row" id="Row1">\
                        <div class="col offset col0" style="width:20%; position:fixed; overflow-y:scroll; height:calc(100vh - 140px); top:140px" id="col0"></div>\
                        <div closs="col " style="width:80%; right:0px; position:fixed; overflow-y:scroll; height:calc(100vh - 140px);top:140px" id="col1"></div>\
                        <div class="row"><div class="col" style="width:100%" id="col3"></div></div>';
        $(document.body).append(layout);
	loadCSS("css/projects.css");
        loadJS("js/myprojects.js");
}



function bufferToBase64(buf) {
    var binstr = Array.prototype.map.call(buf, function (ch) {
        return String.fromCharCode(ch);
    }).join('');
    return btoa(binstr);
}

function b64tob64u(a){
    a=a.replace(/\=/g,"");
    a=a.replace(/\+/g,"-");
    a=a.replace(/\//g,"_");
    return a
}

function BuildSignedAuth(uri, op, contentType, callback) {
	var returnObject = {};
	var currentDate = new Date;
        var formattedDate = currentDate.toGMTString().replace( /GMT/, '+0000');
	var stringToSign = op +'\n\n'+contentType+'\n'+formattedDate+'\n'+uri
	returnObject['formattedDate'] = formattedDate;
        const buffer = new TextEncoder( 'utf-8' ).encode( stringToSign );
	if ( mylocalStorage['secretKey'] !== undefined && mylocalStorage['secretKey'].length > 0)
	{

		var hash = CryptoJS.HmacSHA1(stringToSign, mylocalStorage['secretKey']);
		returnObject['signedString'] = CryptoJS.enc.Base64.stringify(hash);
	}
	else
		returnObject['signedString'] = '';
	callback(returnObject);
//	return(returnObject);
}

function mainpage(){
	clearDocument();
	// Must load the default home page
	loadHTML("navbar.html");
	loadJS("js/navbar.js");
	navbarHover();
	loginBtn();
	loadHTML("home.html");

	if (( "string" !== typeof(mylocalStorage['secretKey']) ) & ( "string" !== typeof(mylocalStorage['accessKey']) ))
	{
		$('#signup').css("display", "");
	}

	loadJS("js/projects.js");
	loadJS("js/forms.js");
	loadJS("js/base.js");
	loadHTML("footer.html");
	formSubmission('#signup','createUser','User created - Please check your email','User exist');
}

function main(){
// Let's empty the document first
	if ( getUrlParameter('loginValidated') == "1" )
	{
		// We must check if the registration is ok
		loadHTML("navbar.html");
		loadJS("js/navbar.js");
		navbarHover();
		loginBtn();
                $(document.body).append("<center><h1>Welcome Back !</h1></center>");
		loadHTML("loginForm.html");
		loadJS("js/login.js");
		managePasswordForgotten();
		loadJS("js/forms.js");
		formSubmission('#login','getToken','','Password missmatch');
		loadHTML("footer.html");
	}
	else
	{
		// We can't reset without the validation string otherwise
		// somebody could hijack user account

		if ( getUrlParameter('resetPassword') == "1" )
		{
			loadHTML("navbar.html");
			loadJS("js/navbar.js");
			navbarHover();
			loginBtn();
			$(document.body).append("<center><h1>Welcome Back !</h1><center>");
			loadHTML("resetPassword.html");
			$('#username').val(getUrlParameter('username'));
			$('#username').prop('disabled', true);
			$('#validation').val(getUrlParameter('validation'));
                        $('#validation').prop('disabled', true);
			loadJS("js/forms.js");
			formSubmission('#resetPassword','resetPassword','Password successfully reset','Reset link expired');
			loadHTML("footer.html");

		}
		else
			if (( "string" === typeof(mylocalStorage['secretKey']) ) & ( "string" === typeof(mylocalStorage['accessKey']) ))
			{
				logged();
			}
			else
			{
				mainpage();

			}
	}
}
