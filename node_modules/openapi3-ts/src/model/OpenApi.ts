// Typed interfaces for OpenAPI 3.0.0-RC
// see https://github.com/OAI/OpenAPI-Specification/blob/3.0.0-rc0/versions/3.0.md

import { ISpecificationExtension, SpecificationExtension } from "./SpecificationExtension";

export function getExtension(obj: ISpecificationExtension, extensionName: string): any {
    if (SpecificationExtension.isValidExtension(extensionName)) {
        return obj[extensionName];
    }
    return undefined;
}
export function addExtension(obj: ISpecificationExtension, extensionName: string, extension: any): void {
    if (SpecificationExtension.isValidExtension(extensionName)) {
        obj[extensionName] = extension;
    }
}

export interface OpenAPIObject extends ISpecificationExtension {
    openapi: string;
    info: InfoObject;
    servers?: ServerObject[];
    paths: PathsObject;
    components?: ComponentsObject;
    security?: SecurityRequirementObject[];
    tags?: TagObject[];
    externalDocs?: ExternalDocumentationObject;
}
export interface InfoObject extends ISpecificationExtension {
    title: string;
    description?: string;
    termsOfService?: string;
    contact?: ContactObject;
    license?: LicenseObject;
    version: string;
}
export interface ContactObject extends ISpecificationExtension {
    name?: string;
    url?: string;
    email?: string;
}
export interface LicenseObject extends ISpecificationExtension {
    name: string;
    url?: string;
}
export interface ServerObject extends ISpecificationExtension {
    url: string;
    description?: string;
    variables?: { [v: string]: ServerVariableObject };
}
export interface ServerVariableObject extends ISpecificationExtension {
    enum?: string[] | boolean[] | number[];
    default: string | boolean | number;
    description?: string;
}
export interface ComponentsObject extends ISpecificationExtension {
    schemas?: { [schema: string]: SchemaObject | ReferenceObject };
    responses?: { [response: string]: ResponseObject | ReferenceObject };
    parameters?: { [parameter: string]: ParameterObject | ReferenceObject };
    examples?: { [example: string]: ExampleObject | ReferenceObject };
    requestBodies?: { [request: string]: RequestBodyObject | ReferenceObject };
    headers?: { [heaer: string]: HeaderObject | ReferenceObject };
    securitySchemes?: { [securityScheme: string]: SecuritySchemeObject | ReferenceObject };
    links?: { [link: string]: LinkObject | ReferenceObject };
    callbacks?: { [callback: string]: CallbackObject | ReferenceObject };
}

/**
 * Rename it to Paths Object to be consistent with the spec
 * See https://github.com/OAI/OpenAPI-Specification/blob/master/versions/3.0.0.md#pathsObject
 */
export interface PathsObject extends ISpecificationExtension {
    // [path: string]: PathItemObject;
    [path: string]: PathItemObject | any;   // Hack for allowing ISpecificationExtension
}

/**
 * @deprecated
 * Create a type alias for backward compatibility
 */
export type PathObject = PathsObject;

export function getPath(pathsObject: PathsObject, path: string): PathItemObject | undefined {
    if (SpecificationExtension.isValidExtension(path)) {
        return undefined;
    }
    return pathsObject[path] as PathItemObject;
}

export interface PathItemObject extends ISpecificationExtension {
    $ref?: string;
    summary?: string;
    description?: string;
    get?: OperationObject;
    put?: OperationObject;
    post?: OperationObject;
    delete?: OperationObject;
    options?: OperationObject;
    head?: OperationObject;
    patch?: OperationObject;
    trace?: OperationObject;
    servers?: ServerObject[];
    parameters?: (ParameterObject | ReferenceObject)[];
}
export interface OperationObject extends ISpecificationExtension {
    tags?: string[];
    summary?: string;
    description?: string;
    externalDocs?: ExternalDocumentationObject;
    operationId?: string;
    parameters?: (ParameterObject | ReferenceObject)[];
    requestBody?: RequestBodyObject | ReferenceObject;
    responses: ResponsesObject;
    callbacks?: CallbacksObject;
    deprecated?: boolean;
    security?: SecurityRequirementObject[];
    servers?: ServerObject[];
}
export interface ExternalDocumentationObject extends ISpecificationExtension {
    description?: string;
    url: string;
}

