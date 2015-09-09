/* eslint no-unused-vars: 0 */
var Router = require('./router');
var algo = require('./utils/algo');
var commits = require('./page/commits');
var pvd = require('./utils/pvd');
var query = require('./utils/query');

var React = require('react');

var mainSelector = '#main';

module.exports = {
  // this is the the whole app initter
  run: function () {
    this.initRouter();
    this.initNavigation();
  },

  // This is how you navigate around the app.
  // this gets called by a global click handler that handles
  // all the <a> tags in the app.
  // it expects a url without a leading slash.
  // for example: "costello/settings".
  navigate: function (page) {
    var url = (page.charAt(0) === '/') ? page.slice(1) : page;
    this.router.history.navigate(url, {trigger: true});
  },

  /**
   * One socket for everything we want out of the server.
   */
  socket: new WebSocket("ws://" + window.location.hostname + ":" + window.location.port + "/sock"),

  /**
   * Keeping the socket value handy in case we change pages.
   *
   * Will not work once we have proper delta updates from the server.
   */
  socketCache: false,

  initRouter: function () {
    // init our URL handlers and the history tracker
    this.router = new Router();

    // listen for new pages from the router
    this.router.on('newPage', function render(page) {
      var mod = module.exports;
      var display = React.render(React.createElement(page), document.querySelector(mainSelector));
      mod.socket.onmessage = function (m) {
        mod.socketCache = m;
        display.setProps({graph: pvd.pvGraph(JSON.parse(mod.socketCache.data))});
      };
      if (mod.socketCache) {
        mod.socket.onmessage.call(mod.socket, mod.socketCache);
      }
    });

    // we have what we need, we can now start our router and show the appropriate page
    this.router.history.start({pushState: true, root: '/'});
  },

  initNavigation: function () {
    var header = document.querySelector('#header');
    header.addEventListener('click', function (event) {
      if (event.target.matches('a')) {
        event.preventDefault();
        event.stopPropagation();
        module.exports.navigate(event.target.getAttribute('href'));
      }
    });
    var links = ['commits', 'data'].map(function (link) {
      return `<a href="${link}">${link}</a> `;
    }).forEach(function (link) {
      header.innerHTML += link;
    });
  }
};

// run it
module.exports.run();
