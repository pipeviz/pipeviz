var index = require('../index');
var data = require('../viz/data');

module.exports = {
  // this is the the whole app initter
  run: function () {
    index.openSocket();
    index.display(data.App, document.querySelector('#main'));
  }
};

// run it
module.exports.run();
