var index = require('../index');
var commitPipeline = require('../viz/commit-pipeline');

index.openSocket();
index.display(commitPipeline.App, document.querySelector('#main'));
