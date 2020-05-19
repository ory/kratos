import "mocha";
import { expect } from "chai";

import { OpenApiBuilder, Server, ServerVariable } from  ".";

describe("Top barrel", () => {
    it("OpenApiBuilder is exported", () => {
        const sut = OpenApiBuilder.create();
        expect(sut).not.null;
    });
    it("Server is exported", () => {
        const sut = new Server("a", "b");
        expect(sut).not.null;
    });
    it("ServerVariable is exported", () => {
        const sut = new ServerVariable("a", "b", "c");
        expect(sut).not.null;
    });
});