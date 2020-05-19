import * as ts from 'typescript';
import * as tslint from 'tslint';
import { NormalizedMessage } from './NormalizedMessage';
import * as eslint from 'eslint';
export declare const makeCreateNormalizedMessageFromDiagnostic: (typescript: typeof ts) => (diagnostic: ts.Diagnostic) => NormalizedMessage;
export declare const makeCreateNormalizedMessageFromRuleFailure: () => (lint: tslint.RuleFailure) => NormalizedMessage;
export declare const makeCreateNormalizedMessageFromEsLintFailure: () => (lint: eslint.Linter.LintMessage, filePath: string) => NormalizedMessage;
export declare const makeCreateNormalizedMessageFromInternalError: () => (error: any) => NormalizedMessage;
