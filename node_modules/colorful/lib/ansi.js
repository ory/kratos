
/* ANSI color support in terminal
 * @author: Hsiaoming Yang <lepture@me.com>
 *
 * http://en.wikipedia.org/wiki/ANSI_escape_code
 */


var util = require('util');
var tty = require('tty');

exports.disabled = false;
exports.isatty = false;

function isColorSupported() {
  if (exports.disabled) return false;

  // you can force to tty
  if (!exports.isatty && !tty.isatty()) return false;

  if ('COLORTERM' in process.env) return true;
  // windows will support color
  if (process.platform === 'win32') return true;

  var term = process.env.TERM;
  if (!term) return false;

  term = term.toLowerCase();
  if (term.indexOf('color') !== -1) return true;
  return term === 'xterm' || term === 'linux';
}


function is256ColorSupported() {
  if (!isColorSupported()) return false;

  var term = process.env.TERM;
  if (!term) return false;

  term = term.toLowerCase();
  return term.indexOf('256') !== -1;
}
exports.isColorSupported = isColorSupported;
exports.is256ColorSupported = is256ColorSupported;

var colors = [
  'black', 'red', 'green', 'yellow', 'blue',
  'magenta', 'cyan', 'white'
];

var styles = [
  'bold', 'faint', 'italic', 'underline', 'blink', 'overline',
  'inverse', 'conceal', 'strike'
];

exports.color = {};

function Color(obj) {
  this.string = obj;

  this.styles = [];
  this.fgcolor = null;
  this.bgcolor = null;
}
util.inherits(Color, String);

for (var i = 0; i < colors.length; i++) {
  (function(i) {
    var name = colors[i];
    Object.defineProperty(Color.prototype, name, {
      get: function() {
        this.fgcolor = i;
        return this;
      }
    });
    Object.defineProperty(Color.prototype, name + '_bg', {
      get: function() {
        this.bgcolor = i;
        return this;
      }
    });
    exports.color[name] = exports[name] = function(text) {
      if (!isColorSupported()) return text;
      return '\x1b[' + (30 + i) + 'm' + text + '\x1b[0m';
    };
    exports.color[name + '_bg'] = exports[name + '_bg'] = function(text) {
      if (!isColorSupported()) return text;
      return '\x1b[' + (40 + i) + 'm' + text + '\x1b[0m';
    };
  })(i);
}
for (var i = 0; i < styles.length; i++) {
  (function(i) {
    var name = styles[i];
    Object.defineProperty(Color.prototype, name, {
      get: function() {
        if (this.styles.indexOf(i) === -1) {
          this.styles = this.styles.concat(i + 1);
        }
        return this;
      }
    });
    exports.color[name] = exports[name] = function(text) {
      if (!isColorSupported()) return text;
      return '\x1b[' + (i + 1) + 'm' + text + '\x1b[0m';
    };
  })(i);
}

exports.color.grey = exports.color.gray = exports.grey = exports.gray = function(text) {
  if (!isColorSupported()) return text;
  if (is256ColorSupported()) {
    return '\x1b[38;5;8m' + text + '\x1b[0m';
  }
  return '\x1b[30;1m' + text + '\x1b[0m';
};
Object.defineProperty(Color.prototype, 'gray', {
  get: function() {
    if (isColorSupported) {
      this.fgcolor = 8;
    } else {
      this.styles = this.styles.concat(1);
      this.fgcolor = 0;
    }
    return this;
  }
});
Object.defineProperty(Color.prototype, 'grey', {
  get: function() {
    if (isColorSupported) {
      this.fgcolor = 8;
    } else {
      this.styles = this.styles.concat(1);
      this.fgcolor = 0;
    }
    return this;
  }
});


Color.prototype.valueOf = function() {
  if (!isColorSupported()) return this.string;
  var is256 = is256ColorSupported();

  var text = this.string;
  var reset = '\x1b[0m';

  if (is256) {
    if (this.fgcolor !== null) {
      text = '\x1b[38;5;' + this.fgcolor + 'm' + text + reset;
    }
    if (this.bgcolor !== null) {
      text = '\x1b[48;5;' + this.bgcolor + 'm' + text + reset;
    }
  } else {
    if (this.fgcolor !== null && this.fgcolor < 8) {
      text = '\x1b[' + (30 + this.fgcolor) + 'm' + text + reset;
    }
    if (this.bgcolor !== null && this.bgcolor < 8) {
      text = '\x1b[' + (40 + this.bgcolor) + 'm' + text + reset;
    }
  }
  if (this.styles.length) {
    text = '\x1b[' + this.styles.join(';') + 'm' + text + reset;
  }
  return text;
};
Color.prototype.toString = function() {
  return this.valueOf();
};
Object.defineProperty(Color.prototype, 'color', {
  get: function() {
    return this.valueOf();
  }
});
Object.defineProperty(Color.prototype, 'style', {
  get: function() {
    return this.valueOf();
  }
});
Object.defineProperty(Color.prototype, 'length', {
  get: function() {
    return this.string.length;
  }
});

exports.Color = Color;
