"use strict";

function _defaults(obj, defaults) { var keys = Object.getOwnPropertyNames(defaults); for (var i = 0; i < keys.length; i++) { var key = keys[i]; var value = Object.getOwnPropertyDescriptor(defaults, key); if (value && value.configurable && obj[key] === undefined) { Object.defineProperty(obj, key, value); } } return obj; }

function _createSuper(Derived) { return function () { var Super = _getPrototypeOf(Derived), result; if (_isNativeReflectConstruct()) { var NewTarget = _getPrototypeOf(this).constructor; result = Reflect.construct(Super, arguments, NewTarget); } else { result = Super.apply(this, arguments); } return _possibleConstructorReturn(this, result); }; }

function _possibleConstructorReturn(self, call) { if (call && (typeof call === "object" || typeof call === "function")) { return call; } return _assertThisInitialized(self); }

function _assertThisInitialized(self) { if (self === void 0) { throw new ReferenceError("this hasn't been initialised - super() hasn't been called"); } return self; }

function _isNativeReflectConstruct() { if (typeof Reflect === "undefined" || !Reflect.construct) return false; if (Reflect.construct.sham) return false; if (typeof Proxy === "function") return true; try { Date.prototype.toString.call(Reflect.construct(Date, [], function () {})); return true; } catch (e) { return false; } }

function _getPrototypeOf(o) { _getPrototypeOf = Object.setPrototypeOf ? Object.getPrototypeOf : function _getPrototypeOf(o) { return o.__proto__ || Object.getPrototypeOf(o); }; return _getPrototypeOf(o); }

function _inheritsLoose(subClass, superClass) { subClass.prototype = Object.create(superClass.prototype); subClass.prototype.constructor = subClass; _defaults(subClass, superClass); }

function _defineProperty(obj, key, value) { if (key in obj) { Object.defineProperty(obj, key, { value: value, enumerable: true, configurable: true, writable: true }); } else { obj[key] = value; } return obj; }

var Declaration = require('../declaration');

var BreakProps = /*#__PURE__*/function (_Declaration) {
  _inheritsLoose(BreakProps, _Declaration);

  var _super = _createSuper(BreakProps);

  function BreakProps() {
    return _Declaration.apply(this, arguments) || this;
  }

  var _proto = BreakProps.prototype;

  /**
   * Change name for -webkit- and -moz- prefix
   */
  _proto.prefixed = function prefixed(prop, prefix) {
    return prefix + "column-" + prop;
  }
  /**
   * Return property name by final spec
   */
  ;

  _proto.normalize = function normalize(prop) {
    if (prop.includes('inside')) {
      return 'break-inside';
    }

    if (prop.includes('before')) {
      return 'break-before';
    }

    return 'break-after';
  }
  /**
   * Change prefixed value for avoid-column and avoid-page
   */
  ;

  _proto.set = function set(decl, prefix) {
    if (decl.prop === 'break-inside' && decl.value === 'avoid-column' || decl.value === 'avoid-page') {
      decl.value = 'avoid';
    }

    return _Declaration.prototype.set.call(this, decl, prefix);
  }
  /**
   * Don’t prefix some values
   */
  ;

  _proto.insert = function insert(decl, prefix, prefixes) {
    if (decl.prop !== 'break-inside') {
      return _Declaration.prototype.insert.call(this, decl, prefix, prefixes);
    }

    if (/region/i.test(decl.value) || /page/i.test(decl.value)) {
      return undefined;
    }

    return _Declaration.prototype.insert.call(this, decl, prefix, prefixes);
  };

  return BreakProps;
}(Declaration);

_defineProperty(BreakProps, "names", ['break-inside', 'page-break-inside', 'column-break-inside', 'break-before', 'page-break-before', 'column-break-before', 'break-after', 'page-break-after', 'column-break-after']);

module.exports = BreakProps;