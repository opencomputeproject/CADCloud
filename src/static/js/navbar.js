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

function navbarHover() {
$('body').on('mouseover mouseout', '.dropdown', function(e) {
    var dropdown = $(e.target).closest('.dropdown');
    var menu = $('.dropdown-menu', dropdown);
    dropdown.addClass('show');
    menu.addClass('show');
    setTimeout(function () {
        dropdown[dropdown.is(':hover') ? 'addClass' : 'removeClass']('show');
        menu[dropdown.is(':hover') ? 'addClass' : 'removeClass']('show');
    }, 300);
});
}

function loginBtn() {
$('#loginNavbar').on('click', function(e) {
	if (( "string" === typeof(mylocalStorage['secretKey']) ) & ( "string" === typeof(mylocalStorage['accessKey']) ))
	{
		disconnect();
	}
	else
	{
		clearDocument();
		loadHTML("navbar.html");
		loadJS("js/navbar.js");
		navbarHover();
		loginBtn();
       		$(document.body).append("<center><h1>Welcome Back !</h1><center>");
	       	loadHTML("loginForm.html");
       		loadJS("js/login.js");
        	managePasswordForgotten();
        	loadJS("js/forms.js");
        	formSubmission('#login','getToken','','Password missmatch');
        	loadHTML("footer.html");
	}
});

$('#MyAccount').on('click', function(e) {
	myAccount();
});

$('#MyProjects').on('click', function(e) {
        myProjects();
});
	// We must check if we are logged in or not ?
	// and replace the button text
	if (( "string" === typeof(mylocalStorage['secretKey']) ) & ( "string" === typeof(mylocalStorage['accessKey']) ))
	{
	        // we must change the login button by a Disconnect button
	        $('#loginNavbar').html('Logout');
		$('#navbarDropdownMenuLink').show();
		// The navBar title must be the login name
		$('#navbarDropdownMenuLink').html(mylocalStorage['username']);
	}
	else
		$('#navbarDropdownMenuLink').hide();
}

$("#Home").on("click", function(event) {
	mainpage();
});
