local subject = {sub: ''} + std.extVar('claims').sub;

{
  identity: {
    traits: {
      subject: std.extVar('claims').sub,
    },
  },
}
