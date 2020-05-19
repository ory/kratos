"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
const NormalizedMessage_1 = require("./NormalizedMessage");
const FsHelper_1 = require("./FsHelper");
const NormalizedMessageFactories_1 = require("./NormalizedMessageFactories");
const path = require("path");
function createEslinter(eslintOptions) {
    // tslint:disable-next-line:no-implicit-dependencies
    const eslint = require('eslint');
    // See https://eslint.org/docs/1.0.0/developer-guide/nodejs-api#cliengine
    const eslinter = new eslint.CLIEngine(eslintOptions);
    const createNormalizedMessageFromEsLintFailure = NormalizedMessageFactories_1.makeCreateNormalizedMessageFromEsLintFailure();
    function getLintsForFile(filepath) {
        try {
            if (eslinter.isPathIgnored(filepath) ||
                path.extname(filepath).localeCompare('.json', undefined, {
                    sensitivity: 'accent'
                }) === 0) {
                return undefined;
            }
            const lints = eslinter.executeOnFiles([filepath]);
            return lints;
        }
        catch (e) {
            FsHelper_1.throwIfIsInvalidSourceFileError(filepath, e);
        }
        return undefined;
    }
    function getFormattedLints(lintReports) {
        const allEsLints = [];
        for (const value of lintReports) {
            for (const lint of value.results) {
                allEsLints.push(...lint.messages.map(message => createNormalizedMessageFromEsLintFailure(message, lint.filePath)));
            }
        }
        return NormalizedMessage_1.NormalizedMessage.deduplicate(allEsLints);
    }
    return { getLints: getLintsForFile, getFormattedLints };
}
exports.createEslinter = createEslinter;
//# sourceMappingURL=createEslinter.js.map