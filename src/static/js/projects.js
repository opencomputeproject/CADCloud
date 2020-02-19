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


$(document).unbind("click");
$(document).click(function(ee) {
	if (  playerdisplayed == 1 && keepplayer == 0 ) {
		$('#Player').addClass("animated fadeOut");
		$('#myplayer').css("display","none");
		playerdisplayed=0;
	}
});

function addCard(cardImage, cardAvatar, xeoglCode, Date, Name, Revisions, Owner) {

	
	card = '<div id="myCard' + Date+'-'+Owner+'-'+Name+'-'+Revisions+
			'" class="card text-center" style="display:inline-block; width: 18rem;">' +
			'<img class="card-img-top" id="myimage'+Date+'-'+Owner+'-'+Name+'-'+Revisions+'" src="" alt="Card image cap">' +
			'<div id="myBody'+Date+'-'+Owner+'-'+Name+'-'+Revisions+'" class="card-body">' +
				'<h5 class="card-title">'+ Name + ' Rev ' + Revisions +'</h5>' +
				'<p class="card-text">Some quick example text to build on the card title and make up the bulk of the card content.</p>' +
				'<div id="cardBtn'+Date+'-'+Owner+'-'+Name+'-'+Revisions+'" class="btn btn-primary" >Go somewhere</div>' +
			'</div>' +
		'</div>';

	$('#projects').append(card);
	jQuery("#myimage"+Date+'-'+Owner+'-'+Name+'-'+Revisions).attr('src', 'data:image/png;base64,' + cardImage);

	$('#cardBtn'+Date+'-'+Owner+'-'+Name+'-'+Revisions).click( function(e) {
		// We do not propagate back the event to the document otherwise 
		// the player is going to be hidden back
		e.stopPropagation();
		var originalScrollTo = window.scrollTo;
		window.scrollTo = function () {};
		$('#Player').attr("srcdoc",xeoglCode);
		$('#Player').addClass("animated fadeIn");
		$('#myplayer').css("display","");
		playerdisplayed = 1;
		window.scrollTo = originalScrollTo;
		clickcount=clickcount+1;
	});

	avatar='<div class="avatarProject-upload">'+
		'<div class="avatarProject-preview">'+
		'    <div id="imagePreview'+Date+'-'+Owner+'-'+Name+'-'+Revisions+'" style="">'+
		'    </div>'+
			'<div style="font-size: 0.5rem;font-weight:bold; text-decoration:underline">'+Owner+'</div>' +
		'</div>'+
	'</div>';
	$('#myBody'+Date+'-'+Owner+'-'+Name+'-'+Revisions).prepend(avatar);
	jQuery("#imagePreview"+Date+'-'+Owner+'-'+Name+'-'+Revisions).css('background-image', 'url("data:image/png;base64,' + cardAvatar + '")');

}


function setupHomePage() {
	var Url = '/projects/getList';
	loadCSS("css/avatarProject.css");
        player = '<div class="embed-responsive embed-responsive-16by9 shadow-lg p-3 mb-5 bg-white rounded" '+
                'id="myplayer" style="width:80%;margin:auto;display:none;"><iframe class="embed-responsive-item" id="Player">'+
                '</iframe></div>';
        $('#player').append(player);
	jQuery.ajaxSetup({async:false});
	var jqxhr = $.ajax({
               type: "GET",
               url: Url,
               success: function postreturn(data) {
			// We are getting Json content which is a list of
			// public project
			var obj = JSON.parse( data );
                        var myarray = Object.keys(obj);
                        for (let i = 0; i  < obj["Entries"].length; i++) {
				var currentRevision = 0;
				var index = 0;
				for ( let l = 0 ; l < obj["Entries"][i]["Revisions"].length ; l++ ) {
					if ( parseInt(obj["Entries"][i]["Revisions"][l]) > currentRevision ) {
						currentRevision = obj["Entries"][i]["Revisions"][l];
						index = l ;
					}
				}
	 			// We need to display the project magnet
				// the code is going to be seating into #projects div
				var magnetUrl = '/projects/getMagnet/'+obj["Entries"][i]["Date"][index]+'/'+
						obj["Entries"][i]["Owner"]+'/'+obj["Entries"][i]["Name"]+'/'+
						obj["Entries"][i]["Revisions"][index];

				var magnetAvatar = '/projects/getAvatar/'+obj["Entries"][i]["Date"][index]+'/'+
                                                obj["Entries"][i]["Owner"]+'/'+obj["Entries"][i]["Name"]+'/'+
                                                obj["Entries"][i]["Revisions"][index];

				var playerCode = '/projects/getPlayerCode/'+obj["Entries"][i]["Date"][index]+'/'+
						obj["Entries"][i]["Owner"]+'/'+obj["Entries"][i]["Name"]+'/'+
						obj["Entries"][i]["Revisions"][index];

				var cardImage="";
				var cardAvatar="";
				var xeoglCode="";


				$.ajax({
				       url: window.location.origin + magnetUrl,
				       type: 'GET',
				       success: function(response) {
				       			cardImage = response;
							$.ajax({
			                                       url: window.location.origin + magnetAvatar,
			                                       type: 'GET',
			                                       success: function(response) {
		                                                        cardAvatar = response;
									$.ajax({
					                                       url: window.location.origin + playerCode,
					                                       type: 'GET',
					                                       success: function(response) {
				                                                        xeoglCode = response;
											addCard(cardImage, cardAvatar, xeoglCode, obj["Entries"][i]["Date"][index],
												obj["Entries"][i]["Name"], obj["Entries"][i]["Revisions"][index],
												obj["Entries"][i]["Owner"]);
						                                       }
						                                });
			                                       }
		                                	});
       				       		}
				});
				
                        }
			// $('#projects').html(data);
                      }
        });
	jQuery.ajaxSetup({async:true});
}
var clickcount=0;
var playerdisplayed=0;
var keepplayer=0;
playerdisplayed=0;
clickcount=0;
setupHomePage()

