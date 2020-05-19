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

var ImageRendering = /*#__PURE__*/function (_Declaration) {
  _inheritsLoose(ImageRendering, _Declaration);

  var _super = _createSuper(ImageRendering);

  function ImageRendering() {
    return _Declaration.apply(this, arguments) || this;
  }

  var _proto = ImageRendering.prototype;

  /**
   * Add hack only for crisp-edges
   */
  _proto.check = function check(decl) {
    return decl.value === 'pixelated';
  }
  /**
   * Change property name for IE
   */
  ;

  _proto.prefixed = function prefixed(prop, prefix) {
    if (prefix === '-ms-') {
      return '-ms-interpolation-mode';
    }

    return _Declaration.prototype.prefixed.call(this, prop, prefix);
  }
  /**
   * Change property and value for IE
   */
  ;

  _proto.set = function set(decl, prefix) {
    if (prefix !== '-ms-') return _Declaration.prototype.set.call(this, decl, prefix);
    decl.prop = '-ms-interpolation-mode';
    decl.value = 'nearest-neighbor';
    return decl;
  }
  /**
   * Return property name by spec
   */
  ;

  _proto.normalize = function normalize() {
    return 'image-rendering';
  }
  /**
   * Warn on old value
   */
  ;

  _proto.process = function process(node, result) {
    return _Declaration.prototype.process.call(this, node, result);
  };

  return ImageRendering;
}(Declaration);

_defineProperty(ImageRendering, "names", ['image-rendering', 'interpolation-mode']);

module.exports = ImageRendering;