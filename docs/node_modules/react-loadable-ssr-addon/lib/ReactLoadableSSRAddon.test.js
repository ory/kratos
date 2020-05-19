"use strict";

var _ava = _interopRequireDefault(require("ava"));

var _path = _interopRequireDefault(require("path"));

var _fs = _interopRequireDefault(require("fs"));

var _webpack = _interopRequireDefault(require("webpack"));

var _webpack2 = _interopRequireDefault(require("../webpack.config"));

var _ReactLoadableSSRAddon = _interopRequireWildcard(require("./ReactLoadableSSRAddon"));

function _getRequireWildcardCache() { if (typeof WeakMap !== "function") return null; var cache = new WeakMap(); _getRequireWildcardCache = function _getRequireWildcardCache() { return cache; }; return cache; }

function _interopRequireWildcard(obj) { if (obj && obj.__esModule) { return obj; } var cache = _getRequireWildcardCache(); if (cache && cache.has(obj)) { return cache.get(obj); } var newObj = {}; if (obj != null) { var hasPropertyDescriptor = Object.defineProperty && Object.getOwnPropertyDescriptor; for (var key in obj) { if (Object.prototype.hasOwnProperty.call(obj, key)) { var desc = hasPropertyDescriptor ? Object.getOwnPropertyDescriptor(obj, key) : null; if (desc && (desc.get || desc.set)) { Object.defineProperty(newObj, key, desc); } else { newObj[key] = obj[key]; } } } } newObj["default"] = obj; if (cache) { cache.set(obj, newObj); } return newObj; }

function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { "default": obj }; }

var outputPath;
var manifestOutputPath;

var runWebpack = function runWebpack(configuration, end, callback) {
  (0, _webpack["default"])(configuration, function (err, stats) {
    if (err) {
      return end(err);
    }

    if (stats.hasErrors()) {
      return end(stats.toString());
    }

    callback();
    end();
  });
};

_ava["default"].beforeEach(function () {
  var publicPathSanitized = _webpack2["default"].output.publicPath.slice(1, -1);

  outputPath = _path["default"].resolve('./example', publicPathSanitized);
  manifestOutputPath = _path["default"].resolve(outputPath, _ReactLoadableSSRAddon.defaultOptions.filename);
});

_ava["default"].cb('outputs with default settings', function (t) {
  _webpack2["default"].plugins = [new _ReactLoadableSSRAddon["default"]()];
  runWebpack(_webpack2["default"], t.end, function () {
    var feedback = _fs["default"].existsSync(manifestOutputPath) ? 'pass' : 'fail';
    t[feedback]();
  });
});

_ava["default"].cb('outputs with custom filename', function (t) {
  var filename = 'new-assets-manifest.json';
  _webpack2["default"].plugins = [new _ReactLoadableSSRAddon["default"]({
    filename: filename
  })];
  runWebpack(_webpack2["default"], t.end, function () {
    var feedback = _fs["default"].existsSync(manifestOutputPath.replace(_ReactLoadableSSRAddon.defaultOptions.filename, filename)) ? 'pass' : 'fail';
    t[feedback]();
  });
});

_ava["default"].cb('outputs with integrity', function (t) {
  _webpack2["default"].plugins = [new _ReactLoadableSSRAddon["default"]({
    integrity: true
  })];
  runWebpack(_webpack2["default"], t.end, function () {
    var manifest = require("" + manifestOutputPath);

    Object.keys(manifest.assets).forEach(function (asset) {
      manifest.assets[asset].js.forEach(function (_ref) {
        var integrity = _ref.integrity;
        t["false"](!integrity);
      });
    });
  });
});