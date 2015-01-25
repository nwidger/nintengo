package http

var index = `
<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="utf-8">
    <meta http-equiv="X-UA-Compatible" content="IE=edge">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <title>nintengo - {{.NES.GameName}}</title>

    <!-- Latest compiled and minified CSS -->
    <link rel="stylesheet" href="https://maxcdn.bootstrapcdn.com/bootstrap/3.2.0/css/bootstrap.min.css">

    <!-- Optional theme -->
    <link href="//maxcdn.bootstrapcdn.com/bootswatch/3.2.0/darkly/bootstrap.min.css" rel="stylesheet">
    <!-- <link rel="stylesheet" href="https://maxcdn.bootstrapcdn.com/bootstrap/3.2.0/css/bootstrap-theme.min.css"> -->

    <style>
     body { padding-top: 70px; }
    </style>
  </head>
  <body>
    <div class='container'>
      <div class='row'>
	<nav class="navbar navbar-default navbar-fixed-top" role="navigation">
	  <div class="container-fluid">
	    <!-- Brand and toggle get grouped for better mobile display -->
	    <div class="navbar-header">
	      <button type="button" class="navbar-toggle collapsed" data-toggle="collapse" data-target="#bs-example-navbar-collapse-1">
		<span class="sr-only">Toggle navigation</span>
		<span class="icon-bar"></span>
		<span class="icon-bar"></span>
		<span class="icon-bar"></span>
	      </button>
	      <a class="navbar-brand" href="#">nintengo</a>
	    </div>

	    <!-- Collect the nav links, forms, and other content for toggling -->
	    <div class="collapse navbar-collapse" id="bs-example-navbar-collapse-1">
	      <ul class="nav navbar-nav">
	  	<li class='active'><a href='#'>{{.NES.GameName}}</a></li>
	  	<li><a href='#' id='pause-link'>Pause</a></li>
		<li><a href='#' id='toggle-stepping-link'>Toggle Stepping</a></li>
		<li><a href='#' id='save-state-link'>Save State</a></li>
		<li><a href='#' id='load-state-link'>Load State</a></li>
		<li><a href='#' id='reset-link'>Reset</a></li>
	      </ul>
	      <ul class="nav navbar-nav navbar-right">
		<li><a href='#' id='run-state'></a></li>
		<li><a href='#' id='step-state'></a></li>
	      </ul>
	    </div><!-- /.navbar-collapse -->
	  </div><!-- /.container-fluid -->
	</nav>

	<div class='col-md-6'>
	  <table class='table table-striped'>

	    <thead><tr><td><strong>CPU Variable</strong></td><td><strong>Value</strong></td></tr></thead>
	    <tbody>
	      <tr><td><kbd>A</kbd></td>   <td><code>{{printf "$%02x" .NES.CPU.M6502.Registers.A}}</code></td></tr>
	      <tr><td><kbd>X</kbd></td>   <td><code>{{printf "$%02x" .NES.CPU.M6502.Registers.X}}</code></td></tr>
	      <tr><td><kbd>Y</kbd></td>   <td><code>{{printf "$%02x" .NES.CPU.M6502.Registers.Y}}</code></td></tr>
	      <tr><td><kbd>P</kbd></td>   <td><code>{{printf "$%02x" .NES.CPU.M6502.Registers.P}}</code></td></tr>
	      <tr><td><kbd>SP</kbd></td>  <td><code>{{printf "$%04x" .NES.CPU.M6502.Registers.SP}}</code></td></tr>
	      <tr><td><kbd>PC</kbd></td>  <td><code>{{printf "$%04x" .NES.CPU.M6502.Registers.PC}}</code></td></tr>
	      <tr><td><kbd>NMI</kbd></td> <td><code>{{.NES.CPU.M6502.Nmi}}</code></td> </tr>
	      <tr><td><kbd>IRQ</kbd></td> <td><code>{{.NES.CPU.M6502.Irq}}</code></td> </tr>
	      <tr><td><kbd>RST</kbd></td> <td><code>{{.NES.CPU.M6502.Rst}}</code></td> </tr>
	    </tbody>

	    <thead><tr><td><strong>DMA Variable</strong></td><td><strong>Value</strong></td></tr></thead>
	    <tbody>
	      <tr><td><kbd>Pending</kbd></td> <td><code>{{printf "$%04x" .NES.CPU.DMA.Pending}}</code></td> </tr>
	    </tbody>

	    <thead><tr><td><strong>APU Variable</strong></td><td><strong>Value</strong></td></tr></thead>
	    <tbody>
	      <tr><td><kbd>Control</kbd></td>   <td><code>{{printf "$%02x" .NES.CPU.APU.Registers.Control}}</code></td></tr>
	      <tr><td><kbd>Status</kbd></td>    <td><code>{{printf "$%02x" .NES.CPU.APU.Registers.Status}}</code></td></tr>
	    </tbody>

	    <thead><tr><td><strong>PPU Variable</strong></td><td><strong>Value</strong></td></tr></thead>
	    <tbody>
	      <tr><td><kbd>Frame</kbd></td>    <td><code>{{.NES.PPU.Frame}}</code></td></tr>
	      <tr><td><kbd>Scanline</kbd></td> <td><code>{{.NES.PPU.Scanline}}</code></td></tr>
	      <tr><td><kbd>Cycle</kbd></td>    <td><code>{{.NES.PPU.Cycle}}</code></td></tr>

	      <tr><td><kbd>Controller</kbd></td> <td><code>{{printf "$%02x" .NES.PPU.Registers.Controller}}</code></td></tr>
	      <tr><td><kbd>Mask</kbd></td>       <td><code>{{printf "$%02x" .NES.PPU.Registers.Mask}}</code></td></tr>
	      <tr><td><kbd>Status</kbd></td>     <td><code>{{printf "$%02x" .NES.PPU.Registers.Status}}</code></td></tr>
	      <tr><td><kbd>OAMAddress</kbd></td> <td><code>{{printf "$%02x" .NES.PPU.Registers.OAMAddress}}</code></td></tr>
	      <tr><td><kbd>Scroll</kbd></td>     <td><code>{{printf "$%04x" .NES.PPU.Registers.Scroll}}</code></td></tr>
	      <tr><td><kbd>Address</kbd></td>    <td><code>{{printf "$%04x" .NES.PPU.Registers.Address}}</code></td></tr>
	      <tr><td><kbd>Data</kbd></td>       <td><code>{{printf "$%02x" .NES.PPU.Registers.Data}}</code></td></tr>

	      <tr><td><kbd>Latch</kbd></td>        <td><code>{{.NES.PPU.Latch}}</code></td></tr>
	      <tr><td><kbd>LatchAddress</kbd></td> <td><code>{{printf "$%04x" .NES.PPU.LatchAddress}}</code></td></tr>
	      <tr><td><kbd>LatchValue</kbd></td>   <td><code>{{printf "$%02x" .NES.PPU.LatchValue}}</code></td></tr>

	      <tr><td><kbd>AddressLine</kbd></td>    <td><code>{{printf "$%04x" .NES.PPU.AddressLine}}</code></td></tr>
	      <tr><td><kbd>PatternAddress</kbd></td> <td><code>{{printf "$%04x" .NES.PPU.PatternAddress}}</code></td></tr>

	      <tr><td><kbd>AttributeNext</kbd></td>  <td><code>{{printf "$%02x" .NES.PPU.AttributeNext}}</code></td></tr>
	      <tr><td><kbd>AttributeLatch</kbd></td> <td><code>{{printf "$%02x" .NES.PPU.AttributeLatch}}</code></td></tr>
	      <tr><td><kbd>Attributes</kbd></td>     <td><code>{{printf "$%04x" .NES.PPU.Attributes}}</code></td></tr>

	      <tr><td><kbd>TilesLow</kbd></td>       <td><code>{{printf "$%02x" .NES.PPU.TilesLow}}</code></td></tr>
	      <tr><td><kbd>TilesHigh</kbd></td>      <td><code>{{printf "$%02x" .NES.PPU.TilesHigh}}</code></td></tr>
	      <tr><td><kbd>TilesLatchLow</kbd></td>  <td><code>{{printf "$%02x" .NES.PPU.TilesLatchLow}}</code></td></tr>
	      <tr><td><kbd>TilesLatchHigh</kbd></td> <td><code>{{printf "$%02x" .NES.PPU.TilesLatchHigh}}</code></td></tr>

	    <thead><tr><td><strong>OAM Variable</strong></td><td><strong>Value</strong></td></tr></thead>
	    <tbody>
	      <tr><td><kbd>Address</kbd></td>  <td><code>{{printf "$%04x" .NES.PPU.OAM.Address}}</code></td></tr>
	      <tr><td><kbd>Latch</kbd></td>    <td><code>{{printf "$%02x" .NES.PPU.OAM.Latch}}</code></td></tr>
	      <tr><td><kbd>SpriteZeroInBuffer</kbd></td>    <td><code>{{.NES.PPU.OAM.SpriteZeroInBuffer}}</code></td></tr>
	    </tbody>

	  </table>

	  <h4>CPU Memory</h4>
	  <pre style='font-size: 11px' class='pre-scrollable'>{{.CPUMemory}}</pre>

	  <h4>PPU Memory</h4>
	  <pre style='font-size: 11px' class='pre-scrollable'>{{.PPUMemory}}</pre>

	  <h4>PPU Palette</h4>
	  <pre style='font-size: 11px' class='pre-scrollable'>{{.PPUPalette}}</pre>

	  <h4>OAM Memory</h4>
	  <pre style='font-size: 11px' class='pre-scrollable'>{{.OAMMemory}}</pre>

	  <h4>OAM Buffer Memory</h4>
	  <pre style='font-size: 11px' class='pre-scrollable'>{{.OAMBufferMemory}}</pre>

	</div>
	<div class='col-md-6'>

	  <table class='table table-striped'>
            {{range $i, $s := .NES.PPU.Sprites}}
	      <thead><tr><td><strong>PPU Sprite {{$i}} Variable</strong></td><td><strong>Value</strong></td></tr></thead>
	      <tbody>

        	<tr><td><kbd>TileLow </kbd></td>   <td><code>{{printf "$%0x2" $s.TileLow}}</code></td></tr>
        	<tr><td><kbd>TileHigh </kbd></td>  <td><code>{{printf "$%0x2" $s.TileHigh}}</code></td></tr>
        	<tr><td><kbd>Sprite </kbd></td>    <td><code>{{printf "$%08x" $s.Sprite}}</code></td></tr>
        	<tr><td><kbd>XPosition </kbd></td> <td><code>{{printf "$%02x" $s.XPosition}}</code></td></tr>
        	<tr><td><kbd>Address </kbd></td>   <td><code>{{printf "$%04x" $s.Address}}</code></td></tr>
        	<tr><td><kbd>Priority </kbd></td>  <td><code>{{printf "$%02x" $s.Priority}}</code></td></tr>
            {{end}}
	      </tbody>
	  </table>

	</div>
      </div>
      <div class='row'>

	<div class='col-md-8'>
	  <h4>PPU Pattern Tables</h4>
	</div>

	<div class='col-md-6'>
          <img alt='left'  style='width:100%;' src='data:image/png;base64,{{.PTLeft}}'  class='img-thumbnail img-responsive' />
	</div>

	<div class='col-md-6'>
          <img alt='right' style='width:100%;' src='data:image/png;base64,{{.PTRight}}' class='img-thumbnail img-responsive' />
	</div>

      </div>

      <div id='load-result' style='display: none'></div>
    </div>

    <!-- jQuery (necessary for Bootstrap's JavaScript plugins) -->
    <script src="https://ajax.googleapis.com/ajax/libs/jquery/1.11.1/jquery.min.js"></script>
    <!-- Latest compiled and minified JavaScript -->
    <script src="https://maxcdn.bootstrapcdn.com/bootstrap/3.2.0/js/bootstrap.min.js"></script>

    <script>
     $('#run-state').load('/run-state');
     $('#run-state').show();

     $('#step-state').show();
     $('#step-state').load('/step-state', function() {
       if ($('#step-state').text() == "NoStep") {
	 $('#pause-link').text('Pause');
       } else {
	 $('#pause-link').text('Step');
       }
     });

     $('#toggle-stepping-link').click(function(e) {
       e.preventDefault();
       $('#step-state').load('/toggle-step-state', function() {
	 if ($('#step-state').text() == "NoStep") {
	   $('#pause-link').text('Pause');
	 } else {
	   $('#pause-link').text('Step');
	 }
       });
     });

     $('#pause-link').click(function(e) {
       e.preventDefault();
       $('#run-state').load('/pause');

       if ($('#step-state').text() != "NoStep") {
	 location.reload();
       }
     });

     $('#save-state-link').click(function(e) {
       e.preventDefault();
       $('#load-result').load('/save-state');
     });

     $('#load-state-link').click(function(e) {
       e.preventDefault();
       $('#load-result').load('/load-state');
     });

     $('#reset-link').click(function(e) {
       e.preventDefault();
       $('#load-result').load('/reset');
     });

     $('code').click(function() {
       if ((match = $(this).text().match(/^\$(.*)$/)) != null) {
	 $(this).attr('orig-value', $(this).text());
	 $(this).text(parseInt(match[1], 16).toString(2).replace(/(....)/g, '$1 '));
       } else if ($(this).is('[orig-value]')) {
	 $(this).text($(this).attr('orig-value'));
       }
     });
    </script>
  </body>
</html>
`
