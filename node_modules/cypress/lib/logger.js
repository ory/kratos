'use strict';

var R = require('ramda');
var chalk = require('chalk');

var logs = [];

var logLevel = function logLevel() {
  return process.env.npm_config_loglevel || 'notice';
};

var error = function error() {
  for (var _len = arguments.length, messages = Array(_len), _key = 0; _key < _len; _key++) {
    messages[_key] = arguments[_key];
  }

  logs.push(messages.join(' '));
  console.log(chalk.red.apply(chalk, messages)); // eslint-disable-line no-console
};

var warn = function warn() {
  for (var _len2 = arguments.length, messages = Array(_len2), _key2 = 0; _key2 < _len2; _key2++) {
    messages[_key2] = arguments[_key2];
  }

  if (logLevel() === 'silent') return;

  logs.push(messages.join(' '));
  console.log(chalk.yellow.apply(chalk, messages)); // eslint-disable-line no-console
};

var log = function log() {
  var _console;

  for (var _len3 = arguments.length, messages = Array(_len3), _key3 = 0; _key3 < _len3; _key3++) {
    messages[_key3] = arguments[_key3];
  }

  if (logLevel() === 'silent' || logLevel() === 'warn') return;

  logs.push(messages.join(' '));
  (_console = console).log.apply(_console, messages); // eslint-disable-line no-console
};

// splits long text into lines and calls log()
// on each one to allow easy unit testing for specific message
var logLines = function logLines(text) {
  var lines = text.split('\n');

  R.forEach(log, lines);
};

var print = function print() {
  return logs.join('\n');
};

var reset = function reset() {
  logs = [];
};

module.exports = {
  log: log,
  warn: warn,
  error: error,
  logLines: logLines,
  print: print,
  reset: reset,
  logLevel: logLevel
};