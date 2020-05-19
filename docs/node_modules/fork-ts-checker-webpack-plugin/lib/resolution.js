"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
function makeResolutionFunctions(resolveModuleName, resolveTypeReferenceDirective) {
    resolveModuleName =
        resolveModuleName ||
            ((
            // tslint:disable-next-line:no-shadowed-variable
            typescript, moduleName, containingFile, 
            // tslint:disable-next-line:no-shadowed-variable
            compilerOptions, moduleResolutionHost) => {
                return typescript.resolveModuleName(moduleName, containingFile, compilerOptions, moduleResolutionHost);
            });
    resolveTypeReferenceDirective =
        resolveTypeReferenceDirective ||
            ((
            // tslint:disable-next-line:no-shadowed-variable
            typescript, typeDirectiveName, containingFile, 
            // tslint:disable-next-line:no-shadowed-variable
            compilerOptions, moduleResolutionHost) => {
                return typescript.resolveTypeReferenceDirective(typeDirectiveName, containingFile, compilerOptions, moduleResolutionHost);
            });
    return { resolveModuleName, resolveTypeReferenceDirective };
}
exports.makeResolutionFunctions = makeResolutionFunctions;
//# sourceMappingURL=resolution.js.map