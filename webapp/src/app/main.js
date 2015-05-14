// just so my syntastic complains less
var React = React || {};

var data = new pvGraph(JSON.parse(document.getElementById("pipe-graph").innerHTML));

var App = React.createClass({
    dispayName: "pipeviz",
    getInitialState: function() {
        return {
            graph: {}
        };
    },
    render: function() {
        return React.createElement("div", {id: "pipeviz"});
    },
});

React.render(React.createElement(App, {graph: data}), document.body)
