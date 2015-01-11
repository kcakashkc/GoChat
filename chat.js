// function loadScript(url)
// {
//     // Adding the script tag to the head as suggested before
//     var head = document.getElementsByTagName('head')[0];
//     var script = document.createElement('script');
//     script.type = 'text/javascript';
//     script.src = url;

//     // Fire the loading
//     head.appendChild(script);
// }

// loadScript("jquery.js");

var ChatEngine=function(){
     var name=" ";
     var msg="";
     var chatZone=document.getElementById("chatZone");
     var sevr=" ";
     var xhr=" ";
     //initialzation
     this.init=function(){
          if(EventSource){
          // this.setName();
          this.initSevr(); 
          } else{
          alert("Use latest Chrome or FireFox");
        }
     };
     //Setting user name
     this.setName=function(nam){
          /*name = prompt("Enter your name:","Chater");
          if (!name || name ==="") {
             name = "Chater";  
          }*/
          name = nam.replace(/(<([^>]+)>)/ig,"");
          document.getElementById("uname").value = name;
     };
     //For sending message
     this.sendMsg=function(e){ 
          msg=document.getElementById("msg").value;
          if (msg!="") {
               chatZone.innerHTML+='<div class="chatmsg"><b>To '+name+'</b>: '+msg+'<br/></div>';
               this.ajaxSent();
               document.getElementById("msg").value="";
          }
     };
     //sending message to server
     this.ajaxSent=function(){
          // alert(name)
          $.post('/sendmsg', {msg: msg, uname: name}, function(data, textStatus, xhr) {
               msg.value="";
               // alert("reached");
          });/*
          $.ajax({
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
          });*/
          
          /*try{
               xhr=new XMLHttpRequest();
          }
          catch(err){
               alert(err);
          }
          params = 'msg='+msg+'&uname='+name;
          xhr.open('POST','/sendmsg',false);
          http.setRequestHeader("Content-length", params.length);
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
          sevr = new EventSource('/chatlisten');
          sevr.onmessage = function(e){ 
          if(e.data!=""){
               chatZone.innerHTML+=e.data;
          }
          };     
     };

     this.close=function(){
          if(sevr!="")
               sevr.close();
     };
};
// Createing Object for Chat Engine
var chat= new ChatEngine();
chat.init();