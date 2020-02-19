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


function readURL(input) {
    if (input.files && input.files[0]) {
        var reader = new FileReader();
        reader.onload = function(e) {
            $('#imagePreview').css('background-image', 'url('+e.target.result +')');
            $('#imagePreview').hide();
            $('#imagePreview').fadeIn(650);
	    Url = '/user/' + mylocalStorage['username'] + '/updateAvatar';
	    BuildSignedAuth(Url, 'PUT' , "image/jpg", function(authString) {
	    $.ajax({
	           url: window.location.origin + '/user/' + mylocalStorage['username'] + '/updateAvatar',
	           type: 'PUT',
		   headers: {
	                "Authorization": "JYP " + mylocalStorage['accessKey'] + ':' + authString['signedString'],
	                "Content-Type" : "image/jpg",
	                "myDate" : authString['formattedDate']
                   },
	           data: e.target.result,
	           contentType: 'image/jpg',
	           success: function(response) {
        	   }
        	});
             });
        }
        reader.readAsDataURL(input.files[0]);
    }
}
$("#imageUpload").change(function() {
    readURL(this);
});

// We can initialize the content
Url ='/user/' + mylocalStorage['username'] + '/getAvatar';
BuildSignedAuth(Url, 'GET' , "application/json", function(authString) {
$.ajax({
       url: window.location.origin + '/user/' + mylocalStorage['username'] + '/getAvatar',
       type: 'GET',
       headers: {
		"Authorization": "JYP " + mylocalStorage['accessKey'] + ':' + authString['signedString'],
		"Content-Type" : "application/json",
		"myDate" : authString['formattedDate']
                },
       success: function(response) {
			jQuery("#imagePreview").css('background-image', 'url("data:image/png;base64,' + response + '")');
       }
});
});

