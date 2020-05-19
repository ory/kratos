"use strict";
var __awaiter = (this && this.__awaiter) || function (thisArg, _arguments, P, generator) {
    return new (P || (P = Promise))(function (resolve, reject) {
        function fulfilled(value) { try { step(generator.next(value)); } catch (e) { reject(e); } }
        function rejected(value) { try { step(generator["throw"](value)); } catch (e) { reject(e); } }
        function step(result) { result.done ? resolve(result.value) : new P(function (resolve) { resolve(result.value); }).then(fulfilled, rejected); }
        step((generator = generator.apply(thisArg, _arguments || [])).next());
    });
};
Object.defineProperty(exports, "__esModule", { value: true });
const process = require("process");
const IncrementalChecker_1 = require("./IncrementalChecker");
const CancellationToken_1 = require("./CancellationToken");
const ApiIncrementalChecker_1 = require("./ApiIncrementalChecker");
const NormalizedMessageFactories_1 = require("./NormalizedMessageFactories");
const worker_rpc_1 = require("worker-rpc");
const RpcTypes_1 = require("./RpcTypes");
const patchTypescript_1 = require("./patchTypescript");
const createEslinter_1 = require("./createEslinter");
const rpc = new worker_rpc_1.RpcProvider(message => {
    try {
        process.send(message, undefined, undefined, error => {
            if (error) {
                process.exit();
            }
        });
    }
    catch (e) {
        // channel closed...
        process.exit();
    }
});
process.on('message', message => rpc.dispatch(message));
const typescript = require(process.env.TYPESCRIPT_PATH);
const patchConfig = {
    skipGetSyntacticDiagnostics: process.env.USE_INCREMENTAL_API === 'true' &&
        process.env.CHECK_SYNTACTIC_ERRORS !== 'true'
};
patchTypescript_1.patchTypescript(typescript, patchConfig);
// message factories
exports.createNormalizedMessageFromDiagnostic = NormalizedMessageFactories_1.makeCreateNormalizedMessageFromDiagnostic(typescript);
exports.createNormalizedMessageFromRuleFailure = NormalizedMessageFactories_1.makeCreateNormalizedMessageFromRuleFailure();
exports.createNormalizedMessageFromInternalError = NormalizedMessageFactories_1.makeCreateNormalizedMessageFromInternalError();
const resolveModuleName = process.env.RESOLVE_MODULE_NAME
    ? require(process.env.RESOLVE_MODULE_NAME).resolveModuleName
    : undefined;
const resolveTypeReferenceDirective = process.env
    .RESOLVE_TYPE_REFERENCE_DIRECTIVE
    ? require(process.env.RESOLVE_TYPE_REFERENCE_DIRECTIVE)
        .resolveTypeReferenceDirective
    : undefined;
const eslinter = process.env.ESLINT === 'true'
    ? createEslinter_1.createEslinter(JSON.parse(process.env.ESLINT_OPTIONS))
    : undefined;
function createChecker(useIncrementalApi) {
    const apiIncrementalCheckerParams = {
        typescript,
        context: process.env.CONTEXT,
        programConfigFile: process.env.TSCONFIG,
        compilerOptions: JSON.parse(process.env.COMPILER_OPTIONS),
        createNormalizedMessageFromDiagnostic: exports.createNormalizedMessageFromDiagnostic,
        linterConfigFile: process.env.TSLINT === 'true' ? true : process.env.TSLINT || false,
        linterAutoFix: process.env.TSLINTAUTOFIX === 'true',
        createNormalizedMessageFromRuleFailure: exports.createNormalizedMessageFromRuleFailure,
        eslinter,
        checkSyntacticErrors: process.env.CHECK_SYNTACTIC_ERRORS === 'true',
        resolveModuleName,
        resolveTypeReferenceDirective,
        vue: JSON.parse(process.env.VUE)
    };
    if (useIncrementalApi) {
        return new ApiIncrementalChecker_1.ApiIncrementalChecker(apiIncrementalCheckerParams);
    }
    const incrementalCheckerParams = Object.assign({}, apiIncrementalCheckerParams, {
        watchPaths: process.env.WATCH === '' ? [] : process.env.WATCH.split('|'),
        workNumber: parseInt(process.env.WORK_NUMBER, 10) || 0,
        workDivision: parseInt(process.env.WORK_DIVISION, 10) || 1
    });
    return new IncrementalChecker_1.IncrementalChecker(incrementalCheckerParams);
}
const checker = createChecker(process.env.USE_INCREMENTAL_API === 'true');
function run(cancellationToken) {
    return __awaiter(this, void 0, void 0, function* () {
        let diagnostics = [];
        let lints = [];
        try {
            checker.nextIteration();
            diagnostics = yield checker.getDiagnostics(cancellationToken);
            if (checker.hasEsLinter()) {
                lints = checker.getEsLints(cancellationToken);
            }
            else if (checker.hasLinter()) {
                lints = checker.getLints(cancellationToken);
            }
        }
        catch (error) {
            if (error instanceof typescript.OperationCanceledException) {
                return undefined;
            }
            diagnostics.push(exports.createNormalizedMessageFromInternalError(error));
        }
        if (cancellationToken.isCancellationRequested()) {
            return undefined;
        }
        return {
            diagnostics,
            lints
        };
    });
}
rpc.registerRpcHandler(RpcTypes_1.RUN, message => typeof message !== 'undefined'
    ? run(CancellationToken_1.CancellationToken.createFromJSON(typescript, message))
    : undefined);
process.on('SIGINT', () => {
    process.exit();
});
//# sourceMappingURL=service.js.map