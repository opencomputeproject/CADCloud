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

function formSubmission(id, fn, successMsg, errorMsg) {
        // This is to detect the form submission 
        $(id).submit(function (e) {
		
                // To avoid the default submission
                e.preventDefault();

		// Looping over the form elements to build the parameter string
		var Parameters = "";
		var username = "";

		$($(id).prop('elements')).each(function(){
		    if ( this.type != "submit" )
		    {
			    if ( Parameters.length > 1 )
				    Parameters = Parameters + '&' + this.placeholder +'=' + this.value ;
			    else
				    Parameters = Parameters + this.placeholder +'=' + this.value ;
			    if ( this.placeholder == "username" )
			    {
				    mylocalStorage['username'] = this.value;
				    username = this.value;
			    }
		    }
		});
	
                var Url = '/user/'+username+'/'+fn;
                var jqxhr = $.post(Url, Parameters,
                        function postreturn(data) {
				if ( data != "" ) {
					// Did we got a JSON result ?
					// if yes we must store the value into RAM
					// We are catching up the JSON key and store the data into the global locaStorage
					try {
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
			                $("#formAnswer").css('color', 'green');
			                $("#formAnswer").text(successMsg);
			                $("#btn1").hide();
        			}
			},
                        'text'
		);
		jqxhr.fail( function() {
			$("#formAnswer").css('color', 'red');
			$("#formAnswer").text("Auth Error");
		});
        });
}

