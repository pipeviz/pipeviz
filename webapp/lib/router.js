var Router = require('ampersand-router');
var main = require('./utils/main');

module.exports = Router.extend({
  routes: {
    '': 'home',
    '(*path)': 'catchAll'
  },

  home: function () {
    this.trigger('newPage', main.App);
  },

  catchAll: function () {
    this.redirectTo('');
  }
});
