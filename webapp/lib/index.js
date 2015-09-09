/* eslint no-unused-vars: 0 */
var Router = require('./router');
var algo = require('./utils/algo');
var main = require('./utils/main');
var pvd = require('./utils/pvd');
var query = require('./utils/query');

var React = require('react');

module.exports = {
  // this is the the whole app initter
  run: function () {
    var self = this;

    // init our URL handlers and the history tracker
    this.router = new Router();

    var page = React.render(React.createElement(main.App), document.body);
    var genesis = new WebSocket("ws://" + window.location.hostname + ":" + window.location.port + "/sock");
    genesis.onmessage = function (m) {
      page.setProps({graph: pvd.pvGraph(JSON.parse(m.data))});
    };

    // listen for new pages from the router
    self.router.on('newPage', function render(page) {
      React.render(React.createElement(page), document.body);
    });

    // we have what we need, we can now start our router and show the appropriate page
    self.router.history.start({pushState: true, root: '/'});
  },

  // This is how you navigate around the app.
  // this gets called by a global click handler that handles
  // all the <a> tags in the app.
  // it expects a url without a leading slash.
  // for example: "costello/settings".
  navigate: function (page) {
    var url = (page.charAt(0) === '/') ? page.slice(1) : page;
    this.router.history.navigate(url, {trigger: true});
  }
};

// run it
module.exports.run();
