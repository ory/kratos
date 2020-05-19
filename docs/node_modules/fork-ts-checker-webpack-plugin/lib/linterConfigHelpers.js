"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
const path = require("path");
const minimatch = require("minimatch");
const FsHelper_1 = require("./FsHelper");
function loadLinterConfig(configFile) {
    // tslint:disable-next-line:no-implicit-dependencies
    const tslint = require('tslint');
    return tslint.Configuration.loadConfigurationFromPath(configFile);
}
exports.loadLinterConfig = loadLinterConfig;
function makeGetLinterConfig(linterConfigs, linterExclusions, context) {
    const getLinterConfig = (file) => {
        const dirname = path.dirname(file);
        if (dirname in linterConfigs) {
            return linterConfigs[dirname];
        }
        if (FsHelper_1.fileExistsSync(path.join(dirname, 'tslint.json'))) {
            const config = loadLinterConfig(path.join(dirname, 'tslint.json'));
            if (config.linterOptions && config.linterOptions.exclude) {
                linterExclusions.concat(config.linterOptions.exclude.map(pattern => new minimatch.Minimatch(path.resolve(pattern))));
            }
            linterConfigs[dirname] = config;
        }
        else {
            if (dirname !== context && dirname !== file) {
                linterConfigs[dirname] = getLinterConfig(dirname);
            }
        }
        return linterConfigs[dirname];
    };
    return getLinterConfig;
}
exports.makeGetLinterConfig = makeGetLinterConfig;
//# sourceMappingURL=linterConfigHelpers.js.map