alter table selfservice_login_flows add column organization_id char(36) null;
alter table selfservice_registration_flows add column organization_id char(36) null;
