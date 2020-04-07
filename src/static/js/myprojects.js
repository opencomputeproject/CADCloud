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


var cardPrivacy = [];

function addCard(cardImage, xeoglCode, Date, Name, Revision, Owner, Revisions, Dates, Private) {

	$('#Btn'+Date+'-'+Owner+'-'+Name+'-'+Revision).remove();


	player = '<div class="embed-responsive embed-responsive-16by9 shadow-lg p-3 mb-5 bg-white rounded" '+
                'id="myplayer" style="margin:auto;display:none;"><iframe class="embed-responsive embed-responsive-16by9 shadow-lg p-3 mb-5 bg-white rounded" id="Player">'+
                '</iframe></div>';

	card =	'<div id="player"><img class="embed-responsive embed-responsive-16by9 shadow-lg p-3 mb-5 bg-white rounded" id="myimage'+Date+'-'+Owner+'-'+Name+'-'+Revision+
		'" src="" alt="Card image cap" style="margin-bottom:2rem; ">' + player + '<form class="form-inline">' +
		'<div id="Btn'+Date+'-'+Owner+'-'+Name+'-'+Revision+'" class="btn btn-primary" style="margin-bottom:2rem;">3D View</div>' +
		'<div id="BtnDeleteRevision'+Date+'-'+Owner+'-'+Name+'-'+Revision+'" class="btn btn-primary btn-warning" style="margin-bottom:2rem; margin-left:5px">Delete Revision</div>' +
		'<div id="BtnDeleteProject'+Date+'-'+Owner+'-'+Name+'-'+Revision+'" class="btn btn-primary btn-danger" style="margin-bottom:2rem; margin-left:5px;">Delete Project</div>' +
		'<div class="custom-control custom-switch" style="position:absolute; right:0px; margin-top:-15px;">';

	if ( Private == "0" )
	{
		card = card +  '<input type="checkbox" class="custom-control-input" id="customSwitch-'+Date+'-'+Owner+'-'+Name+'-'+Revision+'" checked> ';
	}
	else
		card = card +  '<input type="checkbox" class="custom-control-input" id="customSwitch-'+Date+'-'+Owner+'-'+Name+'-'+Revision+'" > ';
	cardPrivacy[Date+'-'+Owner+'-'+Name] = Private;

        card = card + '	<label class="custom-control-label" for="customSwitch-'+Date+'-'+Owner+'-'+Name+'-'+Revision+'">Make project Public</label>'+
        	'</div></div>'+
        	'</form>';

	// Let's handle the revision and display them through a slider. 
	// Each time we change the value we must update the associated fields (image and player code)

	if ( Revisions.length > 1 ) {
		RevisionsIndex = "[";
		RevisionsLabel = "[";
		for ( let i = 0 ; i < Revisions.length ; i++ ) {
			RevisionsIndex = RevisionsIndex + Revisions[i];
			RevisionsLabel = RevisionsLabel + '"' + Revisions[i] + '"';
			if ( i < (Revisions.length - 1) ) {
				RevisionsIndex = RevisionsIndex + ",";
				RevisionsLabel = RevisionsLabel + ",";
			}
		}
		RevisionsIndex = RevisionsIndex + "]";
		RevisionsLabel = RevisionsLabel + "]";

		slider = '<div class="col"><input id="ex1'+Date+'-'+Owner+'-'+Name+'-'+Revision+
			'" data-slider-id="ex1Slider" type="text"   data-slider-value="14" data-slider-ticks="'+RevisionsIndex+'"' +
	                  'data-slider-ticks-labels="'+RevisionsLabel+'" style=".tooltip.in {opacity: 1;}"/>';

	}

	$('#col1').html(card);

	if ( Revisions.length > 1 ) {
		$('#col1').append(slider);
	}

	$('#customSwitch-'+Date+'-'+Owner+'-'+Name+'-'+Revision).on("change", function() {
		if ( $('#customSwitch-'+Date+'-'+Owner+'-'+Name+'-'+Revision).is(':checked') ) {
			// We need to push the project to the public status
			Url = '/projects/public/'+Date+'/'+Owner+'/'+Name+'/'
			// We must send back the revision list
			contentToSend = JSON.stringify(Revisions);
			contentToSend = contentToSend + '\n' + JSON.stringify(Dates);
			cardPrivacy[Date+'-'+Owner+'-'+Name] = "0";
		        BuildSignedAuth(Url, 'PUT' , "application/json", function(authString) {
	       		        $.ajax({
	                        	type: "PUT",
		                        headers: {
		                                "Authorization": "JYP " + mylocalStorage['accessKey'] + ':' + authString['signedString'],
		                                "Content-Type" : "application/json",
		                                "myDate" : authString['formattedDate']
		                        },
					data: contentToSend,
		                        url: Url,
		                        success: function postreturn(data) {
						}
				});
			});
		} else {
			// We need to push the project to the private status
			Url = '/projects/private/'+Date+'/'+Owner+'/'+Name+'/'
			contentToSend = JSON.stringify(Revisions);
                        contentToSend = contentToSend + '\n' + JSON.stringify(Dates);
			cardPrivacy[Date+'-'+Owner+'-'+Name] = "1";
                        BuildSignedAuth(Url, 'PUT' , "application/json", function(authString) {
                        	$.ajax({
	                                type: "PUT",
	                                headers: {
	                                        "Authorization": "JYP " + mylocalStorage['accessKey'] + ':' + authString['signedString'],
	                                        "Content-Type" : "application/json",
	                                        "myDate" : authString['formattedDate']
	                                },
					data: contentToSend,
	                               url: Url,
	                               success: function postreturn(data) {
       		                                 }
			        });		     
                        });
		}
		// Now we need to update the view
		// Roughly the menu list in col0
		
	});



	jQuery("#myimage"+Date+'-'+Owner+'-'+Name+'-'+Revision).attr('src', 'data:image/png;base64,' + cardImage);

	$('#Btn'+Date+'-'+Owner+'-'+Name+'-'+Revision).click( function(e) {
		if ( $('#Btn'+Date+'-'+Owner+'-'+Name+'-'+Revision).text() == "3D View" ) {
	                e.stopPropagation();
			var originalScrollTo = window.scrollTo;
			window.scrollTo = function () {};
			if ( ! $('#Player')[0].hasAttribute("srcdoc")) {
				$('#Player').attr("srcdoc",xeoglCode);
				$('#Player').addClass("animated fadeIn");
			}
			$('#myplayer').css("display","");
			$("#myimage"+Date+'-'+Owner+'-'+Name+'-'+Revision).css("display","none");
			window.scrollTo = originalScrollTo;
	                playerdisplayed = 1;
			keepplayer = 1;
			$('#Btn'+Date+'-'+Owner+'-'+Name+'-'+Revision).text("2D View");
		} else
			if ( $('#Btn'+Date+'-'+Owner+'-'+Name+'-'+Revision).text() == "2D View" ) {
				$('#myplayer').css("display","none");
				$("#myimage"+Date+'-'+Owner+'-'+Name+'-'+Revision).css("display","");
			
				$('#Btn'+Date+'-'+Owner+'-'+Name+'-'+Revision).text("3D View");
			}
	});




  if ( Revisions.length > 1 ) {

	  var r = $('#ex1'+Date+'-'+Owner+'-'+Name+'-'+Revision).slider({
	    tooltip: 'always',
	    lock_to_ticks: true,
	    formatter: function(val) {
	      return val;
    	}
  	});
	$('#ex1'+Date+'-'+Owner+'-'+Name+'-'+Revision).slider().on('click', function(e) {
		e.stopPropagation();
	});
}


}

