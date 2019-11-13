package form

//
// func TestNewFormFieldsFromJSON(t *testing.T) {
// 	var js = json.RawMessage(`{"numby":1.5,"stringy":"foobar","objy":{"objy":{},"numby":1.5,"stringy":"foobar"}}`)
//
// 	assert.EqualValues(t, Fields{
// 		"traits.numby":        Field{Name: "traits.numby", Type: "number"},
// 		"traits.objy.numby":   Field{Name: "traits.objy.numby", Type: "number"},
// 		"traits.objy.objy":    Field{Name: "traits.objy.objy", Type: "text"},
// 		"traits.objy.stringy": Field{Name: "traits.objy.stringy", Type: "text"},
// 		"traits.stringy":      Field{Name: "traits.stringy", Type: "text"},
// 		CSRFTokenName:         Field{Name: CSRFTokenName, Type: "text", Value: "foo_token"},
// 	}, NewFormFieldsFromJSON(js, "traits", "foo_token"))
//
// 	assert.EqualValues(t, Fields{
// 		"numby":        Field{Name: "numby", Type: "number", Value:},
// 		"objy.numby":   Field{Name: "objy.numby", Type: "number"},
// 		"objy.objy":    Field{Name: "objy.objy", Type: "text"},
// 		"objy.stringy": Field{Name: "objy.stringy", Type: "text"},
// 		"stringy":      Field{Name: "stringy", Type: "text"},
// 		CSRFTokenName:  Field{Name: CSRFTokenName, Type: "text", Value: "foo_token"},
// 	}, NewFormFieldsFromJSON(js, "", "foo_token"))
// }
//
// func TestNewFormFieldsFromJSONSchema(t *testing.T) {
// 	for k, tc := range []struct {
// 		f        string
// 		prefix   string
// 		expected Fields
// 	}{
// 		{
// 			f:      "../stub/new-form.json",
// 			prefix: "",
// 			expected: Fields{
// 				"numby":        Field{Name: "numby", Type: "number"},
// 				"objy.numby":   Field{Name: "objy.numby", Type: "number"},
// 				"objy.objy":    Field{Name: "objy.objy", Type: "text"},
// 				"objy.stringy": Field{Name: "objy.stringy", Type: "text"},
// 				"stringy":      Field{Name: "stringy", Type: "text"},
// 				CSRFTokenName:  Field{Name: CSRFTokenName, Type: "hidden", Value: "foo_token"},
// 			},
// 		},
// 		{
// 			f:      "../stub/new-form.json",
// 			prefix: "traits",
// 			expected: Fields{
// 				"traits.numby":        Field{Name: "traits.numby", Type: "number"},
// 				"traits.objy.numby":   Field{Name: "traits.objy.numby", Type: "number"},
// 				"traits.objy.objy":    Field{Name: "traits.objy.objy", Type: "text"},
// 				"traits.objy.stringy": Field{Name: "traits.objy.stringy", Type: "text"},
// 				"traits.stringy":      Field{Name: "traits.stringy", Type: "text"},
// 				CSRFTokenName:         Field{Name: CSRFTokenName, Type: "hidden", Value: "foo_token"},
// 			},
// 		},
// 	} {
// 		t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
// 			fields, err := NewFormFieldsFromJSONSchema(tc.f, tc.prefix, "foo_token")
// 			require.NoError(t, err)
// 			assert.EqualValues(t, tc.expected, fields)
// 		})
// 	}
// }
