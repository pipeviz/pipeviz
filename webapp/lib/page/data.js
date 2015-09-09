var React = require('react');


var Data = React.createClass({
  displayName: 'pipeviz-data',
  render: function () {
    return React.createElement("aside", {id: "aside"}, React.createElement("div", {}, "Woot!"));
  }
});

module.exports.App = React.createClass({
  dispayName: "pipeviz",
  render: function () {
    return React.createElement('section', {id: 'pipeviz'}, React.createElement(Data));
  }
});