function updateContent(Date, Owner, Name, Revisions, Private)
{
	var currentRevision = 0;
	var index = 0;
	for ( let l = 0 ; l < Revisions.length ; l++ ) {
		if ( Revisions[l] > currentRevision ) {
			currentRevision = Revisions[l];
			index = l ;
		}
	}

	var magnetUrl = '/projects/'+mylocalStorage['username']+'/getMagnet/'+Date[index]+'/'+
			Owner+'/'+Name+'/'+
			Revisions[index] + '/' + Private;

	var playerCode = '/projects/'+mylocalStorage['username']+'/getPlayerCode/'+Date[index]+'/'+
			Owner+'/'+Name+'/'+
			Revisions[index] + '/' + Private;

	var cardImage="";
	var xeoglCode="";

	// We must get the data with our current credentials
	Url = window.location.origin + magnetUrl;
	BuildSignedAuth(magnetUrl, 'GET' , "image/png", function(authString){
		$.ajax({
			headers: {
				"Authorization": "JYP " + mylocalStorage['accessKey'] + ':' + authString['signedString'],
				"Content-Type" : "image/png",
				"myDate" : authString['formattedDate']
			},
			url: magnetUrl,
			type: 'GET',
			success: function(response) {
				cardImage = response;

				Url = window.location.origin + playerCode;
				BuildSignedAuth(playerCode, 'GET' , "application/html", function(authString){
					$.ajax({
						headers: {
                         			       "Authorization": "JYP " + mylocalStorage['accessKey'] + ':' + authString['signedString'],
			                                "Content-Type" : "application/html",
			                                "myDate" : authString['formattedDate']
			                        },
						url: playerCode,
						type: 'GET',
						success: function(response) {
							xeoglCode = response;
							addCard(cardImage, xeoglCode, Date[index],
								Name, Revisions[index],
								Owner, Revisions, Date, Private);
							}
						});
					});
				}
		});
	});

}

