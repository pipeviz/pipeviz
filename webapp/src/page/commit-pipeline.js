var index = require('../index');
var commitPipeline = require('../viz/commit-pipeline');

module.exports = {
  // this is the the whole app initter
  run: function () {
    index.openSocket();
    index.display(commitPipeline.App, document.querySelector('#main'));
  }
};

// run it
module.exports.run();