/**
 * The location of a parameter.
 * Possible values are "query", "header", "path" or "cookie".
 * Specification:
 * https://github.com/OAI/OpenAPI-Specification/blob/master/versions/3.0.0.md#parameter-locations
 */
export type ParameterLocation = 'query' | 'header' | 'path' | 'cookie';

/**
 * The style of a parameter.
 * Describes how the parameter value will be serialized.
 * (serialization is not implemented yet)
 * Specification:
 * https://github.com/OAI/OpenAPI-Specification/blob/master/versions/3.0.0.md#style-values
 */
export type ParameterStyle =
    'matrix'
  | 'label'
  | 'form'
  | 'simple'
  | 'spaceDelimited'
  | 'pipeDelimited'
  | 'deepObject';

export interface BaseParameterObject extends ISpecificationExtension {
    description?: string;
    required?: boolean;
    deprecated?: boolean;
    allowEmptyValue?: boolean;

    style?: ParameterStyle; // "matrix" | "label" | "form" | "simple" | "spaceDelimited" | "pipeDelimited" | "deepObject";
    explode?: boolean;
    allowReserved?: boolean;
    schema?: SchemaObject | ReferenceObject;
    examples?: { [param: string]: ExampleObject | ReferenceObject };
    example?: any;
    content?: ContentObject;
}

export interface ParameterObject extends BaseParameterObject {
    name: string;
    in: ParameterLocation; // "query" | "header" | "path" | "cookie";
}
export interface RequestBodyObject extends ISpecificationExtension {
    description?: string;
    content: ContentObject;
    required?: boolean;
}
export interface ContentObject {
    [mediatype: string]: MediaTypeObject;
}
export interface MediaTypeObject extends ISpecificationExtension {
    schema?: SchemaObject | ReferenceObject;
    examples?: ExamplesObject;
    example?: any;
    encoding?: EncodingObject;
}
export interface EncodingObject extends ISpecificationExtension {
    // [property: string]: EncodingPropertyObject;
    [property: string]: EncodingPropertyObject | any;   // Hack for allowing ISpecificationExtension
}
export interface EncodingPropertyObject {
    contentType?: string;
    headers?: {[key: string]: HeaderObject | ReferenceObject };
    style?: string;
    explode?: boolean;
    allowReserved?: boolean;
    [key: string]: any;   // (any) = Hack for allowing ISpecificationExtension
}
export interface ResponsesObject extends ISpecificationExtension {
    default?: ResponseObject | ReferenceObject;

    // [statuscode: string]: ResponseObject | ReferenceObject;
    [statuscode: string]: ResponseObject | ReferenceObject | any;   // (any) = Hack for allowing ISpecificationExtension
}
export interface ResponseObject extends ISpecificationExtension {
    description: string;
    headers?: HeadersObject;
    content?: ContentObject;
    links?: LinksObject;
}
export interface CallbacksObject extends ISpecificationExtension {
    // [name: string]: CallbackObject | ReferenceObject;
    [name: string]: CallbackObject | ReferenceObject | any;   // Hack for allowing ISpecificationExtension
}
export interface CallbackObject extends ISpecificationExtension {
    // [name: string]: PathItemObject;
    [name: string]: PathItemObject | any;   // Hack for allowing ISpecificationExtension
}
export interface HeadersObject {
    [name: string]: HeaderObject | ReferenceObject;
}
export interface ExampleObject {
    summary?: string;
    description?: string;
    value?: any;
    externalValue?: string;
    [property: string]: any; // Hack for allowing ISpecificationExtension
}
export interface LinksObject {
    [name: string]: LinkObject | ReferenceObject;
}
export interface LinkObject extends ISpecificationExtension {
    operationRef?: string;
    operationId?: string;
    parameters?: LinkParametersObject;
    requestBody?: any | string;
    description?: string;
    server?: ServerObject;
    [property: string]: any; // Hack for allowing ISpecificationExtension
}
export interface LinkParametersObject {
    [name: string]: any | string;
}
export interface HeaderObject extends BaseParameterObject {
}
export interface TagObject extends ISpecificationExtension {
    name: string;
    description?: string;
    externalDocs?: ExternalDocumentationObject;
    [extension: string]: any; // Hack for allowing ISpecificationExtension
}
export interface ExamplesObject {
    [name: string]: ExampleObject | ReferenceObject;
}

