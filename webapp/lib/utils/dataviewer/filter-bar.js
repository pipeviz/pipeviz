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
