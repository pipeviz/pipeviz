var React = require('react');
var D = React.DOM;

var Leaf = require('../utils/dataviewer/leaf');
var leaf = React.createFactory(Leaf);
var FilterBar = require('../utils/dataviewer/filter-bar');
var filterBar = React.createFactory(FilterBar);

var isEmpty = require('../utils/dataviewer/is-empty');
var lens = require('../utils/dataviewer/lens');
var noop = require('../utils/dataviewer/noop');

var pvd = require('../utils/pvd');

var Data = React.createClass({
  propTypes: {
    graph: React.PropTypes.object.isRequired,
    // For now it expects a factory function, not element.
    search: React.PropTypes.oneOfType([
      React.PropTypes.func,
      React.PropTypes.bool
    ]),
    onClick: React.PropTypes.func,
    validateQuery: React.PropTypes.func,
    isExpanded: React.PropTypes.func,
  },

  getDefaultProps: function() {
    return {
      className: '',
      id: 'pv-dataviewer',
      onClick: noop,
      validateQuery: function(query) {
        return query.length >= 2;
      },
      /**
       * Decide whether the leaf node at given `keypath` should be
       * expanded initially.
       * @param  {String} keypath
       * @param  {Any} value
       * @return {Boolean}
       */
      isExpanded: function(keypath, value) {
        return typeof value.initialExpand === 'function' ? value.initialExpand() : false;
      }
    };
  },
  getInitialState: function() {
    return {
      query: '',
      filter: ''
    };
  },
  render: function() {
    var p = this.props;
    var s = this.state;

    var data = s.filter ? p.graph.verticesWithType(s.filter) : p.graph.vertices();

    var rootNode = leaf({
      data: data,
      onClick: p.onClick,
      id: p.id,
      getOriginal: this.getOriginal,
      query: s.query,
      label: 'pv data',
      root: true,
      isExpanded: p.isExpanded,
      interactiveLabel: p.interactiveLabel,
      graph: p.graph,
      filter: s.filter
    });

    var notFound = D.div({ className: 'pv-dataviewer__not-found' }, 'Nothing found');

    return D.div({ className: 'pv-dataviewer ' + p.className },
                 this.renderToolbar(),
                 isEmpty(data) ? notFound : rootNode);
  },
  renderToolbar: function() {
    return D.div({ className: 'pv-dataviewer__toolbar' },
                 filterBar({
                   onChange: this.setFilter,
                   selected: this.state.filter,
                   options: this.props.graph.types(),
                 }));
  },
  search: function(query) {
    if (query === '' || this.props.validateQuery(query)) {
      this.setState({
        query: query
      });
    }
  },
  setFilter: function(type) {
    this.setState({
      filter: type,
    });
  },
  shouldComponentUpdate: function (p, s) {
    return s.query !== this.state.query ||
      s.filter !== this.state.filter ||
      p.graph.mid !== this.props.graph.mid ||
      p.onClick !== this.props.onClick;
  },
  getOriginal: function(path) {
    return lens(this.props.data, path);
  }
});

module.exports.App = React.createClass({
  displayName: "pipeviz",
  getDefaultProps: function () {
    return {
      graph: pvd.pvGraph({id: 0, vertices: []})
    };
  },
  render: function () {
    return React.createElement('section', {id: 'pipeviz'}, React.createElement(Data, {graph: this.props.graph}));
  }
});