export interface ReferenceObject {
    $ref: string;
}

/**
 * A type guard to check if the given value is a `ReferenceObject`.
 * See https://www.typescriptlang.org/docs/handbook/advanced-types.html#type-guards-and-differentiating-types
 *
 * @param obj The value to check.
 */
export function isReferenceObject(obj: object): obj is ReferenceObject {
    return obj.hasOwnProperty("$ref");
}

export interface SchemaObject extends ISpecificationExtension {
    nullable?: boolean;
    discriminator?: DiscriminatorObject;
    readOnly?: boolean;
    writeOnly?: boolean;
    xml?: XmlObject;
    externalDocs?: ExternalDocumentationObject;
    example?: any;
    examples?: any[];
    deprecated?: boolean;

    type?: string;
    allOf?: (SchemaObject | ReferenceObject)[];
    oneOf?: (SchemaObject | ReferenceObject)[];
    anyOf?: (SchemaObject | ReferenceObject)[];
    not?: SchemaObject | ReferenceObject;
    items?: SchemaObject | ReferenceObject;
    properties?: {[propertyName: string]: (SchemaObject | ReferenceObject)};
    additionalProperties?: (SchemaObject | ReferenceObject | boolean);
    description?: string;
    format?: string;
    default?: any;

    title?: string;
    multipleOf?: number;
    maximum?: number;
    exclusiveMaximum?: boolean;
    minimum?: number;
    exclusiveMinimum?: boolean;
    maxLength?: number;
    minLength?: number;
    pattern?: string;
    maxItems?: number;
    minItems?: number;
    uniqueItems?: boolean;
    maxProperties?: number;
    minProperties?: number;
    required?: string[];
    enum?: any[];
}

/**
 * A type guard to check if the given object is a `SchemaObject`.
 * Useful to distinguish from `ReferenceObject` values that can be used
 * in most places where `SchemaObject` is allowed.
 *
 * See https://www.typescriptlang.org/docs/handbook/advanced-types.html#type-guards-and-differentiating-types
 *
 * @param schema The value to check.
 */
export function isSchemaObject(schema: SchemaObject | ReferenceObject): schema is SchemaObject {
    return !schema.hasOwnProperty('$ref');
}

export interface SchemasObject {
    [schema: string]: SchemaObject;
}

export interface DiscriminatorObject {
    propertyName: string;
    mapping?: {[key: string]: string };
}

export interface XmlObject extends ISpecificationExtension {
    name?: string;
    namespace?: string;
    prefix?: string;
    attribute?: boolean;
    wrapped?: boolean;
}
export type SecuritySchemeType =
      "apiKey"
    | "http"
    | "oauth2"
    | "openIdConnect";

export interface SecuritySchemeObject extends ISpecificationExtension {
    type: SecuritySchemeType;
    description?: string;
    name?: string;              // required only for apiKey
    in?: string;                // required only for apiKey
    scheme?: string;            // required only for http
    bearerFormat?: string;
    flows?: OAuthFlowsObject;     // required only for oauth2
    openIdConnectUrl?: string;  // required only for openIdConnect
}
export interface OAuthFlowsObject extends ISpecificationExtension {
    implicit?: OAuthFlowObject;
    password?: OAuthFlowObject;
    clientCredentials?: OAuthFlowObject;
    authorizationCode?: OAuthFlowObject;
}
export interface OAuthFlowObject extends ISpecificationExtension {
    authorizationUrl?: string;
    tokenUrl?: string;
    refreshUrl?: string;
    scopes: ScopesObject;
}
export interface ScopesObject extends ISpecificationExtension {
    [scope: string]: any; // Hack for allowing ISpecificationExtension
}
export interface SecurityRequirementObject {
    [name: string]: string[];
}
