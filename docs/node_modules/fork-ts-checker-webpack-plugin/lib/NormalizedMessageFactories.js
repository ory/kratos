"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
const NormalizedMessage_1 = require("./NormalizedMessage");
exports.makeCreateNormalizedMessageFromDiagnostic = (typescript) => {
    const createNormalizedMessageFromDiagnostic = (diagnostic) => {
        let file;
        let line;
        let character;
        if (diagnostic.file) {
            file = diagnostic.file.fileName;
            if (diagnostic.start === undefined) {
                throw new Error('Expected diagnostics to have start');
            }
            const position = diagnostic.file.getLineAndCharacterOfPosition(diagnostic.start);
            line = position.line + 1;
            character = position.character + 1;
        }
        return new NormalizedMessage_1.NormalizedMessage({
            type: NormalizedMessage_1.NormalizedMessage.TYPE_DIAGNOSTIC,
            code: diagnostic.code,
            severity: typescript.DiagnosticCategory[diagnostic.category].toLowerCase(),
            content: typescript.flattenDiagnosticMessageText(diagnostic.messageText, '\n'),
            file,
            line,
            character
        });
    };
    return createNormalizedMessageFromDiagnostic;
};
exports.makeCreateNormalizedMessageFromRuleFailure = () => {
    const createNormalizedMessageFromRuleFailure = (lint) => {
        const position = lint.getStartPosition().getLineAndCharacter();
        return new NormalizedMessage_1.NormalizedMessage({
            type: NormalizedMessage_1.NormalizedMessage.TYPE_LINT,
            code: lint.getRuleName(),
            severity: lint.getRuleSeverity(),
            content: lint.getFailure(),
            file: lint.getFileName(),
            line: position.line + 1,
            character: position.character + 1
        });
    };
    return createNormalizedMessageFromRuleFailure;
};
exports.makeCreateNormalizedMessageFromEsLintFailure = () => {
    const createNormalizedMessageFromEsLintFailure = (lint, filePath) => {
        return new NormalizedMessage_1.NormalizedMessage({
            type: NormalizedMessage_1.NormalizedMessage.TYPE_LINT,
            code: lint.ruleId,
            severity: lint.severity === 1
                ? NormalizedMessage_1.NormalizedMessage.SEVERITY_WARNING
                : NormalizedMessage_1.NormalizedMessage.SEVERITY_ERROR,
            content: lint.message,
            file: filePath,
            line: lint.line,
            character: lint.column
        });
    };
    return createNormalizedMessageFromEsLintFailure;
};
exports.makeCreateNormalizedMessageFromInternalError = () => {
    const createNormalizedMessageFromInternalError = (error) => {
        return new NormalizedMessage_1.NormalizedMessage({
            type: NormalizedMessage_1.NormalizedMessage.TYPE_DIAGNOSTIC,
            severity: NormalizedMessage_1.NormalizedMessage.SEVERITY_ERROR,
            code: NormalizedMessage_1.NormalizedMessage.ERROR_CODE_INTERNAL,
            content: typeof error.message === 'string' ? error.message : error.toString(),
            stack: typeof error.stack === 'string' ? error.stack : undefined,
            file: '[internal]'
        });
    };
    return createNormalizedMessageFromInternalError;
};
//# sourceMappingURL=NormalizedMessageFactories.js.map