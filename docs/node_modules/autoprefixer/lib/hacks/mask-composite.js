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

var MaskComposite = /*#__PURE__*/function (_Declaration) {
  _inheritsLoose(MaskComposite, _Declaration);

  var _super = _createSuper(MaskComposite);

  function MaskComposite() {
    return _Declaration.apply(this, arguments) || this;
  }

  var _proto = MaskComposite.prototype;

  /**
   * Prefix mask-composite for webkit
   */
  _proto.insert = function insert(decl, prefix, prefixes) {
    var isCompositeProp = decl.prop === 'mask-composite';
    var compositeValues;

    if (isCompositeProp) {
      compositeValues = decl.value.split(',');
    } else {
      compositeValues = decl.value.match(MaskComposite.regexp) || [];
    }

    compositeValues = compositeValues.map(function (el) {
      return el.trim();
    }).filter(function (el) {
      return el;
    });
    var hasCompositeValues = compositeValues.length;
    var compositeDecl;

    if (hasCompositeValues) {
      compositeDecl = this.clone(decl);
      compositeDecl.value = compositeValues.map(function (value) {
        return MaskComposite.oldValues[value] || value;
      }).join(', ');

      if (compositeValues.includes('intersect')) {
        compositeDecl.value += ', xor';
      }

      compositeDecl.prop = prefix + 'mask-composite';
    }

    if (isCompositeProp) {
      if (!hasCompositeValues) {
        return undefined;
      }

      if (this.needCascade(decl)) {
        compositeDecl.raws.before = this.calcBefore(prefixes, decl, prefix);
      }

      return decl.parent.insertBefore(decl, compositeDecl);
    }

    var cloned = this.clone(decl);
    cloned.prop = prefix + cloned.prop;

    if (hasCompositeValues) {
      cloned.value = cloned.value.replace(MaskComposite.regexp, '');
    }

    if (this.needCascade(decl)) {
      cloned.raws.before = this.calcBefore(prefixes, decl, prefix);
    }

    decl.parent.insertBefore(decl, cloned);

    if (!hasCompositeValues) {
      return decl;
    }

    if (this.needCascade(decl)) {
      compositeDecl.raws.before = this.calcBefore(prefixes, decl, prefix);
    }

    return decl.parent.insertBefore(decl, compositeDecl);
  };

  return MaskComposite;
}(Declaration);

_defineProperty(MaskComposite, "names", ['mask', 'mask-composite']);

_defineProperty(MaskComposite, "oldValues", {
  add: 'source-over',
  substract: 'source-out',
  intersect: 'source-in',
  exclude: 'xor'
});

_defineProperty(MaskComposite, "regexp", new RegExp("\\s+(" + Object.keys(MaskComposite.oldValues).join('|') + ")\\b(?!\\))\\s*(?=[,])", 'ig'));

module.exports = MaskComposite;