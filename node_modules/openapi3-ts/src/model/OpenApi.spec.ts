import "mocha";
import { expect } from "chai";
import { isSchemaObject, isReferenceObject, SchemaObject, ReferenceObject } from "./OpenApi";

describe('type-guards unit tests', () => {
    describe('isSchemaObject()', () => {
        it('returns true for a schema object', () => {
            const schemaObject = new TestSchemaObject();
            expect(isSchemaObject(schemaObject)).to.equal(true);
        });

        it('returns false for a reference object', () => {
            const referenceObject = new TestReferenceObject();
            expect(isSchemaObject(referenceObject)).to.equal(false);
        });
    });

    describe('isReferenceObject()', () => {
        it('returns true for a reference object', () => {
            const referenceObject = new TestReferenceObject();
            expect(isReferenceObject(referenceObject)).to.equal(true);
        });

        it('returns false for a schema object', () => {
            const schemaObject = new TestSchemaObject();
            expect(isReferenceObject(schemaObject)).to.equal(false);
        });
    });
});

class TestSchemaObject implements SchemaObject {
    // empty schema
}

class TestReferenceObject implements ReferenceObject {
    $ref = 'test';
}
