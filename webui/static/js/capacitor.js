	$( "#heuristic" ).change(function() {
		if ('policy' == $(this).val()){
			$( "#workload_approach" ).removeAttr('disabled');
                        $( "#configuration_approach" ).removeAttr('disabled');
		}else{
			$( "#workload_approach" ).attr('disabled', 'disabled');
			$( "#configuration_approach" ).attr('disabled', 'disabled');
		}
	});

        function closeAlertPanel(css){
		$("#alert").hide();
	}

        function setupAlertPanel(css){
		$("#alert").hide();
		$("#alert").show();
		$("#alert").empty();
		$("#alert").removeClass('alert-warning');
		$("#alert").removeClass('alert-success');
		$("#alert").toggleClass(css);
	}

        function showWarningMessage(txt){
		setupAlertPanel('alert-warning');
		//panelBody = '<button type="button" class="close" data-dismiss="alert" aria-label="Close"><span aria-hidden="true">&times;</span></button>';
		panelBody = '<button type="button" class="close" id="data-hide" onclick="closeAlertPanel()">&times;</button>'
		panelBody = panelBody + '<strong>Warning!</strong> '+txt;
		$("#alert").append(panelBody);
	}

        function showSuccessMessage(txt){
		setupAlertPanel('alert-success');
		//panelBody = '<button type="button" class="close" data-dismiss="alert" aria-label="Close"><span aria-hidden="true">&times;</span></button>';
		panelBody = '<button type="button" class="close" id="data-hide" onclick="closeAlertPanel()">&times;</button>'
		panelBody = panelBody + '<strong>Sucess!</strong> '+txt;
                $("#alert").append(panelBody);
	}
	
	function highlightTableFlow(exec) {
		tds = 'fullTrace'+exec
		$( "#tableFullTrace").find('td').each(function () {
				$(this).removeClass('td-marked')
				if ($(this).attr('id') == tds){
					$(this).toggleClass('td-marked')
				}
				});
		aTag = $("a[name='fullTraceExec"+ exec +"']");
		$('html,body').animate({scrollTop: aTag.offset().top},'slow');
	}

	function validateInputs(){
		if ('' == $( "#slo" ).val()){
			$( "#slo" ).val('20000');
		}
		if ('' == $( "#price" ).val()){
			$( "#price" ).val('7.0');
		}
		if ('' == $( "#instances" ).val()){
			$( "#instances" ).val('4');
		}

		if (('false' == $( 'input[name=category]:checked' ).val()) && ('Strict' == $( '#mode' ).val())){
			showWarningMessage('Mode capacity requires Category true');
			return false
		}
		return true
	}

	$("#button_execute").click( 
		function(){
			if (!validateInputs()){
				return false
			}
			
			data = '{"slo":' + $( "#slo" ).val()
			data = data + ', "price":'+$( "#price" ).val()
			data = data + ', "instances":'+$( "#instances" ).val()
			data = data + ', "mode":"'+$( "#mode" ).val()+'"'
			data = data + ', "heuristic":"'+$( '#heuristic' ).val()+'"'
			data = data + ', "category":'+$( 'input[name=category]:checked' ).val()
			data = data + ', "demand":['+$( '#demand' ).val()+']'
			data = data + ', "wkl":"'+$( "#workload_approach" ).val()+'"'
			data = data + ', "configuration":"'+$( "#configuration_approach" ).val()+'"}'

			showSuccessMessage('Waiting reply...')
			
			$.post( "/api/v1/capacitor/", data, 
				function( data ) {
					$( "#totalPrice" ).html(data.price)
					$( "#totalExecs" ).html(data.execs)
					$( "#panelExecPath" ).empty();
					var totalExec = 0
					jQuery.each( data.path, function( i, val ) {
						totalExec = totalExec+1
						var row = '<a href="javascript:highlightTableFlow('+totalExec+')" class="list-group-item"> #'+totalExec+" - "
						row = row + val.level+'<i class="fa fa-leaf fa-fw"></i> '
						row = row + '('+val.size+') '+val.name+' <i class="fa fa-cloud fa-fw"></i> '
						//row = row + val.size + ' <i class="fa fa-gears fa-fw"></i> '
						row = row + '<span class="pull-right text-muted small"><i class="fa fa-users fa-fw"></i><em>'+ val.workload +' users</em></span>'
						row = row + '</a>'
						$("#panelExecPath").append(row)
						});

					$( "#panelFullTrace" ).empty();
					var newTable = '<table class="table" id="tableFullTrace"> <thead> <tr> <th>Config</th>';
					jQuery.each($( '#demand' ).val() , function( i, val ) {
							newTable = newTable + '<th>'+val+'</th>';
					});
					newTable = newTable + '</tr></thead><tbody>';
					jQuery.each( data.spaceInfo, function( i, val ) {
						newTable = newTable + '<tr> <td class="deactive">('+val.size+') '+val.name+'</td>';
						jQuery.each( val.wkl, function( j, wklInfo) {
							newTable = newTable + '<td class="deactive" id="fullTrace'+wklInfo.when+'">'

							executed = 'fa-info-circle'
							if (wklInfo.exec){
								executed = 'fa-check-circle'
								newTable = newTable + '<a name="fullTraceExec'+ wklInfo.when +'"></a>'
							}
							
							right = '<i class="fa fa-thumbs-o-down fa-fw" style="color:red">'
							if (wklInfo.right){
								right = '<i class="fa fa-thumbs-o-up fa-fw" style="color:green">'
							}
							cadidate = 'red'
							if (wklInfo.cadidate){
								cadidate  = 'green'
							}

							newTable = newTable + '<i class="fa '+executed+' fa-fw" style="color:'+cadidate+'"></i>';


							newTable = newTable + wklInfo.when;

							newTable = newTable + right + '</td>'
						});
						newTable = newTable + '</tr>'
					});
					newTable = newTable + '</tbody></table>';
					$( "#panelFullTrace" ).append(newTable);
			      		showSuccessMessage('Done!')

				}
			      , "json" );
				//post end
		 }
	);