function myProjectsCol1() {
	var Url = '/projects/' + mylocalStorage['username'] + '/getList';
	loadCSS("css/avatarProject.css");
	
        var form="<h2> Your project:</h2>";

	$('#title0').html(form);	
	jQuery.ajaxSetup({async:false});
	$('#projects').prepend('<div class="container-fluid">');
	BuildSignedAuth(Url, 'GET' , "application/json", function(authString) {
                $.ajax({
                        type: "GET",
                        headers: {
                                "Authorization": "JYP " + mylocalStorage['accessKey'] + ':' + authString['signedString'],
                                "Content-Type" : "application/json",
                                "myDate" : authString['formattedDate']
                        },
	               url: Url,
	               success: function postreturn(data) {
				// We are getting Json content which is a list of
				// public project
				var obj = JSON.parse( data );
	                        var myarray = Object.keys(obj);
			
				$('#col0').append('<ul class="list-group" id="myprojectlist">');
				for (let i = 0; i  < obj["Entries"].length; i++) {
					$('#myprojectlist').append('<li class="list-group-item list-group-item-action" id="menu'+obj["Entries"][i]["Date"][0]+obj["Entries"][i]["Owner"]+
							    obj["Entries"][i]["Name"]+'">'+obj["Entries"][i]["Name"]+'</li>');
					$('#menu'+obj["Entries"][i]["Date"][0]+obj["Entries"][i]["Owner"]+obj["Entries"][i]["Name"]).on("click", function(e) {
						// Clean up previous element
						$('#myprojectlist').children().each(function () {
					        	var $currentElement = $(this);
							$currentElement.removeClass("active");
						        // Show events handlers of current element
						});
					// Change title and turn on clicked element
					$('#title1').html(obj["Entries"][i]["Name"]);	
					$('#menu'+obj["Entries"][i]["Date"][0]+obj["Entries"][i]["Owner"]+obj["Entries"][i]["Name"]).addClass("active");
					// Update page content

					updateContent(obj["Entries"][i]["Date"], obj["Entries"][i]["Owner"], obj["Entries"][i]["Name"], obj["Entries"][i]["Revisions"], obj["Entries"][i]["Private"]);
					
					});
				}

				var i =0;

				$('#title1').html(obj["Entries"][i]["Name"]);	
				$('#menu'+obj["Entries"][i]["Date"][0]+obj["Entries"][i]["Owner"]+obj["Entries"][i]["Name"]).addClass("active");
				updateContent(obj["Entries"][i]["Date"], obj["Entries"][i]["Owner"], obj["Entries"][i]["Name"], obj["Entries"][i]["Revisions"],  obj["Entries"][i]["Private"]);
	       		}
        	});
	});
	jQuery.ajaxSetup({async:true});
}
playerdisplayed=0;
myProjectsCol1();

