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

var GridEnd = /*#__PURE__*/function (_Declaration) {
  _inheritsLoose(GridEnd, _Declaration);

  var _super = _createSuper(GridEnd);

  function GridEnd() {
    return _Declaration.apply(this, arguments) || this;
  }

  var _proto = GridEnd.prototype;

  /**
   * Change repeating syntax for IE
   */
  _proto.insert = function insert(decl, prefix, prefixes, result) {
    if (prefix !== '-ms-') return _Declaration.prototype.insert.call(this, decl, prefix, prefixes);
    var clonedDecl = this.clone(decl);
    var startProp = decl.prop.replace(/end$/, 'start');
    var spanProp = prefix + decl.prop.replace(/end$/, 'span');

    if (decl.parent.some(function (i) {
      return i.prop === spanProp;
    })) {
      return undefined;
    }

    clonedDecl.prop = spanProp;

    if (decl.value.includes('span')) {
      clonedDecl.value = decl.value.replace(/span\s/i, '');
    } else {
      var startDecl;
      decl.parent.walkDecls(startProp, function (d) {
        startDecl = d;
      });

      if (startDecl) {
        var value = Number(decl.value) - Number(startDecl.value) + '';
        clonedDecl.value = value;
      } else {
        decl.warn(result, "Can not prefix " + decl.prop + " (" + startProp + " is not found)");
      }
    }

    decl.cloneBefore(clonedDecl);
    return undefined;
  };

  return GridEnd;
}(Declaration);

_defineProperty(GridEnd, "names", ['grid-row-end', 'grid-column-end']);

module.exports = GridEnd;