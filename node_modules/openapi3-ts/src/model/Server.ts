import * as oa from "./OpenApi";

// Server & Server Variable

export class Server implements oa.ServerObject {
    url: string;
    description?: string;
    variables: { [v: string]: ServerVariable };

    constructor(url: string, desc?: string) {
        this.url = url;
        this.description = desc;
        this.variables = {};
    }
    addVariable(name: string, variable: ServerVariable) {
        this.variables[name] = variable;
    }
}

export class ServerVariable implements oa.ServerVariableObject {
    enum?: string[] | boolean[] | number[];
    default: string | boolean | number;
    description?: string;

    constructor(defaultValue: any,
                enums?: any,
                description?: string) {
        this.default = defaultValue;
        this.enum = enums;
        this.description = description;
    }
}