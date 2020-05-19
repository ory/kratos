"use strict";

function _defaults(obj, defaults) { var keys = Object.getOwnPropertyNames(defaults); for (var i = 0; i < keys.length; i++) { var key = keys[i]; var value = Object.getOwnPropertyDescriptor(defaults, key); if (value && value.configurable && obj[key] === undefined) { Object.defineProperty(obj, key, value); } } return obj; }

function _createSuper(Derived) { return function () { var Super = _getPrototypeOf(Derived), result; if (_isNativeReflectConstruct()) { var NewTarget = _getPrototypeOf(this).constructor; result = Reflect.construct(Super, arguments, NewTarget); } else { result = Super.apply(this, arguments); } return _possibleConstructorReturn(this, result); }; }

function _possibleConstructorReturn(self, call) { if (call && (typeof call === "object" || typeof call === "function")) { return call; } return _assertThisInitialized(self); }

function _assertThisInitialized(self) { if (self === void 0) { throw new ReferenceError("this hasn't been initialised - super() hasn't been called"); } return self; }

function _isNativeReflectConstruct() { if (typeof Reflect === "undefined" || !Reflect.construct) return false; if (Reflect.construct.sham) return false; if (typeof Proxy === "function") return true; try { Date.prototype.toString.call(Reflect.construct(Date, [], function () {})); return true; } catch (e) { return false; } }

function _getPrototypeOf(o) { _getPrototypeOf = Object.setPrototypeOf ? Object.getPrototypeOf : function _getPrototypeOf(o) { return o.__proto__ || Object.getPrototypeOf(o); }; return _getPrototypeOf(o); }

function _inheritsLoose(subClass, superClass) { subClass.prototype = Object.create(superClass.prototype); subClass.prototype.constructor = subClass; _defaults(subClass, superClass); }

function _defineProperty(obj, key, value) { if (key in obj) { Object.defineProperty(obj, key, { value: value, enumerable: true, configurable: true, writable: true }); } else { obj[key] = value; } return obj; }

var OldValue = require('../old-value');

var Value = require('../value');

function _regexp(name) {
  return new RegExp("(^|[\\s,(])(" + name + "($|[\\s),]))", 'gi');
}

var Intrinsic = /*#__PURE__*/function (_Value) {
  _inheritsLoose(Intrinsic, _Value);

  var _super = _createSuper(Intrinsic);

  function Intrinsic() {
    return _Value.apply(this, arguments) || this;
  }

  var _proto = Intrinsic.prototype;

  _proto.regexp = function regexp() {
    if (!this.regexpCache) this.regexpCache = _regexp(this.name);
    return this.regexpCache;
  };

  _proto.isStretch = function isStretch() {
    return this.name === 'stretch' || this.name === 'fill' || this.name === 'fill-available';
  };

  _proto.replace = function replace(string, prefix) {
    if (prefix === '-moz-' && this.isStretch()) {
      return string.replace(this.regexp(), '$1-moz-available$3');
    }

    if (prefix === '-webkit-' && this.isStretch()) {
      return string.replace(this.regexp(), '$1-webkit-fill-available$3');
    }

    return _Value.prototype.replace.call(this, string, prefix);
  };

  _proto.old = function old(prefix) {
    var prefixed = prefix + this.name;

    if (this.isStretch()) {
      if (prefix === '-moz-') {
        prefixed = '-moz-available';
      } else if (prefix === '-webkit-') {
        prefixed = '-webkit-fill-available';
      }
    }

    return new OldValue(this.name, prefixed, prefixed, _regexp(prefixed));
  };

  _proto.add = function add(decl, prefix) {
    if (decl.prop.includes('grid') && prefix !== '-webkit-') {
      return undefined;
    }

    return _Value.prototype.add.call(this, decl, prefix);
  };

  return Intrinsic;
}(Value);

_defineProperty(Intrinsic, "names", ['max-content', 'min-content', 'fit-content', 'fill', 'fill-available', 'stretch']);

module.exports = Intrinsic;