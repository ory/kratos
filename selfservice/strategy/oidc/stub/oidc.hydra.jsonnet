local subject = std.toString(std.extVar('claims').sub);

if std.length(subject) == 0 then error 'claim sub not set'

{
  identity: {
    traits: {
      subject: subject,
    },
  },
}
