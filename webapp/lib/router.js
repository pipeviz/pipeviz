var Router = require('ampersand-router');
var commitPipeline = require('./page/commit-pipeline');
var data = require('./page/data');

module.exports = Router.extend({
  routes: {
    '': 'commitPipeline',
    'data': 'data',
    'commit-pipeline': 'commitPipeline',
    '(*path)': 'catchAll'
  },

  commitPipeline: function () {
    this.trigger('newPage', commitPipeline.App);
  },

  data: function () {
    this.trigger('newPage', data.App);
  },

  catchAll: function () {
    this.redirectTo('');
  }
});
