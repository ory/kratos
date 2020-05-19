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

var JustifyContent = /*#__PURE__*/function (_Declaration) {
  _inheritsLoose(JustifyContent, _Declaration);

  var _super = _createSuper(JustifyContent);

  function JustifyContent() {
    return _Declaration.apply(this, arguments) || this;
  }

  var _proto = JustifyContent.prototype;

  /**
   * Change property name for 2009 and 2012 specs
   */
  _proto.prefixed = function prefixed(prop, prefix) {
    var spec;

    var _flexSpec = flexSpec(prefix);

    spec = _flexSpec[0];
    prefix = _flexSpec[1];

    if (spec === 2009) {
      return prefix + 'box-pack';
    }

    if (spec === 2012) {
      return prefix + 'flex-pack';
    }

    return _Declaration.prototype.prefixed.call(this, prop, prefix);
  }
  /**
   * Return property name by final spec
   */
  ;

  _proto.normalize = function normalize() {
    return 'justify-content';
  }
  /**
   * Change value for 2009 and 2012 specs
   */
  ;

  _proto.set = function set(decl, prefix) {
    var spec = flexSpec(prefix)[0];

    if (spec === 2009 || spec === 2012) {
      var value = JustifyContent.oldValues[decl.value] || decl.value;
      decl.value = value;

      if (spec !== 2009 || value !== 'distribute') {
        return _Declaration.prototype.set.call(this, decl, prefix);
      }
    } else if (spec === 'final') {
      return _Declaration.prototype.set.call(this, decl, prefix);
    }

    return undefined;
  };

  return JustifyContent;
}(Declaration);

_defineProperty(JustifyContent, "names", ['justify-content', 'flex-pack', 'box-pack']);

_defineProperty(JustifyContent, "oldValues", {
  'flex-end': 'end',
  'flex-start': 'start',
  'space-between': 'justify',
  'space-around': 'distribute'
});

module.exports = JustifyContent;