local claims = std.extVar('claims');

local raw = claims.raw_claims;

if std.length(claims.sub) == 0 then
  error 'claim sub not set'
else
  {
    identity: {
      traits: {
        subject: claims.sub,
        [if "email" in claims && claims.email != '' then "email" else null]: claims.email,
        [if "fullnameEN" in raw && raw.fullnameEN != '' then "name" else null]: {
          full: raw.fullnameEN,
          [if "firstnameEN" in raw && raw.firstnameEN != '' then "first" else null]: raw.firstnameEN,
          [if "lastnameEN" in raw && raw.lastnameEN != '' then "last" else null]: raw.lastnameEN,
        },
        [if "fullnameAR" in raw && raw.fullnameAR != '' then "name_ar" else null]: {
          full: raw.fullnameAR,
          [if "firstnameAR" in raw && raw.firstnameAR != '' then "first" else null]: raw.firstnameAR,
          [if "lastnameAR" in raw && raw.lastnameAR != '' then "last" else null]: raw.lastnameAR,
        },
        [if "uuid" in raw && raw.uuid != '' then "uuid" else null]: raw.uuid,
        [if "unifiedId" in raw && raw.unifiedId != '' then "unified_id" else null]: raw.unifiedId,
        [if "idn" in raw && raw.idn != '' then "idn" else null]: raw.idn,
        [if "userType" in raw && raw.userType != '' then "user_type" else null]: raw.userType,
        [if "nationalityEN" in raw && raw.nationalityEN != '' then "nationality" else null]: raw.nationalityEN,
        [if "gender" in raw && raw.gender != '' then "gender" else null]: raw.gender,
        [if "mobile" in raw && raw.mobile != '' then "phone" else null]: raw.mobile,
      },
    },
  }
