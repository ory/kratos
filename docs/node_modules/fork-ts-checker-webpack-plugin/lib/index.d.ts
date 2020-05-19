import { RpcProvider } from 'worker-rpc';
import { NormalizedMessage } from './NormalizedMessage';
import { ForkTsCheckerHooks } from './hooks';
import { VueOptions } from './types/vue-options';
declare namespace ForkTsCheckerWebpackPlugin {
    type Formatter = (message: NormalizedMessage, useColors: boolean) => string;
    interface Logger {
        error(message?: any): void;
        warn(message?: any): void;
        info(message?: any): void;
    }
    interface Options {
        typescript: string;
        tsconfig: string;
        compilerOptions: object;
        tslint: string | true | undefined;
        tslintAutoFix: boolean;
        eslint: boolean;
        /** Options to supply to eslint https://eslint.org/docs/1.0.0/developer-guide/nodejs-api#cliengine */
        eslintOptions: object;
        watch: string | string[];
        async: boolean;
        ignoreDiagnostics: number[];
        ignoreLints: string[];
        ignoreLintWarnings: boolean;
        reportFiles: string[];
        colors: boolean;
        logger: Logger;
        formatter: 'default' | 'codeframe' | Formatter;
        formatterOptions: any;
        silent: boolean;
        checkSyntacticErrors: boolean;
        memoryLimit: number;
        workers: number;
        vue: boolean | Partial<VueOptions>;
        useTypescriptIncrementalApi: boolean;
        measureCompilationTime: boolean;
        resolveModuleNameModule: string;
        resolveTypeReferenceDirectiveModule: string;
    }
}
/**
 * ForkTsCheckerWebpackPlugin
 * Runs typescript type checker and linter (tslint) on separate process.
 * This speed-ups build a lot.
 *
 * Options description in README.md
 */
declare class ForkTsCheckerWebpackPlugin {
    static readonly DEFAULT_MEMORY_LIMIT = 2048;
    static readonly ONE_CPU = 1;
    static readonly ALL_CPUS: number;
    static readonly ONE_CPU_FREE: number;
    static readonly TWO_CPUS_FREE: number;
    static getCompilerHooks(compiler: any): Record<ForkTsCheckerHooks, any>;
    readonly options: Partial<ForkTsCheckerWebpackPlugin.Options>;
    private tsconfig;
    private compilerOptions;
    private tslint;
    private eslint;
    private eslintOptions;
    private tslintAutoFix;
    private watch;
    private ignoreDiagnostics;
    private ignoreLints;
    private ignoreLintWarnings;
    private reportFiles;
    private logger;
    private silent;
    private async;
    private checkSyntacticErrors;
    private workersNumber;
    private memoryLimit;
    private useColors;
    private colors;
    private formatter;
    private useTypescriptIncrementalApi;
    private resolveModuleNameModule;
    private resolveTypeReferenceDirectiveModule;
    private tsconfigPath;
    private tslintPath;
    private watchPaths;
    private compiler;
    private started;
    private elapsed;
    private cancellationToken;
    private isWatching;
    private checkDone;
    private compilationDone;
    private diagnostics;
    private lints;
    private emitCallback;
    private doneCallback;
    private typescriptPath;
    private typescript;
    private typescriptVersion;
    private tslintVersion;
    private eslintVersion;
    private service?;
    protected serviceRpc?: RpcProvider;
    private vue;
    private measureTime;
    private performance;
    private startAt;
    protected nodeArgs: string[];
    constructor(options?: Partial<ForkTsCheckerWebpackPlugin.Options>);
    private validateTypeScript;
    private validateTslint;
    private validateEslint;
    private static prepareVueOptions;
    private static createFormatter;
    apply(compiler: any): void;
    private computeContextPath;
    private pluginStart;
    private pluginStop;
    private pluginCompile;
    private pluginEmit;
    private pluginDone;
    private spawnService;
    private killService;
    private handleServiceMessage;
    private handleServiceExit;
    private createEmitCallback;
    private createNoopEmitCallback;
    private printLoggerMessage;
    private createDoneCallback;
}
export = ForkTsCheckerWebpackPlugin;
