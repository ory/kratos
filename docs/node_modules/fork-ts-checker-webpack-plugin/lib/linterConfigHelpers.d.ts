import { Configuration } from 'tslint';
import * as minimatch from 'minimatch';
export interface ConfigurationFile extends Configuration.IConfigurationFile {
    linterOptions?: {
        typeCheck?: boolean;
        exclude?: string[];
    };
}
export declare function loadLinterConfig(configFile: string): ConfigurationFile;
export declare function makeGetLinterConfig(linterConfigs: Record<string, ConfigurationFile | undefined>, linterExclusions: minimatch.IMinimatch[], context: string): (file: string) => ConfigurationFile | undefined;
