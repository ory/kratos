// Suport for Specification Extensions
// as described in
// https://github.com/OAI/OpenAPI-Specification/blob/3.0.0-rc0/versions/3.0.md#specificationExtensions

//  Specification Extensions
//   ^x-
export interface ISpecificationExtension {
    // Cannot constraint to "^x-" but can filter them later to access to them
    [extensionName: string]: any;
}

export class SpecificationExtension implements ISpecificationExtension {
    // Cannot constraint to "^x-" but can filter them later to access to them
    [extensionName: string]: any;

    static isValidExtension(extensionName: string) {
        return /^x\-/.test(extensionName);
    }

    getExtension(extensionName: string): any {
        if (!SpecificationExtension.isValidExtension(extensionName)) {
            throw new Error("Invalid specification extension: '" +
                        extensionName + "'. Extensions must start with prefix 'x-");
        }
        if (this[extensionName]) {
            return this[extensionName];
        }
        return null;
    }
    addExtension(extensionName: string, payload: any): void {
        if (!SpecificationExtension.isValidExtension(extensionName)) {
            throw new Error("Invalid specification extension: '" +
                        extensionName + "'. Extensions must start with prefix 'x-");
        }
        this[extensionName] = payload;
    }
    listExtensions(): string[] {
        let res: string[] = [];
        for (let propName in this) {
            if (this.hasOwnProperty(propName)) {
                if (SpecificationExtension.isValidExtension(propName)) {
                    res.push(propName);
                }
            }
        }
        return res;
    }
}