"use strict";

exports.__esModule = true;
exports["default"] = void 0;

var _crypto = _interopRequireDefault(require("crypto"));

function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { "default": obj }; }

function computeIntegrity(algorithms, source) {
  return Array.isArray(algorithms) ? algorithms.map(function (algorithm) {
    var hash = _crypto["default"].createHash(algorithm).update(source, 'utf8').digest('base64');

    return algorithm + "-" + hash;
  }).join(' ') : '';
}

var _default = computeIntegrity;
exports["default"] = _default;