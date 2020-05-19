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

function _defineProperty(obj, key, value) { if (key in obj) { Object.defineProperty(obj, key, { value: value, enumerable: true, configurable: true, writable: true }); } else { obj[key] = value; } return obj; }

var Declaration = require('../declaration');

var TransformDecl = /*#__PURE__*/function (_Declaration) {
  _inheritsLoose(TransformDecl, _Declaration);

  var _super = _createSuper(TransformDecl);

  function TransformDecl() {
    return _Declaration.apply(this, arguments) || this;
  }

  var _proto = TransformDecl.prototype;

  /**
   * Recursively check all parents for @keyframes
   */
  _proto.keyframeParents = function keyframeParents(decl) {
    var parent = decl.parent;

    while (parent) {
      if (parent.type === 'atrule' && parent.name === 'keyframes') {
        return true;
      }

      var _parent = parent;
      parent = _parent.parent;
    }

    return false;
  }
  /**
   * Is transform contain 3D commands
   */
  ;

  _proto.contain3d = function contain3d(decl) {
    if (decl.prop === 'transform-origin') {
      return false;
    }

    for (var _iterator = _createForOfIteratorHelperLoose(TransformDecl.functions3d), _step; !(_step = _iterator()).done;) {
      var func = _step.value;

      if (decl.value.includes(func + "(")) {
        return true;
      }
    }

    return false;
  }
  /**
   * Replace rotateZ to rotate for IE 9
   */
  ;

  _proto.set = function set(decl, prefix) {
    decl = _Declaration.prototype.set.call(this, decl, prefix);

    if (prefix === '-ms-') {
      decl.value = decl.value.replace(/rotatez/gi, 'rotate');
    }

    return decl;
  }
  /**
   * Don't add prefix for IE in keyframes
   */
  ;

  _proto.insert = function insert(decl, prefix, prefixes) {
    if (prefix === '-ms-') {
      if (!this.contain3d(decl) && !this.keyframeParents(decl)) {
        return _Declaration.prototype.insert.call(this, decl, prefix, prefixes);
      }
    } else if (prefix === '-o-') {
      if (!this.contain3d(decl)) {
        return _Declaration.prototype.insert.call(this, decl, prefix, prefixes);
      }
    } else {
      return _Declaration.prototype.insert.call(this, decl, prefix, prefixes);
    }

    return undefined;
  };

  return TransformDecl;
}(Declaration);

_defineProperty(TransformDecl, "names", ['transform', 'transform-origin']);

_defineProperty(TransformDecl, "functions3d", ['matrix3d', 'translate3d', 'translateZ', 'scale3d', 'scaleZ', 'rotate3d', 'rotateX', 'rotateY', 'perspective']);

module.exports = TransformDecl;