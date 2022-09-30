local claims = std.extVar('claims');

{
  identity: {
    traits: {
      name: claims.name,
      name_en: claims.name_en,
      email: claims.user_name,
      // email: claims.users[0].clients[0].business_email,
      lang: claims.lang,
      language: claims.language,
      user_name: claims.user_name,
      account_id: claims.account_id,
    },
  },
}