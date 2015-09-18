var React = require('react');
var pvd = require('./utils/pvd');
var config = require('./config');

module.exports = {

  /**
   * One socket for everything we want out of the server.
   */
  socket: null,

  /**
   * Keeping the socket value handy in case we change pages.
   *
   * Will not work once we have proper delta updates from the server.
   */
  socketCache: false,

  /**
   * Open socket.
   */
  openSocket: function () {
    this.socket = new WebSocket("ws://" + config.server + config.path);
    return this.socket;
  },

  display: function (page, container) {
    var mod = module.exports;
    var display = React.render(React.createElement(page), container);
    mod.socket.onmessage = function (m) {
      mod.socketCache = m;
      display.setProps({graph: pvd.pvGraph(JSON.parse(mod.socketCache.data))});
    };
    if (mod.socketCache) {
      mod.socket.onmessage.call(mod.socket, mod.socketCache);
    }
    return display;
  }
};
