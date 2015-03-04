generateId = function(){
  var buf = new Uint8Array(32);
  window.crypto.getRandomValues(buf);
  return btoa(String.fromCharCode.apply(null, buf));
};

newSocket = function(ns, transport){
  var socket = {
    id: generateId(),
    namespace: ns,
    actions: [],
    pending: []
  };

  socket.on = function(event, fn){
    socket.actions[event] = fn;
  };

  socket.emitLater = function(){
    var obj = arguments;
    var args = Object.keys(obj).map(function (key) {return obj[key]});
    socket.pending.push(function(){
      socket.emit.apply(this, args);
    });
  };
  socket.emit = socket.emitLater;

  socket.emitNow = function(event){
    var obj = arguments;
    var args = Object.keys(obj).map(function (key) {return obj[key]});
    args = args.slice(1);
    transport.send(JSON.stringify({
      namespace: socket.namespace,
      socket: socket.id,
      event: event,
      args: args
    }));
  };

  socket.connect = function(){
    socket.emit = socket.emitNow;
    socket.emit("connection");
    while(socket.pending.length > 0) {
      socket.pending.shift()();
    }
  };

  socket.disconnect = function(){
    socket.emit = socket.emitLater;
    socket.handleEvent("disconnect", []);
  };

  socket.handleEvent = function(event, args){
    if (event in socket.actions) {
      socket.actions[event].apply(this, args);
    }
  };

  transport.sockets.push(socket);
  return socket;
};

newTransport = function(url){
  var transport = new WebSocket(url);
  transport.sockets = [];

  transport.onopen = function(){
    for (var i = 0; i < transport.sockets.length; i++){
      transport.sockets[i].connect();
    }
  };

  transport.onmessage = function(frame){
    var obj = JSON.parse(frame.data);
    for (var i = 0; i < transport.sockets.length; i++){
      if (obj.namespace === "") {
        obj.namespace = "/";
      }
      if (transport.sockets[i].namespace === obj.namespace) {
        transport.sockets[i].handleEvent(obj.event, obj.args);
      };
    }
  };

  transport.onclose = function() {
    for (var i = 0; i < transport.sockets.length; i++) {
      transport.sockets[i].disconnect();
    }
  };

  return transport;
};

var transports = {};
getTransport = function(url){
  if (url in transports) {
    return transports[url];
  } else {
    var transport = newTransport(url);
    transports[url] = transport;
    return transport;
  }
};

io = function(host){
  if ("WebSocket" in window) {
    var parts = host.split("/");
    var ns = "/"+parts.slice(1).join("/");
    var url = "ws://"+parts[0]+"/socket";

    var transport = getTransport(url);
    var socket = newSocket(ns, transport);
    
    return socket;
  } else {
    return {};
  }
};