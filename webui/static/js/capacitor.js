	$( "#heuristic" ).change(function() {
		if ('policy' == $(this).val()){
			$( "#workload_approach" ).removeAttr('disabled');
                        $( "#configuration_approach" ).removeAttr('disabled');
		}else{
			$( "#workload_approach" ).attr('disabled', 'disabled');
			$( "#configuration_approach" ).attr('disabled', 'disabled');
		}
		if ('e' == $(this).val()){
			$( "#maxExecs" ).removeAttr('disabled');
		}else{
			$( "#maxExecs" ).attr('disabled', 'disabled');
		}
	});

        function googleChart(data){
		var chartData = new google.visualization.DataTable();
		chartData.addColumn('string', 'Configuration');
		chartData.addColumn('number', 'Predictions');
		jQuery.each( data.execsByKey, function( i, val ) {
			//console.log(val)
			chartData.addRow([val.key, val.execs]);
			});
		var options = {
			//'title':'Predictions by Configuration',
			width:"90%",
			height:250,
			legend: { position: "none" }

		};

		//Instantiate and draw our chart, passing in some options.
		var chart = new google.charts.Bar(document.getElementById('morris-bar-chart'));
		chart.draw(chartData, options);
	}

	function morrisChart(data){
		new Morris.Bar({
		    element: 'morris-bar-chart',
		    data: data.execsByKey,
		    xkey: 'key',
		    ykeys: ['execs'],
		    labels: ['Predictions'],
		    //xLabelAngle: 75,
		    hideHover: 'auto',
		    resize: true
		});
	}

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

	function generateDSpace(){
		showSuccessMessage('Generating graph...')
			$.post( "/api/v1/capacitor/dot", $( "#dspaceParam" ).val(),
					function( data ) {
						if(data == 'ERROR'){
							showWarningMessage('Error generating graph. The service is probably down');
							panel = '<div class=\"alert alert-info alert-dismissible\" role=\"alert\"><button type=\"button\" class=\"close\" data-dismiss=\"alert\" aria-label=\"Close\"><span aria-hidden=\"true\">&times;</span></button><pre>';
							panel = panel + data
				panel = panel + '</pre></div>'
				$("#mainAlertPanel").append(panel);
						}
						else{
							showSuccessMessage('Done!')
							var svg = Viz(data, "svg");
							panel = '<div class=\"alert alert-info alert-dismissible\" role=\"alert\"><button type=\"button\" class=\"close\" data-dismiss=\"alert\" aria-label=\"Close\"><span aria-hidden=\"true\">&times;</span></button><hr>'+svg;
							$("#mainAlertPanel").append(panel);
						}
					}
					, "text" );
	}

	function downloadDeploymentSpaceNew(){
		$.ajax({
			type:"POST",
		url: "http://graphviz-dev.appspot.com/create_preview",
		data: {'engine':"dot", 'script': $( "#dspaceParam" ).val()},
		timeout: 33000,
		beforeSend: function(){
		},
		success: function(data,status){
			if(data == 'ERROR'){
				panel = '<div class=\"alert alert-info alert-dismissible\" role=\"alert\"><button type=\"button\" class=\"close\" data-dismiss=\"alert\" aria-label=\"Close\"><span aria-hidden=\"true\">&times;</span></button><pre>';
                                panel = panel + data
                                panel = panel + '</pre></div>'
			}
			else{
				panel = '<div class=\"alert alert-info alert-dismissible\" role=\"alert\"><button type=\"button\" class=\"close\" data-dismiss=\"alert\" aria-label=\"Close\"><span aria-hidden=\"true\">&times;</span></button><img src=\'http://graphviz-dev.appspot.com/get_preview?id='+data;
                                panel = panel + '\'/></div>'
			}
			$("#mainAlertPanel").append(panel);
		},
		error: function(req,status){
			console.log(req)
		},
		complete: function(req,status){
		},
		});
	}

	function downloadDeploymentSpace(){
		showSuccessMessage('Generating graph...')
		$.post( "/api/v1/capacitor/draw", $( "#dspaceParam" ).val(),
			function( data ) {
				if(data == 'ERROR'){
					showWarningMessage('Error generating graph. The service is probably down');
					panel = '<div class=\"alert alert-info alert-dismissible\" role=\"alert\"><button type=\"button\" class=\"close\" data-dismiss=\"alert\" aria-label=\"Close\"><span aria-hidden=\"true\">&times;</span></button><pre>';
					panel = panel + data
					panel = panel + '</pre></div>'
					$("#mainAlertPanel").append(panel);
				}
				else{
					showSuccessMessage('Done!')
					var a = $("<a>").attr("href", 'http://graphviz-dev.appspot.com/get_preview?id='+data).attr("download", 'http://graphviz-dev.appspot.com/get_preview?id='+data).appendTo("body");
		                        a[0].click();
					a.remove();
				}
			}
		, "text" );
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
			data = data + ', "app":"'+$( '#app' ).val()+'"'
			data = data + ', "category":'+$( 'input[name=category]:checked' ).val()
		        data = data + ', "isCapacityFirst":'+$( 'input[name=isCapacityFirst]:checked' ).val()
			data = data + ', "useML":'+$( 'input[name=useML]:checked' ).val()
			data = data + ', "demand":['+$( '#demand' ).val()+']'
			if($( '#vmtype' ).val() != null){
				data = data + ', "vmtype":['+$( '#vmtype' ).val()+']'
			}else{
				data = data + ', "vmtype":[]'
			}
		        if ($( "#maxExecs" ).val() != ''){
				data = data + ', "maxExecs":'+$( '#maxExecs' ).val()
			}
			data = data + ', "wkl":"'+$( "#workload_approach" ).val()+'"'
			data = data + ', "configuration":"'+$( "#configuration_approach" ).val()+'"'
			data = data + ', "equiBehavior":'+$( "#equivalents_behavior" ).val()+'}'

			showSuccessMessage('Waiting reply...')
			$( "#dspaceParam" ).val(data);
			
			$.post( "/api/v1/capacitor/", data, 
				function( data ) {
					$( "#totalPrice" ).html(data.price);
					$( "#totalExecs" ).html(data.execs);
				        $( "#fMeasure" ).html(data.fmeasure);
					$( "#success" ).html(data.success);
					$( "#panelExecPath" ).empty();
					$( "#morris-bar-chart" ).empty();

					morrisChart(data);
					//googleChart(data);

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
							td_class = 'deactive'
							if (!wklInfo.right){
								td_class = 'td-error'
							}

							newTable = newTable + '<td class="'+td_class+'" id="fullTrace'+wklInfo.when+'">'

							executed = 'fa-info-circle'
							if (wklInfo.exec){
								executed = 'fa-check-circle'
								newTable = newTable + '<a name="fullTraceExec'+ wklInfo.when +'"></a>'
							}
							
							cadidate = 'red'
							if (wklInfo.cadidate){
								cadidate  = 'green'
							}

							newTable = newTable + '<i class="fa '+executed+' fa-fw" style="color:'+cadidate+'"></i>';


							newTable = newTable + wklInfo.when;

							newTable = newTable + '</td>'
						});
						newTable = newTable + '</tr>'
					});
					newTable = newTable + '</tbody></table>';
					$( "#panelFullTrace" ).append(newTable);
			      		showSuccessMessage('Done! <a href="javascript:generateDSpace()" style="color:green"> View DSpace</a>')

				}
			      , "json" );
				//post end
		 }
	);
	//google.load("visualization", "1.1", {packages:["bar"]});
