// UAE PASS identity mapper
// Maps UAE PASS claims to Kratos identity traits
//
// UAE PASS provides the following claims (depending on scopes):
// - sub: Subject identifier (unique user ID)
// - email: User's email address
// - fullnameEN / fullnameAR: Full name in English/Arabic
// - firstnameEN / firstnameAR: First name in English/Arabic
// - lastnameEN / lastnameAR: Last name in English/Arabic
// - uuid: UUID identifier
// - unifiedID: UAE unified identifier
// - idn: Identity number
// - userType: Profile type (SOP1=National, SOP2=Resident, SOP3=Visitor)
// - nationalityEN: Nationality
// - gender: Gender
// - dob: Date of birth
// - mobile: Mobile number

local claims = std.extVar('claims');

// Ensure we have a subject identifier
if std.length(claims.sub) == 0 then
  error 'claim sub not set'
else
  {
    identity: {
      traits: {
        // Required: subject identifier
        subject: claims.sub,

        // Email (if available)
        [if 'email' in claims && claims.email != '' then 'email' else null]: claims.email,

        // Name mapping - prefer English names
        [if 'fullnameEN' in claims && claims.fullnameEN != '' then 'name' else null]: {
          full: claims.fullnameEN,
          [if 'firstnameEN' in claims && claims.firstnameEN != '' then 'first' else null]: claims.firstnameEN,
          [if 'lastnameEN' in claims && claims.lastnameEN != '' then 'last' else null]: claims.lastnameEN,
        },

        // Arabic name (optional)
        [if 'fullnameAR' in claims && claims.fullnameAR != '' then 'name_ar' else null]: {
          full: claims.fullnameAR,
          [if 'firstnameAR' in claims && claims.firstnameAR != '' then 'first' else null]: claims.firstnameAR,
          [if 'lastnameAR' in claims && claims.lastnameAR != '' then 'last' else null]: claims.lastnameAR,
        },

        // UAE PASS specific identifiers
        [if 'uuid' in claims && claims.uuid != '' then 'uuid' else null]: claims.uuid,
        [if 'unifiedID' in claims && claims.unifiedID != '' then 'unified_id' else null]: claims.unifiedID,
        [if 'idn' in claims && claims.idn != '' then 'idn' else null]: claims.idn,

        // Profile metadata
        [if 'userType' in claims && claims.userType != '' then 'user_type' else null]: claims.userType,
        [if 'nationalityEN' in claims && claims.nationalityEN != '' then 'nationality' else null]: claims.nationalityEN,
        [if 'gender' in claims && claims.gender != '' then 'gender' else null]: claims.gender,
        [if 'dob' in claims && claims.dob != '' then 'date_of_birth' else null]: claims.dob,
        [if 'mobile' in claims && claims.mobile != '' then 'phone' else null]: claims.mobile,
      },
    },
  }
