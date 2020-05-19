exports = module.exports = require('./lib/ansi');

exports.paint = function(text) {
  return new exports.Color(text);
};

exports.colorful = function() {
  // don't overwrite
  if (String.prototype.to) return;
  Object.defineProperty(String.prototype, 'to', {
    get: function() {
      return new exports.Color(this.valueOf());
    }
  });
};

exports.toxic = function() {
  // poison the String prototype
  var colors = exports.color;
  Object.keys(colors).forEach(function(key) {
    var fn = colors[key];
    Object.defineProperty(String.prototype, key, {
      get: function() {
        return fn(this.valueOf());
      }
    });
  });
};

Object.defineProperty(exports, 'isSupported', {
  get: exports.isColorSupported
});
