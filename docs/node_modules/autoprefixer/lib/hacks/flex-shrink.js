"use strict";

function _defaults(obj, defaults) { var keys = Object.getOwnPropertyNames(defaults); for (var i = 0; i < keys.length; i++) { var key = keys[i]; var value = Object.getOwnPropertyDescriptor(defaults, key); if (value && value.configurable && obj[key] === undefined) { Object.defineProperty(obj, key, value); } } return obj; }

function _createSuper(Derived) { return function () { var Super = _getPrototypeOf(Derived), result; if (_isNativeReflectConstruct()) { var NewTarget = _getPrototypeOf(this).constructor; result = Reflect.construct(Super, arguments, NewTarget); } else { result = Super.apply(this, arguments); } return _possibleConstructorReturn(this, result); }; }

function _possibleConstructorReturn(self, call) { if (call && (typeof call === "object" || typeof call === "function")) { return call; } return _assertThisInitialized(self); }

function _assertThisInitialized(self) { if (self === void 0) { throw new ReferenceError("this hasn't been initialised - super() hasn't been called"); } return self; }

function _isNativeReflectConstruct() { if (typeof Reflect === "undefined" || !Reflect.construct) return false; if (Reflect.construct.sham) return false; if (typeof Proxy === "function") return true; try { Date.prototype.toString.call(Reflect.construct(Date, [], function () {})); return true; } catch (e) { return false; } }

function _getPrototypeOf(o) { _getPrototypeOf = Object.setPrototypeOf ? Object.getPrototypeOf : function _getPrototypeOf(o) { return o.__proto__ || Object.getPrototypeOf(o); }; return _getPrototypeOf(o); }

function _inheritsLoose(subClass, superClass) { subClass.prototype = Object.create(superClass.prototype); subClass.prototype.constructor = subClass; _defaults(subClass, superClass); }

function _defineProperty(obj, key, value) { if (key in obj) { Object.defineProperty(obj, key, { value: value, enumerable: true, configurable: true, writable: true }); } else { obj[key] = value; } return obj; }

var flexSpec = require('./flex-spec');

var Declaration = require('../declaration');

var FlexShrink = /*#__PURE__*/function (_Declaration) {
  _inheritsLoose(FlexShrink, _Declaration);

  var _super = _createSuper(FlexShrink);

  function FlexShrink() {
    return _Declaration.apply(this, arguments) || this;
  }

  var _proto = FlexShrink.prototype;

  /**
   * Return property name by final spec
   */
  _proto.normalize = function normalize() {
    return 'flex-shrink';
  }
  /**
   * Return flex property for 2012 spec
   */
  ;

  _proto.prefixed = function prefixed(prop, prefix) {
    var spec;

    var _flexSpec = flexSpec(prefix);

    spec = _flexSpec[0];
    prefix = _flexSpec[1];

    if (spec === 2012) {
      return prefix + 'flex-negative';
    }

    return _Declaration.prototype.prefixed.call(this, prop, prefix);
  }
  /**
   * Ignore 2009 spec and use flex property for 2012
   */
  ;

  _proto.set = function set(decl, prefix) {
    var spec;

    var _flexSpec2 = flexSpec(prefix);

    spec = _flexSpec2[0];
    prefix = _flexSpec2[1];

    if (spec === 2012 || spec === 'final') {
      return _Declaration.prototype.set.call(this, decl, prefix);
    }

    return undefined;
  };

  return FlexShrink;
}(Declaration);

_defineProperty(FlexShrink, "names", ['flex-shrink', 'flex-negative']);

module.exports = FlexShrink;