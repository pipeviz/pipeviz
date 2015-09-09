var Router = require('ampersand-router');
var commits = require('./page/commits');
var data = require('./page/data');

module.exports = Router.extend({
  routes: {
    '': 'commits',
    'data': 'data',
    'commits': 'commits',
    '(*path)': 'catchAll'
  },

  commits: function () {
    this.trigger('newPage', commits.App);
  },

  data: function () {
    this.trigger('newPage', data.App);
  },

  catchAll: function () {
    this.redirectTo('');
  }
});
