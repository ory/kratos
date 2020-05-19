import "mocha";
import { expect } from "chai";
import { OpenApiBuilder } from ".";
import * as oa from "../model";

describe("OpenApiBuilder", () => {
    it("Build empty Spec", () => {
        expect(OpenApiBuilder.create().getSpec()).eql({
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
        });
    });
    it("Build with custom object", () => {
        const obj: oa.OpenAPIObject = {
            openapi: "3.0.0",
            info: {
                title: "app1",
                version: "version2"
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
        expect(OpenApiBuilder.create(obj).getSpec()).eql(obj);
    });
    it("addTitle", () => {
        let sut = OpenApiBuilder.create().addTitle("app7").rootDoc;
        expect(sut.info.title).eql("app7");
    });
    it("addDescription", () => {
        let sut = OpenApiBuilder.create().addDescription("desc 6").rootDoc;
        expect(sut.info.description).eql("desc 6");
    });
    it("addOpenApiVersion valid", () => {
        let sut = OpenApiBuilder.create().addOpenApiVersion("3.2.4").rootDoc;
        expect(sut.openapi).eql("3.2.4");
    });
    it("addOpenApiVersion invalid", (done) => {
        try {
            let sut = OpenApiBuilder.create().addOpenApiVersion("a.b.4").rootDoc;
            done("failed");
        }
        catch (err) {
            done();
        }
    });
    it("addOpenApiVersion empty", (done) => {
        try {
            let sut = OpenApiBuilder.create().addOpenApiVersion("").rootDoc;
            done("failed");
        }
        catch (err) {
            done();
        }
    });
    it("addOpenApiVersion null", (done) => {
        try {
            let sut = OpenApiBuilder.create().addOpenApiVersion(null).rootDoc;
            done("failed");
        }
        catch (err) {
            done();
        }
    });
    it("addOpenApiVersion lower than 3", (done) => {
        try {
            let sut = OpenApiBuilder.create().addOpenApiVersion("2.5.6").rootDoc;
            done("failed");
        }
        catch (err) {
            done();
        }
    });
     it("addInfo", () => {
        let info: oa.InfoObject = {
            title: "app9",
            version: "11.34.678"
        };
        let sut = OpenApiBuilder.create().addInfo(info).rootDoc;
        expect(sut.info).eql(info);
    });
    it("addTitle", () => {
        let sut = OpenApiBuilder.create().addTitle("t1").rootDoc;
        expect(sut.info.title).eql("t1");
    });
    it("addDescription", () => {
        let sut = OpenApiBuilder.create().addDescription("desc 2").rootDoc;
        expect(sut.info.description).eql("desc 2");
    });
    it("addTermsOfService", () => {
        let sut = OpenApiBuilder.create().addTermsOfService("tos 7").rootDoc;
        expect(sut.info.termsOfService).eql("tos 7");
    });
    it("addLicense", () => {
        let sut = OpenApiBuilder.create().addLicense({
            name: "MIT",
            url: "http://mit.edu/license"
        }).rootDoc;
        expect(sut.info.license).eql({
            name: "MIT",
            url: "http://mit.edu/license"
        });
    });
    it("addContact", () => {
        let sut = OpenApiBuilder.create().addContact({
            name: "Alicia",
            email: "alicia@acme.com",
            url: "http://acme.com/~alicia"
        }).rootDoc;
        expect(sut.info.contact).eql({
            name: "Alicia",
            email: "alicia@acme.com",
            url: "http://acme.com/~alicia"
        });
    });
    it("addVersion", () => {
        let sut = OpenApiBuilder.create().addVersion("7.52.46").rootDoc;
        expect(sut.info.version).eql("7.52.46");
    });
    it("addPath", () => {
        let path1 = {
            get: {
                responses: {
                    default: {
                        description: "object created"
                    }
                }
            }
        };
        let sut = OpenApiBuilder.create().addPath("/path1", path1).rootDoc;
        expect(sut.paths["/path1"]).eql(path1);
    });
    it("addSchema", () => {
        let schema1 = {
            type: "string",
            format: "email"
        };
        let sut = OpenApiBuilder.create().addSchema("schema01", schema1).rootDoc;
        expect(sut.components.schemas.schema01).eql(schema1);
    });
    it("addSchema reference", () => {
        let schema1 = {
            $ref: "#/components/schemas/id"
        };
        let sut = OpenApiBuilder.create().addSchema("schema01", schema1).rootDoc;
        expect(sut.components.schemas.schema01).eql(schema1);
    });
    it("addResponse", () => {
        let resp00 = {
            description: "object created"
        };
        let sut = OpenApiBuilder.create().addResponse("resp00", resp00).rootDoc;
        expect(sut.components.responses.resp00).eql(resp00);
    });
    it("addResponse reference", () => {
        let resp00 = {
            $ref: "#/components/responses/reference"
        };
        let sut = OpenApiBuilder.create().addResponse("resp00", resp00).rootDoc;
        expect(sut.components.responses.resp00).eql(resp00);
    });
    it("addParameter", () => {
        let par5 = {
            name: "id",
            in:   "header" as oa.ParameterLocation,
            schema: {
                $ref: "#/components/schemas/id"
            }
        };
        let sut = OpenApiBuilder.create().addParameter("par5", par5).rootDoc;
        expect(sut.components.parameters.par5).eql(par5);
    });
    it("addParameter reference", () => {
        let par5 = {
            $ref: "#/components/parameters/id"
        };
        let sut = OpenApiBuilder.create().addParameter("par5", par5).rootDoc;
        expect(sut.components.parameters.par5).eql(par5);
    });
    it("addExample", () => {
        let example4 = {
            a: "a desc",
            b: "a desc"
        };
        let sut = OpenApiBuilder.create().addExample("example4", example4).rootDoc;
        expect(sut.components.examples.example4).eql(example4);
    });
    it("addExample reference", () => {
        let example4 = {
            $ref: "#/components/examples/id"
        };
        let sut = OpenApiBuilder.create().addExample("example4", example4).rootDoc;
        expect(sut.components.examples.example4).eql(example4);
    });
    it("addRequestBody", () => {
        let reqBody9: oa.RequestBodyObject = {
            description: "Request body details",
            content: {
                "application/json": {
                    "schema": {
                        "$ref": "#/components/schemas/User"
                    },
                    "examples": {
                        user: {
                            "$ref": "http://foo.bar/examples/user-example.json"
                        }
                    }
                }
            },
            required: false
        };
        let sut = OpenApiBuilder.create().addRequestBody("reqBody9", reqBody9).rootDoc;
        expect(sut.components.requestBodies.reqBody9).eql(reqBody9);
    });
    it("addRequestBody reference", () => {
        let reqBody9 = {
            $ref: "#/components/requestBodies/id"
        };
        let sut = OpenApiBuilder.create().addRequestBody("reqBody9", reqBody9).rootDoc;
        expect(sut.components.requestBodies.reqBody9).eql(reqBody9);
    });
    it("addHeaders", () => {
        let h5: oa.HeaderObject = {
            description: "header 5"
        };
        let sut = OpenApiBuilder.create().addHeader("h5", h5).rootDoc;
        expect(sut.components.headers.h5).eql(h5);
    });
    it("addHeaders Reference", () => {
        let h5: oa.HeaderObject = {
            $ref: "#/components/headers/id"
        };
        let sut = OpenApiBuilder.create().addHeader("h5", h5).rootDoc;
        expect(sut.components.headers.h5).eql(h5);
    });
    it("addSecuritySchemes", () => {
        let sec7: oa.SecuritySchemeObject = {
           type: "http",
           scheme: "basic"
        };
        let sut = OpenApiBuilder.create().addSecurityScheme("sec7", sec7).rootDoc;
        expect(sut.components.securitySchemes.sec7).eql(sec7);
    });
    it("addSecuritySchemes reference", () => {
        let sec7 = {
            $ref: "#/components/securitySchemes/id"
        };
        let sut = OpenApiBuilder.create().addSecurityScheme("sec7", sec7).rootDoc;
        expect(sut.components.securitySchemes.sec7).eql(sec7);
    });
    it("addLink", () => {
        let link0: oa.LinkObject = {
            href: "/users/10101110/department"
        };
        let sut = OpenApiBuilder.create().addLink("link0", link0).rootDoc;
        expect(sut.components.links.link0).eql(link0);
    });
    it("addLink reference", () => {
        let link0 = {
            $ref: "#/components/links/id"
        };
        let sut = OpenApiBuilder.create().addLink("link0", link0).rootDoc;
        expect(sut.components.links.link0).eql(link0);
    });
    it("addCallback", () => {
        let cb1: oa.CallbackObject = {
            '$request.body#/url': {
                post: {
                    requestBody: {
                        description: "Callback payload",
                        content: {
                            'application/json': {
                                schema: {
                                    $ref: '#/components/schemas/SomePayload'
                                }
                            }
                        }
                    },
                    responses: {
                        "200": {
                            description: "webhook successfully processed an no retries will be performed"
                        }
                    }
                }
            }
        };
        let sut = OpenApiBuilder.create().addCallback("cb1", cb1).rootDoc;
        expect(sut.components.callbacks.cb1).eql(cb1);
    });
    it("addCallback reference", () => {
        let cb1 = {
            $ref: "#/components/callbacks/id"
        };
        let sut = OpenApiBuilder.create().addCallback("cb1", cb1).rootDoc;
        expect(sut.components.callbacks.cb1).eql(cb1);
    });
    it("addTag", () => {
        let t1: oa.TagObject = {
           name: "resource",
           "x-admin": true,
           description: "my own tag",
        };
        let sut = OpenApiBuilder.create().addTag(t1).rootDoc;
        expect(sut.tags[0]).eql(t1);
    });
    it("addExternalDocs", () => {
        let eDocs: oa.ExternalDocumentationObject = {
            url: "https://acme.com/docs",
            description: "Main doc"
        };
        let sut = OpenApiBuilder.create().addExternalDocs(eDocs).rootDoc;
        expect(sut.externalDocs).eql(eDocs);
    });
    it("addServer", () => {
        let s1: oa.ServerObject = {
           url: "http://api.quixote.org",
           variables: {}
        };
        let sut = OpenApiBuilder.create().addServer(s1).rootDoc;
        expect(sut.servers[0]).eql(s1);
    });

    it("getPath", () => {
        let path1 = {
            get: {
                responses: {
                    default: {
                        description: "object created"
                    }
                }
            }
        };
        const sut = OpenApiBuilder.create()
                                .addPath('/service7', path1)
                                .rootDoc;
        oa.addExtension(sut.paths, 'x-my-extension', 42);

        expect(oa.getPath(sut.paths, '/service7')).eql(path1);
        expect(oa.getPath(sut.paths, '/service56')).eql(undefined);
    });
    it("get invalid Path", () => {
        let path1 = {
            get: {
                responses: {
                    default: {
                        description: "object created"
                    }
                }
            }
        };
        const sut = OpenApiBuilder.create()
                                .addPath('/service7', path1)
                                .rootDoc;
        oa.addExtension(sut.paths, 'x-my-extension', 42);

        expect(oa.getPath(sut.paths, 'x-path')).eql(undefined);
    });
    it("getExtension", () => {
        let path1 = {
            get: {
                responses: {
                    default: {
                        description: "object created"
                    }
                }
            }
        };
        const sut = OpenApiBuilder.create()
                                .addPath('/service7', path1)
                                .rootDoc;
        oa.addExtension(sut.paths, 'x-my-extension', 42);

        expect(oa.getExtension(sut.paths, 'x-my-extension')).eql(42);
        expect(oa.getExtension(sut.paths, 'x-other')).eql(undefined);
    });
    it("retrieve invalid extension", () => {
        let path1 = {
            get: {
                responses: {
                    default: {
                        description: "object created"
                    }
                }
            }
        };
        const sut = OpenApiBuilder.create()
                                .addPath('/service7', path1)
                                .rootDoc;
        oa.addExtension(sut.paths, 'x-my-extension', 42);

        expect(oa.getExtension(sut.paths, 'y-other')).eql(undefined);
    });

    describe("Serialize", () => {
        it("getSpecAsJson", () => {
            let sut = OpenApiBuilder.create()
                                    .addTitle("app9")
                                    .addVersion("5.6.7")
                                    .getSpecAsJson();
            expect(sut).eql(
`{"openapi":"3.0.0","info":{"title":"app9","version":"5.6.7"},"paths":{},"components":{"schemas":{},"responses":{},"parameters":{},"examples":{},"requestBodies":{},"headers":{},"securitySchemes":{},"links":{},"callbacks":{}},"tags":[],"servers":[]}`
);
        });
        it("getSpecAsYaml", (done) => {
            try {
                let sut = OpenApiBuilder.create()
                            .addTitle("app9")
                            .addVersion("5.6.7")
                            .getSpecAsYaml();
                done('Not implemented');
            }
            catch (e) {
                expect(e.message).eql('Not yet implemented.');
                done();
            }
        });
    });
});
