---
id: html-forms
title: HTML Form Parser
---

If you're using HTML Forms to sign users up or update profiles, ORY Kratos needs
to assert the type of each field, as HTML Form Field Values are untyped.

ORY Kratos uses the Identity JSON Schema to assert form field types. There are a
few tricks you should know when using this feature.

## Nesting

Assuming this Identity JSON Schema:

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "properties": {
    "traits": {
      "type": "object",
      "properties": {
        "name": {
          "type": "object",
          "properties": {
            "first": {
              "type": "string"
            },
            "last": {
              "type": "string"
            }
          }
        }
      }
    }
  }
}
```

You could address `traits.name.first` this with an HTML Input Form:

```
<input type="text" name="traits.name.last">
```

## Checkboxes

Checkboxes for `true` / `false` can be set up as follows. If the value is
supposed to be false: not checked:

```
<input type="hidden" value="false" name="traits.path.to.my.boolean" />
<input type="checkbox" value="true" name="traits.path.to.my.boolean" />
```
