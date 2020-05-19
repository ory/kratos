import * as ts from 'typescript';
import { RuleFailure } from 'tslint';
import { CancellationToken } from './CancellationToken';
import { NormalizedMessage } from './NormalizedMessage';
import { ResolveTypeReferenceDirective, ResolveModuleName } from './resolution';
import { createEslinter } from './createEslinter';
import { VueOptions } from './types/vue-options';
export interface IncrementalCheckerInterface {
    nextIteration(): void;
    getDiagnostics(cancellationToken: CancellationToken): Promise<NormalizedMessage[]>;
    hasLinter(): boolean;
    getLints(cancellationToken: CancellationToken): NormalizedMessage[];
    hasEsLinter(): boolean;
    getEsLints(cancellationToken: CancellationToken): NormalizedMessage[];
}
export interface ApiIncrementalCheckerParams {
    typescript: typeof ts;
    context: string;
    programConfigFile: string;
    compilerOptions: ts.CompilerOptions;
    createNormalizedMessageFromDiagnostic: (diagnostic: ts.Diagnostic) => NormalizedMessage;
    linterConfigFile: string | boolean;
    linterAutoFix: boolean;
    createNormalizedMessageFromRuleFailure: (ruleFailure: RuleFailure) => NormalizedMessage;
    eslinter: ReturnType<typeof createEslinter> | undefined;
    checkSyntacticErrors: boolean;
    resolveModuleName: ResolveModuleName | undefined;
    resolveTypeReferenceDirective: ResolveTypeReferenceDirective | undefined;
    vue: VueOptions;
}
export interface IncrementalCheckerParams extends ApiIncrementalCheckerParams {
    watchPaths: string[];
    workNumber: number;
    workDivision: number;
}
