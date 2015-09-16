require=(function e(t,n,r){function s(o,u){if(!n[o]){if(!t[o]){var a=typeof require=="function"&&require;if(!u&&a)return a(o,!0);if(i)return i(o,!0);var f=new Error("Cannot find module '"+o+"'");throw f.code="MODULE_NOT_FOUND",f}var l=n[o]={exports:{}};t[o][0].call(l.exports,function(e){var n=t[o][1][e];return s(n?n:e)},l,l.exports,e,t,n,r)}return n[o].exports}var i=typeof require=="function"&&require;for(var o=0;o<r.length;o++)s(r[o]);return s})({181:[function(require,module,exports){
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

},{"../index":179,"../viz/data":194}],194:[function(require,module,exports){
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

},{"../utils/dataviewer/filter-bar":183,"../utils/dataviewer/is-empty":185,"../utils/dataviewer/leaf":186,"../utils/dataviewer/lens":187,"../utils/dataviewer/noop":188,"../utils/pvd":191,"react":178}],187:[function(require,module,exports){
var type = require('./type');

var PATH_DELIMITER = '.';

function lens(data, path) {
    var p = path.split(PATH_DELIMITER);
    var segment = p.shift();

    if (!segment) {
        return data;
    }

    var t = type(data);

    if (t === 'Array' && data[integer(segment)]) {
        return lens(data[integer(segment)], p.join(PATH_DELIMITER));
    } else if (t === 'Object' && data[segment]) {
        return lens(data[segment], p.join(PATH_DELIMITER));
    }
}

function integer(string) {
    return parseInt(string, 10);
}

module.exports = lens;

},{"./type":189}],186:[function(require,module,exports){
var React = require('react');
var D = React.DOM;

var uid = require('./uid');
var type = require('./type');

var Highlighter = require('./highlighter');
var highlighter = React.createFactory(Highlighter);

var pvd = require('../pvd.js');
var query = require('../query.js');
var _ = require('lodash');

var PATH_PREFIX = '.root.';

var Leaf = React.createClass({
  getInitialState: function() {
    return {
      expanded: this._isInitiallyExpanded(this.props)
    };
  },
  getDefaultProps: function() {
    return {
      root: false,
      filter: '',
      prefix: ''
    };
  },
  render: function() {
      var id = 'id_' + uid();
      var p = this.props;

      var d = {
        path: this.keypath(),
        key: p.label.toString(),
        value: p.data
      };

      var onLabelClick = this._onClick.bind(this, d);

      return D.div({ className: this.getClassName(), id: 'leaf-' + this._rootPath() },
        D.input({ className: 'pv-dataviewer__radio', type: 'radio', name: p.id, id: id, tabIndex: -1 }),
        D.label({ className: 'pv-dataviewer__line', htmlFor: id, onClick: onLabelClick },
          D.div({ className: 'pv-dataviewer__flatpath' }, d.path),
          D.span({ className: 'pv-dataviewer__key' },
            this.format(d.key),
            ':',
            this.renderInteractiveLabel(d.key, true)),
          this.renderTitle(),
          this.renderShowOriginalButton()),
        this.renderChildren());
  },
  renderTitle: function() {
    var data = this.data();
    var t = type(data);

    // special case for root
    if (this.props.root) {
      return D.span({ className: 'pv-dataviewer__value pv-dataviewer__value_helper' },
                    _.filter([data.length, this.props.filter, "vertices", '(at message ' + this.props.graph.mid + ')']).join(' ')
                   );
    }

    switch (t) {
      case 'Array':
        return D.span({ className: 'pv-dataviewer__value pv-dataviewer__value_helper' },
                      '[] ' + items(data.length));
        case 'Object':
          if (pvd.isVertex(data) || pvd.isEdge(data)) {
            return D.span({ className: 'pv-dataviewer__value pv-dataviewer__value_helper' },
                          data.Typ() + (query.objectLabel(data) ? ' ' + query.objectLabel(data) : ''));
          } else if (_.has(data, 'isPropObj')) {
            // Condense property objects into a single "value"
            return D.span({ className: 'pv-dataviewer__value pv-dataviewer__propvalue pv-dataviewer__value_' + type(data.value).toLowerCase() },
                          data.value, D.span({ className: 'pv-dataviewer__msgsrc' }, "  (from msg: " + data.msgsrc + ")"),
                          this.renderInteractiveLabel(data.value, false));
          }

          return D.span({ className: 'pv-dataviewer__value pv-dataviewer__value_helper' },
                        '{} ' + items(Object.keys(data).length));
        default:
          return D.span({ className: 'pv-dataviewer__value pv-dataviewer__value_' + t.toLowerCase() },
                        this.format(String(data)),
                        this.renderInteractiveLabel(data, false));
    }
  },
  renderChildren: function() {
    var p = this.props;
    var childPrefix = this._rootPath();
    var data = this.data();
    var shared = {
      prefix: childPrefix,
      onClick: p.onClick,
      id: p.id,
      query: p.query,
      getOriginal: this.state.original ? null : p.getOriginal,
      isExpanded: p.isExpanded,
      interactiveLabel: p.interactiveLabel,
      graph: p.graph
    };

    if (this.state.expanded && !isPrimitive(data) && !_.has(data, 'isPropObj')) {
      if (pvd.isVertex(data) || pvd.isEdge(data)) {
        var isv = pvd.isVertex(data);
        // id up front
        var children = [leaf(_.assign({
          data: data.id,
          label: "id",
          key: getLeafKey("id", data.id)
        }, shared))];

        // etype/vtype next
        children.push(leaf(_.assign({
          data: data.Typ(),
          label: isv ? "vtype" : "etype",
          key: getLeafKey(isv ? "vtype" : "etype", data.Typ())
        }, shared)));

        if (!isv) {
          // stick the source and target in now for edges
          children.push(leaf(_.assign({
            data: p.graph.get(data.source),
            label: "source",
            key: getLeafKey("source", {})
          }, shared)));

          children.push(leaf(_.assign({
            data: p.graph.get(data.target),
            label: "target",
            key: getLeafKey("target", {})
          }, shared)));
        }

        // now, props
        children.push(leaf(_.assign({
          data: _.mapValues(pvd.isVertex(data) ? data.vertex.properties : data.properties, function(v) {
            return _.assign({isPropObj: true}, v);
          }),
          label: "properties",
          key: getLeafKey("properties", {}) // just cheat
        }, shared, { isExpanded: function() { return true; }}))); // always expand from prop level downwards

        // finally, if it's a vertex, add the edges
        if (isv) {
          children.push(leaf(_.assign({
            data: _.assign(Object.create(expander),
                           _.zipObject(_.map(data.outEdges, function(eid) {
                             return [eid, p.graph.get(eid)];
                           }))),
            label: "outEdges",
            key: getLeafKey("outEdges", {}) // just cheat
          }, shared)));

          children.push(leaf(_.assign({
            data: _.assign(Object.create(expander),
                           _.zipObject(_.map(data.inEdges, function(eid) {
                             return [eid, p.graph.get(eid)];
                           }))),
            label: "inEdges",
            key: getLeafKey("inEdges", {}) // just cheat
          }, shared)));
        }

        return children;
      } else {
        return Object.keys(data).map(function(key) {
          var value = data[key];

          // hardcode: use the vertex id instead of array position
          if (p.root) {
            key = data[key].id;
          }

          return leaf({
            data: value,
            label: key,
            prefix: childPrefix,
            onClick: p.onClick,
            id: p.id,
            query: p.query,
            getOriginal: this.state.original ? null : p.getOriginal,
            key: getLeafKey(key, value),
            isExpanded: p.isExpanded,
            interactiveLabel: p.interactiveLabel,
            graph: p.graph
          });
        }, this);
      }
    }

    return null;
  },
  renderShowOriginalButton: function() {
    var p = this.props;

    if (isPrimitive(p.data) || _.has(p.data, 'isPropObj') || this.state.original || !p.getOriginal || !p.query || contains(this.keypath(), p.query)) {
      return null;
    }

    return D.span({
      className: 'pv-dataviewer__show-original',
      onClick: this._onShowOriginalClick
    });
  },
  renderInteractiveLabel: function(originalValue, isKey) {
    if (typeof this.props.interactiveLabel === 'function') {
      return this.props.interactiveLabel({
        // The distinction between `value` and `originalValue` is
        // provided to have backwards compatibility.
        value: String(originalValue),
        originalValue: originalValue,
        isKey: isKey,
        keypath: this.keypath()
      });
    }

    return null;
  },
  componentWillReceiveProps: function(p) {
    if (p.query) {
      this.setState({
        expanded: !contains(p.label, p.query)
      });
    }

    // Restore original expansion state when switching from search mode
    // to full browse mode.
    if (this.props.query && !p.query) {
      this.setState({
        expanded: this._isInitiallyExpanded(p)
      });
    }
  },
  _rootPath: function() {
    return this.props.prefix + '.' + this.props.label;
  },
  keypath: function() {
    return this._rootPath().substr(PATH_PREFIX.length);
  },
  data: function() {
    return this.state.original || this.props.data;
    //return this.state.original || (_.has(this.props.data, 'isPropObj') ? this.props.data.value : this.props.data);
  },
  format: function(string) {
    return highlighter({
      string: string,
      highlight: this.props.query
    });
  },
  getClassName: function() {
    var cn = 'pv-dataviewer__leaf';

    if (this.props.root) {
      cn += ' pv-dataviewer__leaf_root';
    }

    if (this.state.expanded) {
      cn += ' pv-dataviewer__leaf_expanded';
    }

    if (!isPrimitive(this.props.data)) {
      cn += ' pv-dataviewer__leaf_composite';
    }

    if (pvd.isVertex(this.props.data)) {
      cn += ' pv-dataviewer__vertex_leaf';
    }

    if (pvd.isEdge(this.props.data)) {
      cn += ' pv-dataviewer__edge_leaf';
    }

    return cn;
  },
  toggle: function() {
    this.setState({
      expanded: !this.state.expanded
    });
  },
  _onClick: function(data, e) {
    this.toggle();
    this.props.onClick(data);

    e.stopPropagation();
  },
  _onShowOriginalClick: function(e) {
    this.setState({
      original: this.props.getOriginal(this.keypath())
    });

    e.stopPropagation();
  },
  _isInitiallyExpanded: function(p) {
    var keypath = this.keypath();

    if (p.root) {
      return true;
    }

    if (p.query === '') {
      return p.isExpanded(keypath, p.data);
    } else {
      // When a search query is specified, first check if the keypath
      // contains the search query: if it does, then the current leaf
      // is itself a search result and there is no need to expand further.
      //
      // Having a `getOriginal` function passed signalizes that current
      // leaf only displays a subset of data, thus should be rendered
      // expanded to reveal the children that is being searched for.
      return !contains(keypath, p.query) && (typeof p.getOriginal === 'function');
    }
  }
});

var expander = {
  initialExpand: function() { return true; }
};

// FIXME: There should be a better way to call a component factory from inside
// component definition.
var leaf = React.createFactory(Leaf);

function items(count) {
    return count + (count === 1 ? ' item' : ' items');
}

function getLeafKey(key, value) {
    if (isPrimitive(value)) {
        return key + ':' + value;
    } else {
        return key + '[' + type(value) + ']';
    }
}

function contains(string, substring) {
    return string.indexOf(substring) !== -1;
}

function isPrimitive(value) {
    var t = type(value);
    //return (t !== 'Object' || _.has(value, 'isPropObj')) && t !== 'Array';
    return t !== 'Object' && t !== 'Array';
}

module.exports = Leaf;

},{"../pvd.js":191,"../query.js":192,"./highlighter":184,"./type":189,"./uid":190,"lodash":23,"react":178}],190:[function(require,module,exports){
var id = Math.ceil(Math.random() * 10);

module.exports = function() {
    return ++id;
};

},{}],189:[function(require,module,exports){
module.exports = function(value) {
    return Object.prototype.toString.call(value).slice(8, -1);
};

},{}],184:[function(require,module,exports){
var React = require('react');
var span = React.DOM.span;

module.exports = React.createClass({
    getDefaultProps: function() {
        return {
            string: '',
            highlight: ''
        };
    },
    shouldComponentUpdate: function(p) {
        return p.highlight !== this.props.highlight;
    },
    render: function() {
        var p = this.props;

        if (!p.highlight || p.string.indexOf(p.highlight) === -1) {
            return span(null, p.string);
        }

        return span(null,
            p.string.split(p.highlight).map(function(part, index) {
                return span({ key: index },
                    index > 0 ?
                        span({ className: 'pv-dataviewer__hl' }, p.highlight) :
                        null,
                    part);
            }));
    }
});

},{"react":178}],185:[function(require,module,exports){
module.exports = function(object) {
    return Object.keys(object).length === 0;
};

},{}],183:[function(require,module,exports){
var _ = require('lodash');
var React = require('react');

var noop = require('./noop');

module.exports = React.createClass({
  getDefaultProps: function() {
    return {
      selected: '',
      options: [],
      onChange: noop
    };
  },
  render: function() {
    var p = this.props;

    return React.DOM.span({className: 'pv-dataviewer__filter'}, p.options.concat(['']).map(function(opt) {
      return React.DOM.label({}, React.DOM.input({
        key: opt,
        type: 'radio',
        checked: p.selected === opt,
        onChange: p.onChange.bind(this, opt),
      }), opt === '' ? '(All)' : opt);
    }));
  },
});

},{"./noop":188,"lodash":23,"react":178}],188:[function(require,module,exports){
module.exports = function() {};

},{}]},{},[181])
//# sourceMappingURL=data.js.map
