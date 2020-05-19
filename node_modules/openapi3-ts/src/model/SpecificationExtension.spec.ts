import "mocha";
import { expect } from "chai";
import { SpecificationExtension, ISpecificationExtension } from "./";

describe("SpecificationExtension", () => {
    it("addExtension() ok", () => {
        let sut = new SpecificationExtension();
        let extensionValue = { payload: 5 };
        sut.addExtension("x-name", extensionValue);

        expect(sut["x-name"]).eql(extensionValue);
    });
    it("addExtension() invalid", (done) => {
        let sut = new SpecificationExtension();
        let extensionValue = { payload: 5 };
        try {
            sut.addExtension("y-name", extensionValue);
            done("Must fail. Invalid extension");
        }
        catch (err) {
            done();
        }
    });
    it("getExtension() ok", () => {
        let sut = new SpecificationExtension();
        let extensionValue1 = { payload: 5 };
        let extensionValue2 = { payload: 6 };
        sut.addExtension("x-name", extensionValue1);
        sut.addExtension("x-load", extensionValue2);

        expect(sut.getExtension("x-name")).eql(extensionValue1);
        expect(sut.getExtension("x-load")).eql(extensionValue2);
    });
    it("getExtension() invalid", (done) => {
        let sut = new SpecificationExtension();
        try {
            sut.getExtension("y-name");
            done("Error. invalid extension");
        }
        catch (err) {
            done();
        }
    });
    it("getExtension() not found", () => {
        let sut = new SpecificationExtension();
        expect(sut.getExtension("x-resource")).eql(null);
    });
    it("listExtensions()", () => {
        let sut = new SpecificationExtension();
        let extensionValue1 = { payload: 5 };
        let extensionValue2 = { payload: 6 };
        sut.addExtension("x-name", extensionValue1);
        sut.addExtension("x-load", extensionValue2);

        expect(sut.listExtensions()).eql(["x-name", "x-load"]);
    });
});
