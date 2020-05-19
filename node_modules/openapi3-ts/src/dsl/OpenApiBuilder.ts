import * as oa from "../model";

// Internal DSL for building an OpenAPI 3.0.x contract
// using a fluent interface

export class OpenApiBuilder {
    rootDoc: oa.OpenAPIObject;

    static create(doc?: oa.OpenAPIObject): OpenApiBuilder {
        return new OpenApiBuilder(doc);
    }

    constructor(doc?: oa.OpenAPIObject) {
        this.rootDoc = doc || {
            openapi: "3.0.0",
            info: {
                title: "app",
                version: "version"
            },
            paths:  {},
            components:  {
                schemas: {},
                responses: {},
                parameters: {},
                examples: {},
                requestBodies: {},
                headers: {},
                securitySchemes: {},
                links: {},
                callbacks: {}
            },
            tags: [],
            servers: []
        };
    }

    getSpec(): oa.OpenAPIObject {
        return this.rootDoc;
    }
    getSpecAsJson(replacer?: (key: string, value: any ) => any,
                  space?: string | number): string {
        return JSON.stringify(this.rootDoc, replacer, space);
    }
    getSpecAsYaml(): string {
        // Todo
        throw Error("Not yet implemented.");
    }

    private static isValidOpenApiVersion(v: string = ""): boolean {
        let match = /(\d+)\.(\d+).(\d+)/.exec(v);
        if (match) {
            let major = parseInt(match[1], 10);
            if (major >= 3) {
                return true;
            }
        }
        return false;
    }

    addOpenApiVersion(openApiVersion: string): OpenApiBuilder {
        if (!OpenApiBuilder.isValidOpenApiVersion(openApiVersion)) {
            throw new Error("Invalid OpnApi version: " + openApiVersion + ". Follow convention: 3.x.y");
        }
        this.rootDoc.openapi = openApiVersion;
        return this;
    }
    addInfo(info: oa.InfoObject): OpenApiBuilder {
        this.rootDoc.info = info;
        return this;
    }
    addContact(contact: oa.ContactObject): OpenApiBuilder {
        this.rootDoc.info.contact = contact;
        return this;
    }
    addLicense(license: oa.LicenseObject): OpenApiBuilder {
        this.rootDoc.info.license = license;
        return this;
    }
    addTitle(title: string): OpenApiBuilder {
        this.rootDoc.info.title = title;
        return this;
    }
    addDescription(description: string): OpenApiBuilder {
        this.rootDoc.info.description = description;
        return this;
    }
    addTermsOfService(termsOfService: string): OpenApiBuilder {
        this.rootDoc.info.termsOfService = termsOfService;
        return this;
    }
    addVersion(version: string): OpenApiBuilder {
        this.rootDoc.info.version = version;
        return this;
    }
    addPath(path: string, pathItem: oa.PathItemObject): OpenApiBuilder {
        this.rootDoc.paths[path] = pathItem;
        return this;
    }
    addSchema(name: string, schema: oa.SchemaObject | oa.ReferenceObject): OpenApiBuilder {
        this.rootDoc.components.schemas[name] = schema;
        return this;
    }
    addResponse(name: string, response: oa.ResponseObject | oa.ReferenceObject): OpenApiBuilder {
        this.rootDoc.components.responses[name] = response;
        return this;
    }
    addParameter(name: string, parameter: oa.ParameterObject | oa.ReferenceObject): OpenApiBuilder {
        this.rootDoc.components.parameters[name] = parameter;
        return this;
    }
    addExample(name: string, example: oa.ExampleObject | oa.ReferenceObject): OpenApiBuilder {
        this.rootDoc.components.examples[name] = example;
        return this;
    }
    addRequestBody(name: string, reqBody: oa.RequestBodyObject | oa.ReferenceObject): OpenApiBuilder {
        this.rootDoc.components.requestBodies[name] = reqBody;
        return this;
    }
    addHeader(name: string, header: oa.HeaderObject | oa.ReferenceObject): OpenApiBuilder {
        this.rootDoc.components.headers[name] = header;
        return this;
    }
    addSecurityScheme(name: string, secScheme: oa.SecuritySchemeObject | oa.ReferenceObject): OpenApiBuilder {
        this.rootDoc.components.securitySchemes[name] = secScheme;
        return this;
    }
    addLink(name: string, link: oa.LinkObject | oa.ReferenceObject): OpenApiBuilder {
        this.rootDoc.components.links[name] = link;
        return this;
    }
    addCallback(name: string, callback: oa.CallbackObject | oa.ReferenceObject): OpenApiBuilder {
        this.rootDoc.components.callbacks[name] = callback;
        return this;
    }
    addServer(server: oa.ServerObject): OpenApiBuilder {
        this.rootDoc.servers.push(server);
        return this;
    }
    addTag(tag: oa.TagObject): OpenApiBuilder {
        this.rootDoc.tags.push(tag);
        return this;
    }
    addExternalDocs(extDoc: oa.ExternalDocumentationObject): OpenApiBuilder {
        this.rootDoc.externalDocs = extDoc;
        return this;
    }
}
