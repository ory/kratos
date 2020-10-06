# Schema Tests

This test works by validating cases of the payload against a schema.
To add a case, place a file in the `${schemaName}.schema.test.${success|failure}` folder.
A success case should validate, a failure case should throw a validation error.

## Handling of "$ref"

To allow testing definitions on their own, and reduce the size of individual cases,
every `"$ref"` in the schema will be replaced with `"const"` before validation.
This means that a case only has to define the pointer as a string. Example:

```yaml
default_browser_return_url: "#/definitions/defaultReturnTo"
hooks:
  - "#/definitions/selfServiceSessionRevokerHook"
```
