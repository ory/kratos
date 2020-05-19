"use strict";
const path = require("path");
const process = require("process");
const childProcess = require("child_process");
const worker_rpc_1 = require("worker-rpc");
const semver = require("semver");
const chalk_1 = require("chalk");
const micromatch = require("micromatch");
const os = require("os");
const CancellationToken_1 = require("./CancellationToken");
const NormalizedMessage_1 = require("./NormalizedMessage");
const defaultFormatter_1 = require("./formatter/defaultFormatter");
const codeframeFormatter_1 = require("./formatter/codeframeFormatter");
const FsHelper_1 = require("./FsHelper");
const hooks_1 = require("./hooks");
const RpcTypes_1 = require("./RpcTypes");
const checkerPluginName = 'fork-ts-checker-webpack-plugin';
/**
 * ForkTsCheckerWebpackPlugin
 * Runs typescript type checker and linter (tslint) on separate process.
 * This speed-ups build a lot.
 *
 * Options description in README.md
 */
class ForkTsCheckerWebpackPlugin {
    constructor(options) {
        this.tslint = false;
        this.eslint = false;
        this.eslintOptions = {};
        this.tslintAutoFix = false;
        this.tsconfigPath = undefined;
        this.tslintPath = undefined;
        this.watchPaths = [];
        this.compiler = undefined;
        this.started = undefined;
        this.elapsed = undefined;
        this.cancellationToken = undefined;
        this.isWatching = false;
        this.checkDone = false;
        this.compilationDone = false;
        this.diagnostics = [];
        this.lints = [];
        this.eslintVersion = undefined;
        this.startAt = 0;
        this.nodeArgs = [];
        options = options || {};
        this.options = Object.assign({}, options);
        this.watch =
            typeof options.watch === 'string' ? [options.watch] : options.watch || [];
        this.ignoreDiagnostics = options.ignoreDiagnostics || [];
        this.ignoreLints = options.ignoreLints || [];
        this.ignoreLintWarnings = options.ignoreLintWarnings === true;
        this.reportFiles = options.reportFiles || [];
        this.logger = options.logger || console;
        this.silent = options.silent === true; // default false
        this.async = options.async !== false; // default true
        this.checkSyntacticErrors = options.checkSyntacticErrors === true; // default false
        this.resolveModuleNameModule = options.resolveModuleNameModule;
        this.resolveTypeReferenceDirectiveModule =
            options.resolveTypeReferenceDirectiveModule;
        this.workersNumber = options.workers || ForkTsCheckerWebpackPlugin.ONE_CPU;
        this.memoryLimit =
            options.memoryLimit || ForkTsCheckerWebpackPlugin.DEFAULT_MEMORY_LIMIT;
        this.useColors = options.colors !== false; // default true
        this.colors = new chalk_1.default.constructor({ enabled: this.useColors });
        this.formatter =
            options.formatter && typeof options.formatter === 'function'
                ? options.formatter
                : ForkTsCheckerWebpackPlugin.createFormatter(options.formatter || 'default', options.formatterOptions || {});
        this.emitCallback = this.createNoopEmitCallback();
        this.doneCallback = this.createDoneCallback();
        const { typescript, typescriptPath, typescriptVersion, tsconfig, compilerOptions } = this.validateTypeScript(options);
        this.typescript = typescript;
        this.typescriptPath = typescriptPath;
        this.typescriptVersion = typescriptVersion;
        this.tsconfig = tsconfig;
        this.compilerOptions = compilerOptions;
        if (options.eslint === true) {
            const { eslintVersion, eslintOptions } = this.validateEslint(options);
            this.eslint = true;
            this.eslintVersion = eslintVersion;
            this.eslintOptions = eslintOptions;
        }
        else {
            const { tslint, tslintVersion, tslintAutoFix } = this.validateTslint(options);
            this.tslint = tslint;
            this.tslintVersion = tslintVersion;
            this.tslintAutoFix = tslintAutoFix;
        }
        this.vue = ForkTsCheckerWebpackPlugin.prepareVueOptions(options.vue);
        this.useTypescriptIncrementalApi =
            options.useTypescriptIncrementalApi === undefined
                ? semver.gte(this.typescriptVersion, '3.0.0') && !this.vue.enabled
                : options.useTypescriptIncrementalApi;
        this.measureTime = options.measureCompilationTime === true;
        if (this.measureTime) {
            // Node 8+ only
            this.performance = require('perf_hooks').performance;
        }
    }
    static getCompilerHooks(compiler) {
        return hooks_1.getForkTsCheckerWebpackPluginHooks(compiler);
    }
    validateTypeScript(options) {
        const typescriptPath = options.typescript || require.resolve('typescript');
        const tsconfig = options.tsconfig || './tsconfig.json';
        const compilerOptions = typeof options.compilerOptions === 'object'
            ? options.compilerOptions
            : {};
        let typescript, typescriptVersion;
        try {
            typescript = require(typescriptPath);
            typescriptVersion = typescript.version;
        }
        catch (_ignored) {
            throw new Error('When you use this plugin you must install `typescript`.');
        }
        if (semver.lt(typescriptVersion, '2.1.0')) {
            throw new Error(`Cannot use current typescript version of ${typescriptVersion}, the minimum required version is 2.1.0`);
        }
        return {
            typescriptPath,
            typescript,
            typescriptVersion,
            tsconfig,
            compilerOptions
        };
    }
    validateTslint(options) {
        const tslint = options.tslint
            ? options.tslint === true
                ? true
                : options.tslint
            : undefined;
        let tslintAutoFix, tslintVersion;
        try {
            tslintAutoFix = options.tslintAutoFix || false;
            tslintVersion = tslint
                ? // tslint:disable-next-line:no-implicit-dependencies
                    require('tslint').Linter.VERSION
                : undefined;
        }
        catch (_ignored) {
            throw new Error('When you use `tslint` option, make sure to install `tslint`.');
        }
        if (tslintVersion && semver.lt(tslintVersion, '4.0.0')) {
            throw new Error(`Cannot use current tslint version of ${tslintVersion}, the minimum required version is 4.0.0`);
        }
        return { tslint, tslintAutoFix, tslintVersion };
    }
    validateEslint(options) {
        let eslintVersion;
        const eslintOptions = typeof options.eslintOptions === 'object' ? options.eslintOptions : {};
        try {
            eslintVersion = require('eslint').Linter.version;
        }
        catch (_ignored) {
            throw new Error('When you use `eslint` option, make sure to install `eslint`.');
        }
        return { eslintVersion, eslintOptions };
    }
    static prepareVueOptions(vueOptions) {
        const defaultVueOptions = {
            compiler: 'vue-template-compiler',
            enabled: false
        };
        if (typeof vueOptions === 'boolean') {
            return Object.assign(defaultVueOptions, { enabled: vueOptions });
        }
        else if (typeof vueOptions === 'object' && vueOptions !== null) {
            return Object.assign(defaultVueOptions, vueOptions);
        }
        else {
            return defaultVueOptions;
        }
    }
    static createFormatter(type, options) {
        switch (type) {
            case 'default':
                return defaultFormatter_1.createDefaultFormatter();
            case 'codeframe':
                return codeframeFormatter_1.createCodeframeFormatter(options);
            default:
                throw new Error('Unknown "' + type + '" formatter. Available are: default, codeframe.');
        }
    }
    apply(compiler) {
        this.compiler = compiler;
        this.tsconfigPath = this.computeContextPath(this.tsconfig);
        this.tslintPath =
            typeof this.tslint === 'string'
                ? this.computeContextPath(this.tslint)
                : undefined;
        this.watchPaths = this.watch.map(this.computeContextPath.bind(this));
        // validate config
        const tsconfigOk = FsHelper_1.fileExistsSync(this.tsconfigPath);
        const tslintOk = !this.tslintPath || FsHelper_1.fileExistsSync(this.tslintPath);
        if (this.useTypescriptIncrementalApi && this.workersNumber !== 1) {
            throw new Error('Using typescript incremental compilation API ' +
                'is currently only allowed with a single worker.');
        }
        // validate logger
        if (this.logger) {
            if (!this.logger.error || !this.logger.warn || !this.logger.info) {
                throw new Error("Invalid logger object - doesn't provide `error`, `warn` or `info` method.");
            }
        }
        if (tsconfigOk && tslintOk) {
            this.pluginStart();
            this.pluginStop();
            this.pluginCompile();
            this.pluginEmit();
            this.pluginDone();
        }
        else {
            if (!tsconfigOk) {
                throw new Error('Cannot find "' +
                    this.tsconfigPath +
                    '" file. Please check webpack and ForkTsCheckerWebpackPlugin configuration. \n' +
                    'Possible errors: \n' +
                    '  - wrong `context` directory in webpack configuration' +
                    ' (if `tsconfig` is not set or is a relative path in fork plugin configuration)\n' +
                    '  - wrong `tsconfig` path in fork plugin configuration' +
                    ' (should be a relative or absolute path)');
            }
            if (!tslintOk) {
                throw new Error('Cannot find "' +
                    this.tslintPath +
                    '" file. Please check webpack and ForkTsCheckerWebpackPlugin configuration. \n' +
                    'Possible errors: \n' +
                    '  - wrong `context` directory in webpack configuration' +
                    ' (if `tslint` is not set or is a relative path in fork plugin configuration)\n' +
                    '  - wrong `tslint` path in fork plugin configuration' +
                    ' (should be a relative or absolute path)\n' +
                    '  - `tslint` path is not set to false in fork plugin configuration' +
                    ' (if you want to disable tslint support)');
            }
        }
    }
    computeContextPath(filePath) {
        return path.isAbsolute(filePath)
            ? filePath
            : path.resolve(this.compiler.options.context, filePath);
    }
    pluginStart() {
        const run = (_compiler, callback) => {
            this.isWatching = false;
            callback();
        };
        const watchRun = (_compiler, callback) => {
            this.isWatching = true;
            callback();
        };
        if ('hooks' in this.compiler) {
            // webpack 4+
            this.compiler.hooks.run.tapAsync(checkerPluginName, run);
            this.compiler.hooks.watchRun.tapAsync(checkerPluginName, watchRun);
        }
        else {
            // webpack 2 / 3
            this.compiler.plugin('run', run);
            this.compiler.plugin('watch-run', watchRun);
        }
    }
    pluginStop() {
        const watchClose = () => {
            this.killService();
        };
        const done = (_stats) => {
            if (!this.isWatching) {
                this.killService();
            }
        };
        if ('hooks' in this.compiler) {
            // webpack 4+
            this.compiler.hooks.watchClose.tap(checkerPluginName, watchClose);
            this.compiler.hooks.done.tap(checkerPluginName, done);
        }
        else {
            // webpack 2 / 3
            this.compiler.plugin('watch-close', watchClose);
            this.compiler.plugin('done', done);
        }
        process.on('exit', () => {
            this.killService();
        });
    }
    pluginCompile() {
        if ('hooks' in this.compiler) {
            // webpack 4+
            const forkTsCheckerHooks = ForkTsCheckerWebpackPlugin.getCompilerHooks(this.compiler);
            this.compiler.hooks.compile.tap(checkerPluginName, () => {
                this.compilationDone = false;
                forkTsCheckerHooks.serviceBeforeStart.callAsync(() => {
                    if (this.cancellationToken) {
                        // request cancellation if there is not finished job
                        this.cancellationToken.requestCancellation();
                        forkTsCheckerHooks.cancel.call(this.cancellationToken);
                    }
                    this.checkDone = false;
                    this.started = process.hrtime();
                    // create new token for current job
                    this.cancellationToken = new CancellationToken_1.CancellationToken(this.typescript);
                    if (!this.service || !this.service.connected) {
                        this.spawnService();
                    }
                    try {
                        if (this.measureTime) {
                            this.startAt = this.performance.now();
                        }
                        this.serviceRpc.rpc(RpcTypes_1.RUN, this.cancellationToken.toJSON()).then(result => {
                            if (result) {
                                this.handleServiceMessage(result);
                            }
                        });
                    }
                    catch (error) {
                        if (!this.silent && this.logger) {
                            this.logger.error(this.colors.red('Cannot start checker service: ' +
                                (error ? error.toString() : 'Unknown error')));
                        }
                        forkTsCheckerHooks.serviceStartError.call(error);
                    }
                });
            });
        }
        else {
            // webpack 2 / 3
            this.compiler.plugin('compile', () => {
                this.compilationDone = false;
                this.compiler.applyPluginsAsync(hooks_1.legacyHookMap.serviceBeforeStart, () => {
                    if (this.cancellationToken) {
                        // request cancellation if there is not finished job
                        this.cancellationToken.requestCancellation();
                        this.compiler.applyPlugins(hooks_1.legacyHookMap.cancel, this.cancellationToken);
                    }
                    this.checkDone = false;
                    this.started = process.hrtime();
                    // create new token for current job
                    this.cancellationToken = new CancellationToken_1.CancellationToken(this.typescript, undefined, undefined);
                    if (!this.service || !this.service.connected) {
                        this.spawnService();
                    }
                    try {
                        this.serviceRpc.rpc(RpcTypes_1.RUN, this.cancellationToken.toJSON()).then(result => {
                            if (result) {
                                this.handleServiceMessage(result);
                            }
                        });
                    }
                    catch (error) {
                        if (!this.silent && this.logger) {
                            this.logger.error(this.colors.red('Cannot start checker service: ' +
                                (error ? error.toString() : 'Unknown error')));
                        }
                        this.compiler.applyPlugins(hooks_1.legacyHookMap.serviceStartError, error);
                    }
                });
            });
        }
    }
    pluginEmit() {
        const emit = (compilation, callback) => {
            if (this.isWatching && this.async) {
                callback();
                return;
            }
            this.emitCallback = this.createEmitCallback(compilation, callback);
            if (this.checkDone) {
                this.emitCallback();
            }
            this.compilationDone = true;
        };
        if ('hooks' in this.compiler) {
            // webpack 4+
            this.compiler.hooks.emit.tapAsync(checkerPluginName, emit);
        }
        else {
            // webpack 2 / 3
            this.compiler.plugin('emit', emit);
        }
    }
    pluginDone() {
        if ('hooks' in this.compiler) {
            // webpack 4+
            const forkTsCheckerHooks = ForkTsCheckerWebpackPlugin.getCompilerHooks(this.compiler);
            this.compiler.hooks.done.tap(checkerPluginName, (_stats) => {
                if (!this.isWatching || !this.async) {
                    return;
                }
                if (this.checkDone) {
                    this.doneCallback();
                }
                else {
                    if (this.compiler) {
                        forkTsCheckerHooks.waiting.call(this.tslint !== undefined);
                    }
                    if (!this.silent && this.logger) {
                        this.logger.info(this.tslint
                            ? 'Type checking and linting in progress...'
                            : 'Type checking in progress...');
                    }
                }
                this.compilationDone = true;
            });
        }
        else {
            // webpack 2 / 3
            this.compiler.plugin('done', () => {
                if (!this.isWatching || !this.async) {
                    return;
                }
                if (this.checkDone) {
                    this.doneCallback();
                }
                else {
                    if (this.compiler) {
                        this.compiler.applyPlugins(hooks_1.legacyHookMap.waiting, this.tslint !== undefined);
                    }
                    if (!this.silent && this.logger) {
                        this.logger.info(this.tslint
                            ? 'Type checking and linting in progress...'
                            : 'Type checking in progress...');
                    }
                }
                this.compilationDone = true;
            });
        }
    }
    spawnService() {
        const env = Object.assign({}, process.env, { TYPESCRIPT_PATH: this.typescriptPath, TSCONFIG: this.tsconfigPath, COMPILER_OPTIONS: JSON.stringify(this.compilerOptions), TSLINT: this.tslintPath || (this.tslint ? 'true' : ''), CONTEXT: this.compiler.options.context, TSLINTAUTOFIX: String(this.tslintAutoFix), ESLINT: String(this.eslint), ESLINT_OPTIONS: JSON.stringify(this.eslintOptions), WATCH: this.isWatching ? this.watchPaths.join('|') : '', WORK_DIVISION: String(Math.max(1, this.workersNumber)), MEMORY_LIMIT: String(this.memoryLimit), CHECK_SYNTACTIC_ERRORS: String(this.checkSyntacticErrors), USE_INCREMENTAL_API: String(this.useTypescriptIncrementalApi === true), VUE: JSON.stringify(this.vue) });
        if (typeof this.resolveModuleNameModule !== 'undefined') {
            env.RESOLVE_MODULE_NAME = this.resolveModuleNameModule;
        }
        else {
            delete env.RESOLVE_MODULE_NAME;
        }
        if (typeof this.resolveTypeReferenceDirectiveModule !== 'undefined') {
            env.RESOLVE_TYPE_REFERENCE_DIRECTIVE = this.resolveTypeReferenceDirectiveModule;
        }
        else {
            delete env.RESOLVE_TYPE_REFERENCE_DIRECTIVE;
        }
        this.service = childProcess.fork(path.resolve(__dirname, this.workersNumber > 1 ? './cluster.js' : './service.js'), [], {
            env,
            execArgv: (this.workersNumber > 1
                ? []
                : ['--max-old-space-size=' + this.memoryLimit]).concat(this.nodeArgs),
            stdio: ['inherit', 'inherit', 'inherit', 'ipc']
        });
        this.serviceRpc = new worker_rpc_1.RpcProvider(message => this.service.send(message));
        this.service.on('message', message => this.serviceRpc.dispatch(message));
        if ('hooks' in this.compiler) {
            // webpack 4+
            const forkTsCheckerHooks = ForkTsCheckerWebpackPlugin.getCompilerHooks(this.compiler);
            forkTsCheckerHooks.serviceStart.call(this.tsconfigPath, this.tslintPath, this.watchPaths, this.workersNumber, this.memoryLimit);
        }
        else {
            // webpack 2 / 3
            this.compiler.applyPlugins(hooks_1.legacyHookMap.serviceStart, this.tsconfigPath, this.tslintPath, this.watchPaths, this.workersNumber, this.memoryLimit);
        }
        if (!this.silent && this.logger) {
            this.logger.info('Starting type checking' +
                (this.tslint ? ' and linting' : '') +
                ' service...');
            this.logger.info('Using ' +
                this.colors.bold(this.workersNumber === 1
                    ? '1 worker'
                    : this.workersNumber + ' workers') +
                ' with ' +
                this.colors.bold(this.memoryLimit + 'MB') +
                ' memory limit');
            if (this.watchPaths.length && this.isWatching) {
                this.logger.info('Watching:' +
                    (this.watchPaths.length > 1 ? '\n' : ' ') +
                    this.watchPaths.map(wpath => this.colors.grey(wpath)).join('\n'));
            }
        }
        this.service.on('exit', (code, signal) => this.handleServiceExit(code, signal));
    }
    killService() {
        if (!this.service) {
            return;
        }
        try {
            if (this.cancellationToken) {
                this.cancellationToken.cleanupCancellation();
            }
            this.service.kill();
            this.service = undefined;
            this.serviceRpc = undefined;
        }
        catch (e) {
            if (this.logger && !this.silent) {
                this.logger.error(e);
            }
        }
    }
    handleServiceMessage(message) {
        if (this.measureTime) {
            const delta = this.performance.now() - this.startAt;
            const deltaRounded = Math.round(delta * 100) / 100;
            this.logger.info(`Compilation took: ${deltaRounded} ms.`);
        }
        if (this.cancellationToken) {
            this.cancellationToken.cleanupCancellation();
            // job is done - nothing to cancel
            this.cancellationToken = undefined;
        }
        this.checkDone = true;
        this.elapsed = process.hrtime(this.started);
        this.diagnostics = message.diagnostics.map(NormalizedMessage_1.NormalizedMessage.createFromJSON);
        this.lints = message.lints.map(NormalizedMessage_1.NormalizedMessage.createFromJSON);
        if (this.ignoreDiagnostics.length) {
            this.diagnostics = this.diagnostics.filter(diagnostic => !this.ignoreDiagnostics.includes(parseInt(diagnostic.code, 10)));
        }
        if (this.ignoreLints.length) {
            this.lints = this.lints.filter(lint => !this.ignoreLints.includes(lint.code));
        }
        if (this.reportFiles.length) {
            const reportFilesPredicate = (diagnostic) => {
                if (diagnostic.file) {
                    const relativeFileName = path.relative(this.compiler.options.context, diagnostic.file);
                    const matchResult = micromatch([relativeFileName], this.reportFiles);
                    if (matchResult.length === 0) {
                        return false;
                    }
                }
                return true;
            };
            this.diagnostics = this.diagnostics.filter(reportFilesPredicate);
            this.lints = this.lints.filter(reportFilesPredicate);
        }
        if ('hooks' in this.compiler) {
            // webpack 4+
            const forkTsCheckerHooks = ForkTsCheckerWebpackPlugin.getCompilerHooks(this.compiler);
            forkTsCheckerHooks.receive.call(this.diagnostics, this.lints);
        }
        else {
            // webpack 2 / 3
            this.compiler.applyPlugins(hooks_1.legacyHookMap.receive, this.diagnostics, this.lints);
        }
        if (this.compilationDone) {
            this.isWatching && this.async ? this.doneCallback() : this.emitCallback();
        }
    }
    handleServiceExit(_code, signal) {
        if (signal !== 'SIGABRT') {
            return;
        }
        // probably out of memory :/
        if (this.compiler) {
            if ('hooks' in this.compiler) {
                // webpack 4+
                const forkTsCheckerHooks = ForkTsCheckerWebpackPlugin.getCompilerHooks(this.compiler);
                forkTsCheckerHooks.serviceOutOfMemory.call();
            }
            else {
                // webpack 2 / 3
                this.compiler.applyPlugins(hooks_1.legacyHookMap.serviceOutOfMemory);
            }
        }
        if (!this.silent && this.logger) {
            this.logger.error(this.colors.red('Type checking and linting aborted - probably out of memory. ' +
                'Check `memoryLimit` option in ForkTsCheckerWebpackPlugin configuration.'));
        }
    }
    createEmitCallback(compilation, callback) {
        return function emitCallback() {
            if (!this.elapsed) {
                throw new Error('Execution order error');
            }
            const elapsed = Math.round(this.elapsed[0] * 1e9 + this.elapsed[1]);
            if ('hooks' in this.compiler) {
                // webpack 4+
                const forkTsCheckerHooks = ForkTsCheckerWebpackPlugin.getCompilerHooks(this.compiler);
                forkTsCheckerHooks.emit.call(this.diagnostics, this.lints, elapsed);
            }
            else {
                // webpack 2 / 3
                this.compiler.applyPlugins(hooks_1.legacyHookMap.emit, this.diagnostics, this.lints, elapsed);
            }
            this.diagnostics.concat(this.lints).forEach(message => {
                // webpack message format
                const formatted = {
                    rawMessage: message.severity.toUpperCase() +
                        ' ' +
                        message.getFormattedCode() +
                        ': ' +
                        message.content,
                    message: this.formatter(message, this.useColors),
                    location: {
                        line: message.line,
                        character: message.character
                    },
                    file: message.file
                };
                if (message.isWarningSeverity()) {
                    if (!this.ignoreLintWarnings) {
                        compilation.warnings.push(formatted);
                    }
                }
                else {
                    compilation.errors.push(formatted);
                }
            });
            callback();
        };
    }
    createNoopEmitCallback() {
        // tslint:disable-next-line:no-empty
        return function noopEmitCallback() { };
    }
    printLoggerMessage(message, formattedMessage) {
        if (message.isWarningSeverity()) {
            if (this.ignoreLintWarnings) {
                return;
            }
            this.logger.warn(formattedMessage);
        }
        else {
            this.logger.error(formattedMessage);
        }
    }
    createDoneCallback() {
        return function doneCallback() {
            if (!this.elapsed) {
                throw new Error('Execution order error');
            }
            const elapsed = Math.round(this.elapsed[0] * 1e9 + this.elapsed[1]);
            if (this.compiler) {
                if ('hooks' in this.compiler) {
                    // webpack 4+
                    const forkTsCheckerHooks = ForkTsCheckerWebpackPlugin.getCompilerHooks(this.compiler);
                    forkTsCheckerHooks.done.call(this.diagnostics, this.lints, elapsed);
                }
                else {
                    // webpack 2 / 3
                    this.compiler.applyPlugins(hooks_1.legacyHookMap.done, this.diagnostics, this.lints, elapsed);
                }
            }
            if (!this.silent && this.logger) {
                if (this.diagnostics.length || this.lints.length) {
                    (this.lints || []).concat(this.diagnostics).forEach(message => {
                        const formattedMessage = this.formatter(message, this.useColors);
                        this.printLoggerMessage(message, formattedMessage);
                    });
                }
                if (!this.diagnostics.length) {
                    this.logger.info(this.colors.green('No type errors found'));
                }
                if (this.tslint && !this.lints.length) {
                    this.logger.info(this.colors.green('No lint errors found'));
                }
                this.logger.info('Version: typescript ' +
                    this.colors.bold(this.typescriptVersion) +
                    (this.eslint
                        ? ', eslint ' + this.colors.bold(this.eslintVersion)
                        : this.tslint
                            ? ', tslint ' + this.colors.bold(this.tslintVersion)
                            : ''));
                this.logger.info('Time: ' +
                    this.colors.bold(Math.round(elapsed / 1e6).toString()) +
                    'ms');
            }
        };
    }
}
ForkTsCheckerWebpackPlugin.DEFAULT_MEMORY_LIMIT = 2048;
ForkTsCheckerWebpackPlugin.ONE_CPU = 1;
ForkTsCheckerWebpackPlugin.ALL_CPUS = os.cpus && os.cpus() ? os.cpus().length : 1;
ForkTsCheckerWebpackPlugin.ONE_CPU_FREE = Math.max(1, ForkTsCheckerWebpackPlugin.ALL_CPUS - 1);
ForkTsCheckerWebpackPlugin.TWO_CPUS_FREE = Math.max(1, ForkTsCheckerWebpackPlugin.ALL_CPUS - 2);
module.exports = ForkTsCheckerWebpackPlugin;
//# sourceMappingURL=index.js.map