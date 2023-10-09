alter table selfservice_login_flows drop column organization_id;
alter table selfservice_registration_flows drop column organization_id;
alter table identities drop column organization_id;
