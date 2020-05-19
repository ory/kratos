"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
const fs = require("fs");
const path = require("path");
const FilesRegister_1 = require("./FilesRegister");
const FilesWatcher_1 = require("./FilesWatcher");
const linterConfigHelpers_1 = require("./linterConfigHelpers");
const WorkSet_1 = require("./WorkSet");
const NormalizedMessage_1 = require("./NormalizedMessage");
const resolution_1 = require("./resolution");
const minimatch = require("minimatch");
const VueProgram_1 = require("./VueProgram");
const FsHelper_1 = require("./FsHelper");
class IncrementalChecker {
    constructor({ typescript, context, programConfigFile, compilerOptions, createNormalizedMessageFromDiagnostic, linterConfigFile, linterAutoFix, createNormalizedMessageFromRuleFailure, eslinter, watchPaths, workNumber = 0, workDivision = 1, vue, checkSyntacticErrors = false, resolveModuleName, resolveTypeReferenceDirective }) {
        // it's shared between compilations
        this.linterConfigs = {};
        this.files = new FilesRegister_1.FilesRegister(() => ({
            // data shape
            source: undefined,
            linted: false,
            tslints: [],
            eslints: []
        }));
        // Use empty array of exclusions in general to avoid having
        // to check of its existence later on.
        this.linterExclusions = [];
        this.getLinterConfig = linterConfigHelpers_1.makeGetLinterConfig(this.linterConfigs, this.linterExclusions, this.context);
        this.typescript = typescript;
        this.context = context;
        this.programConfigFile = programConfigFile;
        this.compilerOptions = compilerOptions;
        this.createNormalizedMessageFromDiagnostic = createNormalizedMessageFromDiagnostic;
        this.linterConfigFile = linterConfigFile;
        this.linterAutoFix = linterAutoFix;
        this.createNormalizedMessageFromRuleFailure = createNormalizedMessageFromRuleFailure;
        this.eslinter = eslinter;
        this.watchPaths = watchPaths;
        this.workNumber = workNumber;
        this.workDivision = workDivision;
        this.vue = vue;
        this.checkSyntacticErrors = checkSyntacticErrors;
        this.resolveModuleName = resolveModuleName;
        this.resolveTypeReferenceDirective = resolveTypeReferenceDirective;
        this.hasFixedConfig = typeof this.linterConfigFile === 'string';
    }
    static loadProgramConfig(typescript, configFile, compilerOptions) {
        const tsconfig = typescript.readConfigFile(configFile, typescript.sys.readFile).config;
        tsconfig.compilerOptions = tsconfig.compilerOptions || {};
        tsconfig.compilerOptions = Object.assign({}, tsconfig.compilerOptions, compilerOptions);
        const parsed = typescript.parseJsonConfigFileContent(tsconfig, typescript.sys, path.dirname(configFile));
        return parsed;
    }
    static createProgram(typescript, programConfig, files, watcher, oldProgram, userResolveModuleName, userResolveTypeReferenceDirective) {
        const host = typescript.createCompilerHost(programConfig.options);
        const realGetSourceFile = host.getSourceFile;
        const { resolveModuleName, resolveTypeReferenceDirective } = resolution_1.makeResolutionFunctions(userResolveModuleName, userResolveTypeReferenceDirective);
        host.resolveModuleNames = (moduleNames, containingFile) => {
            return moduleNames.map(moduleName => {
                return resolveModuleName(typescript, moduleName, containingFile, programConfig.options, host).resolvedModule;
            });
        };
        host.resolveTypeReferenceDirectives = (typeDirectiveNames, containingFile) => {
            return typeDirectiveNames.map(typeDirectiveName => {
                return resolveTypeReferenceDirective(typescript, typeDirectiveName, containingFile, programConfig.options, host).resolvedTypeReferenceDirective;
            });
        };
        host.getSourceFile = (filePath, languageVersion, onError) => {
            // first check if watcher is watching file - if not - check it's mtime
            if (!watcher.isWatchingFile(filePath)) {
                try {
                    const stats = fs.statSync(filePath);
                    files.setMtime(filePath, stats.mtime.valueOf());
                }
                catch (e) {
                    // probably file does not exists
                    files.remove(filePath);
                }
            }
            // get source file only if there is no source in files register
            if (!files.has(filePath) || !files.getData(filePath).source) {
                files.mutateData(filePath, data => {
                    data.source = realGetSourceFile(filePath, languageVersion, onError);
                });
            }
            return files.getData(filePath).source;
        };
        return typescript.createProgram(programConfig.fileNames, programConfig.options, host, oldProgram // re-use old program
        );
    }
    createLinter(program) {
        // tslint:disable-next-line:no-implicit-dependencies
        const tslint = require('tslint');
        return new tslint.Linter({ fix: this.linterAutoFix }, program);
    }
    hasLinter() {
        return !!this.linter;
    }
    hasEsLinter() {
        return this.eslinter !== undefined;
    }
    static isFileExcluded(filePath, linterExclusions) {
        return (filePath.endsWith('.d.ts') ||
            linterExclusions.some(matcher => matcher.match(filePath)));
    }
    nextIteration() {
        if (!this.watcher) {
            const watchExtensions = this.vue.enabled
                ? ['.ts', '.tsx', '.vue']
                : ['.ts', '.tsx'];
            this.watcher = new FilesWatcher_1.FilesWatcher(this.watchPaths, watchExtensions);
            // connect watcher with register
            this.watcher.on('change', (filePath, stats) => {
                this.files.setMtime(filePath, stats.mtime.valueOf());
            });
            this.watcher.on('unlink', (filePath) => {
                this.files.remove(filePath);
            });
            this.watcher.watch();
        }
        if (!this.linterConfig && this.hasFixedConfig) {
            this.linterConfig = linterConfigHelpers_1.loadLinterConfig(this.linterConfigFile);
            if (this.linterConfig.linterOptions &&
                this.linterConfig.linterOptions.exclude) {
                // Pre-build minimatch patterns to avoid additional overhead later on.
                // Note: Resolving the path is required to properly match against the full file paths,
                // and also deals with potential cross-platform problems regarding path separators.
                this.linterExclusions = this.linterConfig.linterOptions.exclude.map(pattern => new minimatch.Minimatch(path.resolve(pattern)));
            }
        }
        this.program = this.vue.enabled
            ? this.loadVueProgram(this.vue)
            : this.loadDefaultProgram();
        if (this.linterConfigFile) {
            this.linter = this.createLinter(this.program);
        }
    }
    loadVueProgram(vueOptions) {
        this.programConfig =
            this.programConfig ||
                VueProgram_1.VueProgram.loadProgramConfig(this.typescript, this.programConfigFile, this.compilerOptions);
        return VueProgram_1.VueProgram.createProgram(this.typescript, this.programConfig, path.dirname(this.programConfigFile), this.files, this.watcher, this.program, this.resolveModuleName, this.resolveTypeReferenceDirective, vueOptions);
    }
    loadDefaultProgram() {
        this.programConfig =
            this.programConfig ||
                IncrementalChecker.loadProgramConfig(this.typescript, this.programConfigFile, this.compilerOptions);
        return IncrementalChecker.createProgram(this.typescript, this.programConfig, this.files, this.watcher, this.program, this.resolveModuleName, this.resolveTypeReferenceDirective);
    }
    getDiagnostics(cancellationToken) {
        const { program } = this;
        if (!program) {
            throw new Error('Invoked called before program initialized');
        }
        const diagnostics = [];
        // select files to check (it's semantic check - we have to include all files :/)
        const filesToCheck = program.getSourceFiles();
        // calculate subset of work to do
        const workSet = new WorkSet_1.WorkSet(filesToCheck, this.workNumber, this.workDivision);
        // check given work set
        workSet.forEach(sourceFile => {
            if (cancellationToken) {
                cancellationToken.throwIfCancellationRequested();
            }
            const diagnosticsToRegister = this
                .checkSyntacticErrors
                ? program
                    .getSemanticDiagnostics(sourceFile, cancellationToken)
                    .concat(program.getSyntacticDiagnostics(sourceFile, cancellationToken))
                : program.getSemanticDiagnostics(sourceFile, cancellationToken);
            diagnostics.push(...diagnosticsToRegister);
        });
        // normalize and deduplicate diagnostics
        return Promise.resolve(NormalizedMessage_1.NormalizedMessage.deduplicate(diagnostics.map(this.createNormalizedMessageFromDiagnostic)));
    }
    getLints(cancellationToken) {
        const { linter } = this;
        if (!linter) {
            throw new Error('Cannot get lints - checker has no linter.');
        }
        // select files to lint
        const filesToLint = this.files
            .keys()
            .filter(filePath => !this.files.getData(filePath).linted &&
            !IncrementalChecker.isFileExcluded(filePath, this.linterExclusions));
        // calculate subset of work to do
        const workSet = new WorkSet_1.WorkSet(filesToLint, this.workNumber, this.workDivision);
        // lint given work set
        workSet.forEach(fileName => {
            cancellationToken.throwIfCancellationRequested();
            const config = this.hasFixedConfig
                ? this.linterConfig
                : this.getLinterConfig(fileName);
            if (!config) {
                return;
            }
            try {
                // Assertion: `.lint` second parameter can be undefined
                linter.lint(fileName, undefined, config);
            }
            catch (e) {
                FsHelper_1.throwIfIsInvalidSourceFileError(fileName, e);
            }
        });
        // set lints in files register
        linter.getResult().failures.forEach(lint => {
            const filePath = lint.getFileName();
            this.files.mutateData(filePath, data => {
                data.linted = true;
                data.tslints.push(lint);
            });
        });
        // set all files as linted
        this.files.keys().forEach(filePath => {
            this.files.mutateData(filePath, data => {
                data.linted = true;
            });
        });
        // get all lints
        const lints = this.files
            .keys()
            .reduce((innerLints, filePath) => innerLints.concat(this.files.getData(filePath).tslints), []);
        // normalize and deduplicate lints
        return NormalizedMessage_1.NormalizedMessage.deduplicate(lints.map(this.createNormalizedMessageFromRuleFailure));
    }
    getEsLints(cancellationToken) {
        // select files to lint
        const filesToLint = this.files
            .keys()
            .filter(filePath => !this.files.getData(filePath).linted &&
            !IncrementalChecker.isFileExcluded(filePath, this.linterExclusions));
        // calculate subset of work to do
        const workSet = new WorkSet_1.WorkSet(filesToLint, this.workNumber, this.workDivision);
        // lint given work set
        const currentEsLintErrors = new Map();
        workSet.forEach(fileName => {
            cancellationToken.throwIfCancellationRequested();
            const lints = this.eslinter.getLints(fileName);
            if (lints !== undefined) {
                currentEsLintErrors.set(fileName, lints);
            }
        });
        // set lints in files register
        for (const [filePath, lint] of currentEsLintErrors) {
            this.files.mutateData(filePath, data => {
                data.linted = true;
                data.eslints.push(lint);
            });
        }
        // set all files as linted
        this.files.keys().forEach(filePath => {
            this.files.mutateData(filePath, data => {
                data.linted = true;
            });
        });
        // get all lints
        const allEsLintReports = this.files
            .keys()
            .reduce((innerLints, filePath) => innerLints.concat(this.files.getData(filePath).eslints), []);
        return this.eslinter.getFormattedLints(allEsLintReports);
    }
}
exports.IncrementalChecker = IncrementalChecker;
//# sourceMappingURL=IncrementalChecker.js.map