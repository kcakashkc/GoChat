var ChatEngine=function(){
     var name="",
          msg="",
          fileName = "",
          TheFile = document.getElementById("TheFile"),
          msgid = document.getElementById("msg"),
          chatZone=null,
          fileForm = document.getElementById("fileForm"),
          sevr="",
          errPopup = $("#myPopup"),
          progressNumber = document.getElementById("progressNumber"),
          xhr="",
          ThreadList = document.getElementById("ThreadList");
          chatpage = $("#ChatMain"),
          serverHttp = "http://192.168.137.1:8080";
     //initialzation
     this.init=function(){
          if(EventSource){
               this.initSevr();
          }
          else{
               errPopup.html("Use latest Chrome or FireFox").popup("open", {transition : "pop"});
          }
     };
     //Setting user name
     this.setName=function(nam){
          /*name = prompt("Enter your name:","Chater");
          if (!name || name ==="") {
             name = "Chater";  
          }*/
          name = nam.replace(/(<([^>]+)>)/ig,"");
          // document.getElementById("uname").value = name;
     };
     this.getName = function(){
          return name;
     }
     var getPage = function(uname){
          chatpagename = "chat" + uname;
          c = document.getElementById(chatpagename);
          if (c !== null) {
               return c;
          }
          else{
               d=document.createElement('div');
               $(d).addClass("chats showChat")
                    .attr('id', chatpagename)
                    .html(/*Android.getChatHistory(name)*/'')
                    .appendTo(chatpage);
               return d;
          }
     };
     this.setmaxw=function(w){
          maxw = w*0.75;
     };
     //For sending message
     this.sendMsg=function(){ 
          msg=msgid.value;
          if (msg!="" && name!="") {
               getPage(name).innerHTML+='<div class="ChatDivRight"><span>'+msg+'</span></div>';
               // Android.saveChat(name, '<div class="ChatDivRight"><span>'+msg+'</span></div>');
               this.ajaxSent();
          }
     };
     //For sending file
     this.sendFile=function() {
          try {
               var file = TheFile.files[0];
               if (file) {
                    fileName = file.name;
               }
          } catch(err) {
               //nothing
          }
          this.uploadFile(file);
     };

     this.uploadFile=function(file) {
          try {
               var fd = new FormData();
               fd.append("uname", name);
               fd.append("file", file);
               /*$.post('/sendfile', fd, function(data, textStatus, xhr) {
                    msgid.value="";
                    // alert("reached");
               });*/
               xhr = new XMLHttpRequest();
               xhr.upload.addEventListener("progress", this.uploadProgress, false);
               xhr.addEventListener("loadend", this.fileSentComplete, false);
               xhr.addEventListener("load", this.uploadComplete, false);
               xhr.addEventListener("error", this.uploadFailed, false);
               xhr.addEventListener("abort", this.uploadCanceled, false);
               xhr.open("POST", serverHttp+"/sendfile");
               xhr.send(fd);
          } catch(err) {
               errPopup.html(err).popup("open", {transition : "pop"});
               fileForm.submit();
          }
     };

     this.uploadProgress=function(event) {
          if (event.lengthComputable) {
               var percentComplete = Math.round(event.loaded * 100 / event.total);
               progressNumber.innerHTML = percentComplete.toString() + '%';
          }
     };

     this.fileSentComplete=function(event) {
          errPopup.html('File upload completed for ' + fileName).popup("open", {transition : "pop"});
          progressNumber.innerHTML = "";
          TheFile = "";
     };

     this.uploadComplete=function(event) {
          progressNumber.innerHTML = 'Upload in the final stage for ' + fileName;
          TheFile = "";
     };

     this.uploadFailed=function(event) {
          errPopup.html('File upload failed for '+fileName).popup("open", {transition : "pop"});
          progressNumber.innerHTML = "";
     };

     this.uploadCanceled=function(event) {
          errPopup.html('File upload canceled for '+fileName).popup("open", {transition : "pop"});
          progressNumber.innerHTML = "";
     };
     //sending message to server
     this.ajaxSent=function(){
          // alert(name)
          $.post(serverHttp+'/sendmsg', {msg: msg, uname: name}, function(data, textStatus, xhr) {
               if(data == "1"){
                    msg.value="";
                    msgid.value="";
               }
               else{
                    errPopup.html("Failed to send the message").popup("open", {transition : "pop"});
               }
          });
          /*$.ajax({
               url: '/sendmsg',
               type: 'POST',
               data: {
                    msg: msg,
                    uname: name
               },
               success: function(data){
                    msg.value="";
                    alert("reached");
               },
               error: function( xhr, status, errorThrown ) {
                  alert( "Sorry, there was a problem!" );
                  console.log( "Error: " + errorThrown );
                  console.log( "Status: " + status );
                  console.dir( xhr );
              }
          });
          
          try{
               xhr=new XMLHttpRequest();
          }
          catch(err){
               alert(err);
          }
          params = 'msg='+msg+'&uname='+name;
          xhr.open('POST','/sendmsg',false);
          xhr.setRequestHeader("Content-length", params.length);
          xhr.onreadystatechange = function(){
               if(xhr.readyState == 4) {
                    if(xhr.status == 200) {
                         msg.value="";
                    }
               }     
          };
          xhr.send(params);*/
     };
     //HTML5 SSE(Server Sent Event) initilization
     this.initSevr=function(){
          sevr = new EventSource(serverHttp+'/chatlisten');
          sevr.onmessage = function(e){
               if(e.data!=""){
                    // getPage(obj.uname).innerHTML+=e.data;
               }
          };
          sevr.addEventListener("msg",function(e){
               var obj = JSON.parse(e.data);
               if (obj.msg != "") {
                    getPage(obj.uname).innerHTML+='<div class="ChatDivLeft"><span>'+obj.msg+'</span></div>';
                    if (document.getElementById("Thread_"+obj.uname) == null) {
                         ThreadList.innerHTML = '<li id="Thread_'+obj.uname+'"><a href="#Chatpage" data-transition="flip"><div class="usern">asd</div><span class="ui-li-count">0</span></a></li>' + ThreadList.innerHTML;
                    }
                    a = $("#Thread_"+obj.uname+" > a > span.ui-li-count");
                    a.html(a.html() + 1);
                    // Android.saveChat(obj.uname, '<div class="ChatDivLeft"><span>'+obj.msg+'</span></div>');
               }
          });
          sevr.addEventListener("file",function(e){
               var obj = JSON.parse(e.data);
               if (obj.filename != "" && obj.file != "") {
                    chatFile = '<div class="ChatDivLeft"><span><div style="position:absolute;"><a href="'+obj.filename+'" class="ChatFile">'+obj.file+'</a></div><div style="position:absolute;display:none;" class="FileProgress"></div></span></div>';
                    getPage(obj.uname).innerHTML += chatFile;
                    // Android.saveChat(obj.uname, chatFile);
               }
          });
          sevr.addEventListener("canvas",function(e){
               var obj = JSON.parse(e.data);
               chatCanvas = '<div class="ChatDivLeft"><span><canvas class="ChatCanvas" onload="DrawScreen.loadCanvas(this, "'+obj.img+'")" height="'+obj.nheight+'" width="'+obj.nwidth+'" style="background:'+obj.bck+';"></canvas></span></div>';
               getPage(obj.uname).innerHTML+=chatCanvas;
               // Android.saveChat(obj.uname, chatCanvas);
          });
          sevr.onerror = function(e) {
               errPopup.html("Failed to connect").popup("open", {transition : "pop"});
          };
     };

     this.close=function(){
          if(sevr!="")
               sevr.close();
          window.location = serverHttp+"/logout";
     };
};

var chat= new ChatEngine();
// chat.init();