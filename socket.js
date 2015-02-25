io = function(host){
  if ("WebSocket" in window) {
    var socket = new WebSocket("ws://"+host);
    socket.actions = {};
    socket.pending = [];

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
        socket.send(JSON.stringify({
          namespace: "ms",
          event: event,
          args: args.slice(1)
        }));
      };

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