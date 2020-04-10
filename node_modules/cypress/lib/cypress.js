'use strict';

// https://github.com/cypress-io/cypress/issues/316

var Promise = require('bluebird');
var tmp = Promise.promisifyAll(require('tmp'));

var fs = require('./fs');
var _open = require('./exec/open');
var _run = require('./exec/run');
var util = require('./util');

var cypressModuleApi = {
  open: function open() {
    var options = arguments.length > 0 && arguments[0] !== undefined ? arguments[0] : {};

    options = util.normalizeModuleOptions(options);

    return _open.start(options);
  },
  run: function run() {
    var options = arguments.length > 0 && arguments[0] !== undefined ? arguments[0] : {};

    options = util.normalizeModuleOptions(options);

    return tmp.fileAsync().then(function (outputPath) {
      options.outputPath = outputPath;

      return _run.start(options).then(function (failedTests) {
        return fs.readJsonAsync(outputPath, { throws: false }).then(function (output) {
          if (!output) {
            return {
              failures: failedTests,
              message: 'Could not find Cypress test run results'
            };
          }

          return output;
        });
      });
    });
  }
};

module.exports = cypressModuleApi;