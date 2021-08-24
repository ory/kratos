# Manual changelog for UBnity version

Quick list of things which changed, so it can be compared with similar changes which may
be available in the upstream repository.

**Last Kratos upstream check**: 2021-08-24

---

Format:
```
* {checkbox - checked if similar change also in upstream} {`*` if not in a stable version} {title / explanation}
```
Examples:
> * [ ] Latest awesome change
> * [x] * Awesome change already in upstream, but not in a stable version
> * [x] Awesome change already in upstream, stable version

## Log

See the Git history for timeline and details of code changes:
* [ ] Allow building and pushing docker images to UBnity AWS registry for dev branches and tagged versions
* [x] * Disable password check at api.pwnedpasswords.com (https://haveibeenpwned.com) - see this [official Kratos PR](https://github.com/ory/kratos/pull/1445)
* [x] * Fix issue where Kratos would use the public (internet) URL to download the schema for some validations, instead of using the local file system schema from the config - see this [official Kratos PR](https://github.com/ory/kratos/pull/1449)
* [ ] Returns more specific error codes for min length and incorrect format when submitted data fails schema validation, and a generic code otherwise (the code used to return the generic code for everything - note: while other schema validation modes
**might** be enforceable, Kratos does not provide error codes for those use cases, and they were not considered here either - at this
time UBnity is not using any other validation modes anyway)
