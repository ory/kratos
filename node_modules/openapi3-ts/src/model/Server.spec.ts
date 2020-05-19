import "mocha";
import { expect } from "chai";
import { Server, ServerVariable } from "./Server";
import * as oa from "../model";

describe("Server", () => {
    it("create server", () => {
        let v1 = new ServerVariable("dev", ["dev", "qa", "prod"], "environment");
        let sut = new Server("http://api.qa.machine.org", "qa maquine");
        sut.addVariable("environment", v1);

        expect(sut.url).eql("http://api.qa.machine.org");
        expect(sut.description).eql("qa maquine");
        expect(sut.variables.environment.default).eql("dev");
    });
});

describe("ServerVariable", () => {
    it("server var", () => {
        let sut = new ServerVariable("dev", ["dev", "qa", "prod"], "environment");

        expect(sut.default).eql("dev");
        expect(sut.description).eql("environment");
        expect(sut.enum).eql(["dev", "qa", "prod"]);
    });
});
