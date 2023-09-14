alter table selfservice_login_flows add column organization_id uuid null;
alter table selfservice_registration_flows add column organization_id uuid null;
alter table identities add column organization_id uuid null;
