io = function(host){
  if ("WebSocket" in window) {
    var parts = host.split("/");
    var ns = "/"+parts.slice(1).join("/");
    var url = "ws://"+parts[0]+"/socket";
    var socket = new WebSocket(url);
    socket.actions = {};
    socket.pending = [];

    // extend this for sub-namespaces
    socket.namespace = ns;
    var buf = new Uint8Array(32);
    window.crypto.getRandomValues(buf);
    socket.id = btoa(String.fromCharCode.apply(null, buf));
    console.log(socket.id);

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

    socket.onopen = function(){
      socket.emit = function(event){
        var obj = arguments;
        var args = Object.keys(obj).map(function (key) {return obj[key]});
        args = args.slice(1);
        socket.send(JSON.stringify({
          namespace: socket.namespace,
          socket: socket.id,
          event: event,
          args: args
        }));
      };

      socket.emit("connection");

      while(socket.pending.length > 0) {
        socket.pending.shift()();
      }
    };

    socket.onmessage = function(frame){
      var obj = JSON.parse(frame.data);
      if (obj.event in socket.actions) {
        socket.actions[obj.event].apply(this, obj.args);
      }
    };

    socket.onclose = function() {
      socket.emit = socket.emitLater;
      if ("disconnect" in socket.actions) {
        socket.actions["disconnect"]();
      }
    };
    
    return socket;
  } else {
    return {};
  }
};