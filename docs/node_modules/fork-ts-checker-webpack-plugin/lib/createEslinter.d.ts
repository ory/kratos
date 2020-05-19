import * as eslinttypes from 'eslint';
import { NormalizedMessage } from './NormalizedMessage';
export declare function createEslinter(eslintOptions: object): {
    getLints: (filepath: string) => eslinttypes.CLIEngine.LintReport | undefined;
    getFormattedLints: (lintReports: IterableIterator<eslinttypes.CLIEngine.LintReport> | eslinttypes.CLIEngine.LintReport[]) => NormalizedMessage[];
};
