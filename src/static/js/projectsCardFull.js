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

function loadJS(filename){
        jQuery.ajaxSetup({async:false});
        jQuery.getScript(filename, function(data, textStatus, jqxhr) {
                });
        jQuery.ajaxSetup({async:true});
}

function ProjectCardFull(cardImage, xeoglCode, Date, Name, Revision, Owner, Revisions) {

        clearDocument();
        loadHTML("navbar.html");
        loadJS("js/navbar.js");
        navbarHover();
        loginBtn();

        // We must put in place the layout here and allow various entries to be available
        // The first one is personal settings but others might be coming up
        var layout = '  <div class="container-fluid" style="padding-left:2px;"><div class="row" id="Title">\
                        <div class="col offset col0" style="width:20%; position:fixed; overflow-y:scroll; height:100%;" id="title0"></div>\
                        <h2><div closs="col " style="width:80%; right:0px; position:fixed;" id="title1"></div></h2>\
                        </div>\
                        <div class="container-fluid" style="padding-left:2px;"><div class="row" id="Row1">\
                        <div class="col offset col0" style="width:20%; position:fixed; overflow-y:scroll; height:calc(100vh - 80px); top:80px" id="col0"></div>\
                        <div class="container-fluid" style="padding-left:2px;"><div class="row" id="Row1">\
                        <div class="col offset col0" style="width:20%; position:fixed; overflow-y:scroll; height:calc(100vh - 80px); top:80px" id="col0"></div>\
                        <div closs="col " style="width:80%; right:0px; position:fixed; overflow-y:scroll; height:calc(100vh - 80px); top:80px" id="col1"></div>\
                        <div class="row"><div class="col" style="width:100%" id="col3"></div></div>';
        $(document.body).append(layout);
        loadCSS("css/projects.css");

        $('#Btn'+Date+'-'+Owner+'-'+Name+'-'+Revision).remove();


        player = '<div class="embed-responsive embed-responsive-16by9 shadow-lg p-3 mb-5 bg-white rounded" '+
                'id="myplayer" style="margin:auto;display:none;"><iframe class="embed-responsive embed-responsive-16by9 shadow-lg p-3 mb-5 bg-white rounded" id="Player">'+
                '</iframe></div>';

        card =  '<div id="player"><img class="embed-responsive embed-responsive-16by9 shadow-lg p-3 mb-5 bg-white rounded" id="myimage'+Date+'-'+Owner+'-'+Name+'-'+Revision+
                '" src="" alt="Card image cap" style="margin-bottom:2rem; ">' + player + '<form class="form-inline">' +
                '<div id="Btn'+Date+'-'+Owner+'-'+Name+'-'+Revision+'" class="btn btn-primary" style="margin-bottom:2rem;">3D View</div>'

        card = card +
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
