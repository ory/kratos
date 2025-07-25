local claims = std.extVar('claims');
local session = std.extVar('session');

{
  claims: {
    foo: "bar",
    sub: "can not be overwritten",
    schema_id: session.identity.schema_id,
    aal: session.authenticator_assurance_level,
    second_claim: claims.exp,
    [if std.objectHas(session.identity, 'external_id') then 'external_id']: session.identity.external_id,
  }
}
