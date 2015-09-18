var index = require('../index');
var data = require('../viz/data');

index.openSocket();
index.display(data.App, document.querySelector('#main'));
