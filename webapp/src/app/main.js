var data = new pvGraph(JSON.parse(document.getElementById("pipe-graph").innerHTML));

var App = React.createClass({
    dispayName: "pipeviz",
    getInitialState: function() {
        return {
            graph: {}
        };
    },
});

React.render(React.createElement(App, {graph: data}), document.body)
