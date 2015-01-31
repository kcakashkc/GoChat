var drawing = function(){
  var canvas,
      context,
      isbackground = false,
      backgroundColor = "#FFFFFF",
      color = "#FFCC66",
      tool = "marker",
      size = 6,
      clickX = new Array(),
      clickY = new Array(),
      clickChange = new Array(),
      clickDrag = new Array(),
      clickColor = new Array(),
      clickTool = new Array(),
      clickSize = new Array(),
      paint=false,
      selectedColor,
      editCanvas = null,
      viewCanvas = null,
      divpreview,
      serverHttp = "http://192.168.137.1:8080";

  var pushClick = function(x, y, dragging, colorchg, toolchg, sizechg){
    c = 0;
    clickX.push(x);
    clickY.push(y);
    if (dragging!=clickDrag.top) {
      clickDrag.push(dragging);
      c += 1;
    };
    if (colorchg) {
      clickColor.push(color);
      c += 2;
    };
    if (toolchg) {
      clickTool.push(tool);
      c += 4;
    };
    if (sizechg) {
      clickSize.push(size);
      c += 8;
    };
    clickChange.push(c);
  };

  var draw = function(){
    context.strokeStyle = color;
    context.lineJoin = "round";
    context.lineWidth = size;
        i = clickX.length-1;
    // for(var i=0; i < clickX.length; i++) {    
      context.beginPath();
      if(clickDrag[i] && i){
        context.moveTo(clickX[i-1], clickY[i-1]);
       }else{
         context.moveTo(clickX[i]-1, clickY[i]);
       }
       context.lineTo(clickX[i], clickY[i]);
       context.closePath();
       context.stroke();
    // }
  };


  var createEventListeners = function(){
    var pressed = function(e){
      Coffset = $(this).offset();
      var mouseX = e.pageX - Coffset.left;
      var mouseY = e.pageY - Coffset.top;
        
      paint = true;
      pushClick(mouseX, mouseY, false,(color!=clickColor.top),(tool!=clickTool.top),size!=clickSize.top);
      draw();
    },
    dragging = function(e){
      if(paint){
        Coffset = $(this).offset();
        pushClick(e.pageX - Coffset.left, e.pageY - Coffset.top, true,false,false,false);
        draw();
      }
      e.preventDefault();
    },
    released = function(e){
      paint = false;
    },
    increaseSize = function(){
      size+=2;
    },
    decreaseSize = function(){
      if(size>2)
        size-=2;
    },
    setBackground = function(){
      isbackground = true;
      selectedColor.style.visibility="hidden";
    },
    clearCanvas = function(){
      context.clearRect(0, 0, context.canvas.width, context.canvas.height);
    },
    uploadProgress = function(ob){
      if (event.lengthComputable) {
          var percentComplete = Math.round(event.loaded * 100 / event.total);
          ob.children('.FileProgress').html(percentComplete.toString() + '%');
      }
    },
    fileSentComplete = function(ob){
      ob.children('.FileProgress').html("Complete").css('display', 'none');
      ob.attr('href', '');
    },
    uploadComplete = function(ob){
      ob.children('.FileProgress').html("Building");
      ob.children('.FileProgress').addClass('FileProgressClose');
    },
    uploadFailed = function(ob){
      ob.children('.FileProgress').html("Failed");
      ob.children('.FileProgress').addClass('FileProgressClose');
    },
    uploadCanceled = function(ob){
      ob.children('.FileProgress').html("Canceled");
      ob.children('.FileProgress').addClass('FileProgressClose');
    },
    editCanvasFunc = function(){
      editCanvas.css('display', 'none');
      viewCanvas.css('display', 'block');
      canvas.setAttribute("id","canvas");
    },
    viewCanvasFunc = function(){
      editCanvas.css('display', 'block');
      viewCanvas.css('display', 'none');
      canvas.removeAttribute("id");
    };

      $( "#DrawSize" ).on( 'slidestop', function( event ) {
        size = $(this).val()*2;
        alert(size);
      });
      $( document ).on( "vmousedown", "#canvas", pressed);
      $( document ).on( "vmousemove", "#canvas", dragging);
      $( document ).on( "vmouseup", "#canvas", released);
      $( document ).on( "vmouseout", "#canvas", released);

    document.getElementById("setBackground").addEventListener("click",setBackground);
    document.getElementById("clearCanvas").addEventListener("click",clearCanvas);
    editCanvas.on("click",editCanvasFunc);
    viewCanvas.on("click",viewCanvasFunc);
    // document.getElementById("undoCanvas").addEventListener("click",clearCanvas);
    // document.getElementById("redoCanvas").addEventListener("click",clearCanvas);
    $(".FileProgressClose").on('click', function(event) {
      event.preventDefault();
      $(this).css('display', 'none');
    });
    $(".ChatFile").on('click', function(event) {
      event.preventDefault();
      try {
           xhr = new XMLHttpRequest();
           xhr.upload.addEventListener("progress", function(){uploadProgress($(this));}, false);
           xhr.addEventListener("loadend", function(){fileSentComplete($(this));}, false);
           xhr.addEventListener("load", function(){uploadComplete($(this));}, false);
           xhr.addEventListener("error", function(){uploadFailed($(this));}, false);
           xhr.addEventListener("abort", function(){uploadCanceled($(this));}, false);
           xhr.open("GET", $(this).attr('href'));
           xhr.send(fd);
           $(this).children('.FileProgress').css('display', 'block');
      } catch(err) {
           alert(err);
           fileForm.submit();
      }
    });
  };

  this.mouseOverColor = function(hex){
    divpreview.style.backgroundColor=hex;
    document.body.style.cursor="pointer";
  };

  this.mouseOutMap = function(){
    divpreview.style.backgroundColor=color;
    document.body.style.cursor="";
  };

  this.clickColor = function(colorhex,seltop,selleft){
    if (seltop>-1 && selleft>-1){
      selectedColor.style.top=seltop + "px";
      selectedColor.style.left=selleft + "px";
      selectedColor.style.visibility="visible";
    }
    else{
      divpreview.style.backgroundColor=colorhex;
      selectedColor.style.visibility="hidden";
    }
    if (isbackground) {
      backgroundColor = colorhex;
      canvas.style.background=backgroundColor;
      isbackground = false;
      selectedColor.style.visibility="visible";
    }
    else{
      color = colorhex;
    }
  };

  this.loadCanvas = function (new_canvas, durl) {
    var img = new Image, c = new_canvas;
    img.load(function(){
      c.height = c.height * factor;
      c.width = c.width * factor;
      c.getContext("2d").drawImage(this,0,0,c.width,c.height).scale (factor, factor);
    }).attr("src", durl);
  };

  this.send = function(name){
    var dataURL = canvas.toDataURL();
    $.post(serverHttp+'/sendcanvas', {bck:backgroundColor, uname: name, img: dataURL, height: canvas.height, width: canvas.width}, function(data, textStatus, xhr) {

    });
  }

  this.resizeScreen = function(){
    var screenh = $.mobile.getScreenHeight(),
        header = $("#DrawHeader").outerHeight() - 1,
        footer = $("#DrawFooter").outerHeight() - 1,
        contentCurrent = $("#DrawMain").outerHeight() - $("#DrawMain").height(),
        content = screenh - header - footer - contentCurrent;
    $("#DrawMain").height(content);
    canvas.height = content - 5;
    canvas.width = $("#DrawMain").innerWidth() - 32;
  };

  this.init = function(){
    canvas = document.getElementById('canvas');
    context = canvas.getContext("2d");
    selectedColor = document.getElementById("selectedColor");
    divpreview = document.getElementById("divpreview");
    editCanvas = $("#editCanvas");
    viewCanvas = $("#viewCanvas");
    clickColor.push(color);
    clickTool.push(tool);
    clickSize.push(size);
    createEventListeners();
    this.resizeScreen();
    factor = 1;
    if ((window.devicePixelRatio> 1) && ((context.webkitBackingStorePixelRatio <2) || (context.webkitBackingStorePixelRatio == undefined))) {
      factor = 2; 
    }    
  };

};

var DrawScreen = new drawing();