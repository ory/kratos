"use strict";

function _defaults(obj, defaults) { var keys = Object.getOwnPropertyNames(defaults); for (var i = 0; i < keys.length; i++) { var key = keys[i]; var value = Object.getOwnPropertyDescriptor(defaults, key); if (value && value.configurable && obj[key] === undefined) { Object.defineProperty(obj, key, value); } } return obj; }

function _createForOfIteratorHelperLoose(o) { var i = 0; if (typeof Symbol === "undefined" || o[Symbol.iterator] == null) { if (Array.isArray(o) || (o = _unsupportedIterableToArray(o))) return function () { if (i >= o.length) return { done: true }; return { done: false, value: o[i++] }; }; throw new TypeError("Invalid attempt to iterate non-iterable instance.\nIn order to be iterable, non-array objects must have a [Symbol.iterator]() method."); } i = o[Symbol.iterator](); return i.next.bind(i); }

function _unsupportedIterableToArray(o, minLen) { if (!o) return; if (typeof o === "string") return _arrayLikeToArray(o, minLen); var n = Object.prototype.toString.call(o).slice(8, -1); if (n === "Object" && o.constructor) n = o.constructor.name; if (n === "Map" || n === "Set") return Array.from(n); if (n === "Arguments" || /^(?:Ui|I)nt(?:8|16|32)(?:Clamped)?Array$/.test(n)) return _arrayLikeToArray(o, minLen); }

function _arrayLikeToArray(arr, len) { if (len == null || len > arr.length) len = arr.length; for (var i = 0, arr2 = new Array(len); i < len; i++) { arr2[i] = arr[i]; } return arr2; }

function _createSuper(Derived) { return function () { var Super = _getPrototypeOf(Derived), result; if (_isNativeReflectConstruct()) { var NewTarget = _getPrototypeOf(this).constructor; result = Reflect.construct(Super, arguments, NewTarget); } else { result = Super.apply(this, arguments); } return _possibleConstructorReturn(this, result); }; }

function _possibleConstructorReturn(self, call) { if (call && (typeof call === "object" || typeof call === "function")) { return call; } return _assertThisInitialized(self); }

function _assertThisInitialized(self) { if (self === void 0) { throw new ReferenceError("this hasn't been initialised - super() hasn't been called"); } return self; }

function _isNativeReflectConstruct() { if (typeof Reflect === "undefined" || !Reflect.construct) return false; if (Reflect.construct.sham) return false; if (typeof Proxy === "function") return true; try { Date.prototype.toString.call(Reflect.construct(Date, [], function () {})); return true; } catch (e) { return false; } }

function _getPrototypeOf(o) { _getPrototypeOf = Object.setPrototypeOf ? Object.getPrototypeOf : function _getPrototypeOf(o) { return o.__proto__ || Object.getPrototypeOf(o); }; return _getPrototypeOf(o); }

function _inheritsLoose(subClass, superClass) { subClass.prototype = Object.create(superClass.prototype); subClass.prototype.constructor = subClass; _defaults(subClass, superClass); }

var n2f = require('num2fraction');

var Prefixer = require('./prefixer');

var utils = require('./utils');

var REGEXP = /(min|max)-resolution\s*:\s*\d*\.?\d+(dppx|dpi|x)/gi;
var SPLIT = /(min|max)-resolution(\s*:\s*)(\d*\.?\d+)(dppx|dpi|x)/i;

var Resolution = /*#__PURE__*/function (_Prefixer) {
  _inheritsLoose(Resolution, _Prefixer);

  var _super = _createSuper(Resolution);

  function Resolution() {
    return _Prefixer.apply(this, arguments) || this;
  }

  var _proto = Resolution.prototype;

  /**
     * Return prefixed query name
     */
  _proto.prefixName = function prefixName(prefix, name) {
    if (prefix === '-moz-') {
      return name + '--moz-device-pixel-ratio';
    } else {
      return prefix + name + '-device-pixel-ratio';
    }
  }
  /**
     * Return prefixed query
     */
  ;

  _proto.prefixQuery = function prefixQuery(prefix, name, colon, value, units) {
    if (units === 'dpi') {
      value = Number(value / 96);
    }

    if (prefix === '-o-') {
      value = n2f(value);
    }

    return this.prefixName(prefix, name) + colon + value;
  }
  /**
     * Remove prefixed queries
     */
  ;

  _proto.clean = function clean(rule) {
    var _this = this;

    if (!this.bad) {
      this.bad = [];

      for (var _iterator = _createForOfIteratorHelperLoose(this.prefixes), _step; !(_step = _iterator()).done;) {
        var prefix = _step.value;
        this.bad.push(this.prefixName(prefix, 'min'));
        this.bad.push(this.prefixName(prefix, 'max'));
      }
    }

    rule.params = utils.editList(rule.params, function (queries) {
      return queries.filter(function (query) {
        return _this.bad.every(function (i) {
          return !query.includes(i);
        });
      });
    });
  }
  /**
     * Add prefixed queries
     */
  ;

  _proto.process = function process(rule) {
    var _this2 = this;

    var parent = this.parentPrefix(rule);
    var prefixes = parent ? [parent] : this.prefixes;
    rule.params = utils.editList(rule.params, function (origin, prefixed) {
      for (var _iterator2 = _createForOfIteratorHelperLoose(origin), _step2; !(_step2 = _iterator2()).done;) {
        var query = _step2.value;

        if (!query.includes('min-resolution') && !query.includes('max-resolution')) {
          prefixed.push(query);
          continue;
        }

        var _loop = function _loop() {
          var prefix = _step3.value;
          var processed = query.replace(REGEXP, function (str) {
            var parts = str.match(SPLIT);
            return _this2.prefixQuery(prefix, parts[1], parts[2], parts[3], parts[4]);
          });
          prefixed.push(processed);
        };

        for (var _iterator3 = _createForOfIteratorHelperLoose(prefixes), _step3; !(_step3 = _iterator3()).done;) {
          _loop();
        }

        prefixed.push(query);
      }

      return utils.uniq(prefixed);
    });
  };

  return Resolution;
}(Prefixer);

module.exports = Resolution;