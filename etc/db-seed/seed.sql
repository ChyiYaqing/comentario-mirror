--
-- PostgreSQL database dump
--

-- Dumped from database version 11.16 (Debian 11.16-1.pgdg90+1)
-- Dumped by pg_dump version 11.16 (Debian 11.16-1.pgdg90+1)

set statement_timeout = 0;
set lock_timeout = 0;
set idle_in_transaction_session_timeout = 0;
set client_encoding = 'UTF8';
set standard_conforming_strings = on;
select pg_catalog.set_config('search_path', '', false);
set check_function_bodies = false;
set xmloption = content;
set client_min_messages = warning;
set row_security = off;

--
-- Name: sortpolicy; Type: TYPE; Schema: public; Owner: postgres
--

create type public.sortpolicy as enum (
    'score-desc',
    'creationdate-desc',
    'creationdate-asc'
    );


alter type public.sortpolicy owner to postgres;

--
-- Name: commentsinserttriggerfunction(); Type: FUNCTION; Schema: public; Owner: postgres
--

create function public.commentsinserttriggerfunction() returns trigger
    language plpgsql
as
$$
begin
    update public.pages
    set commentCount = commentCount + 1
        where domain = new.domain and path = new.path;

    return NEW;
end;
$$;


alter function public.commentsinserttriggerfunction() owner to postgres;

--
-- Name: viewsinserttriggerfunction(); Type: FUNCTION; Schema: public; Owner: postgres
--

create function public.viewsinserttriggerfunction() returns trigger
    language plpgsql
as
$$
begin
    update public.domains
    set viewsThisMonth = viewsThisMonth + 1
        where domain = new.domain;

    return null;
end;
$$;


alter function public.viewsinserttriggerfunction() owner to postgres;

--
-- Name: votesinserttriggerfunction(); Type: FUNCTION; Schema: public; Owner: postgres
--

create function public.votesinserttriggerfunction() returns trigger
    language plpgsql
as
$$
begin
    update public.comments
    set score = score + new.direction
        where commentHex = new.commentHex;

    return NEW;
end;
$$;


alter function public.votesinserttriggerfunction() owner to postgres;

--
-- Name: votesupdatetriggerfunction(); Type: FUNCTION; Schema: public; Owner: postgres
--

create function public.votesupdatetriggerfunction() returns trigger
    language plpgsql
as
$$
begin
    update public.comments
    set score = score - old.direction + new.direction
        where commentHex = old.commentHex;

    return NEW;
end;
$$;


alter function public.votesupdatetriggerfunction() owner to postgres;

set default_tablespace = '';

set default_with_oids = false;

--
-- Name: commenters; Type: TABLE; Schema: public; Owner: postgres
--

create table public.commenters (
    commenterhex text                        not null,
    email        text                        not null,
    name         text                        not null,
    link         text                        not null,
    photo        text                        not null,
    provider     text                        not null,
    joindate     timestamp without time zone not null,
    state        text default 'ok'::text     not null,
    passwordhash text default ''::text       not null
);


alter table public.commenters
owner to postgres;

--
-- Name: commentersessions; Type: TABLE; Schema: public; Owner: postgres
--

create table public.commentersessions (
    commentertoken text                        not null,
    commenterhex   text default 'none'::text   not null,
    creationdate   timestamp without time zone not null
);


alter table public.commentersessions
owner to postgres;

--
-- Name: comments; Type: TABLE; Schema: public; Owner: postgres
--

create table public.comments (
    commenthex   text                               not null,
    domain       text                               not null,
    path         text                               not null,
    commenterhex text                               not null,
    markdown     text                               not null,
    html         text                               not null,
    parenthex    text                               not null,
    score        integer default 0                  not null,
    state        text    default 'unapproved'::text not null,
    creationdate timestamp without time zone        not null,
    deleted      boolean default false              not null,
    deleterhex   text,
    deletiondate timestamp without time zone
);


alter table public.comments
owner to postgres;

--
-- Name: config; Type: TABLE; Schema: public; Owner: postgres
--

create table public.config (
    version text not null
);


alter table public.config
owner to postgres;

--
-- Name: domains; Type: TABLE; Schema: public; Owner: postgres
--

create table public.domains (
    domain                  text                                                      not null,
    ownerhex                text                                                      not null,
    name                    text                                                      not null,
    creationdate            timestamp without time zone                               not null,
    state                   text              default 'unfrozen'::text                not null,
    importedcomments        text              default false                           not null,
    autospamfilter          boolean           default true                            not null,
    requiremoderation       boolean           default false                           not null,
    requireidentification   boolean           default true                            not null,
    viewsthismonth          integer           default 0                               not null,
    moderateallanonymous    boolean           default true,
    emailnotificationpolicy text              default 'pending-moderation'::text,
    commentoprovider        boolean           default true                            not null,
    googleprovider          boolean           default true                            not null,
    twitterprovider         boolean           default true                            not null,
    githubprovider          boolean           default true                            not null,
    gitlabprovider          boolean           default true                            not null,
    ssoprovider             boolean           default false                           not null,
    ssosecret               text              default ''::text                        not null,
    ssourl                  text              default ''::text                        not null,
    defaultsortpolicy       public.sortpolicy default 'score-desc'::public.sortpolicy not null
);


alter table public.domains
owner to postgres;

--
-- Name: emails; Type: TABLE; Schema: public; Owner: postgres
--

create table public.emails (
    email                      text                        not null,
    unsubscribesecrethex       text                        not null,
    lastemailnotificationdate  timestamp without time zone not null,
    pendingemails              integer default 0           not null,
    sendreplynotifications     boolean default false       not null,
    sendmoderatornotifications boolean default true        not null
);


alter table public.emails
owner to postgres;

--
-- Name: exports; Type: TABLE; Schema: public; Owner: postgres
--

create table public.exports (
    exporthex    text                        not null,
    bindata      bytea                       not null,
    domain       text                        not null,
    creationdate timestamp without time zone not null
);


alter table public.exports
owner to postgres;

--
-- Name: migrations; Type: TABLE; Schema: public; Owner: postgres
--

create table public.migrations (
    filename text not null
);


alter table public.migrations
owner to postgres;

--
-- Name: moderators; Type: TABLE; Schema: public; Owner: postgres
--

create table public.moderators (
    domain  text                        not null,
    email   text                        not null,
    adddate timestamp without time zone not null
);


alter table public.moderators
owner to postgres;

--
-- Name: ownerconfirmhexes; Type: TABLE; Schema: public; Owner: postgres
--

create table public.ownerconfirmhexes (
    confirmhex text not null,
    ownerhex   text not null,
    senddate   text not null
);


alter table public.ownerconfirmhexes
owner to postgres;

--
-- Name: owners; Type: TABLE; Schema: public; Owner: postgres
--

create table public.owners (
    ownerhex       text                        not null,
    email          text                        not null,
    name           text                        not null,
    passwordhash   text                        not null,
    confirmedemail text default false          not null,
    joindate       timestamp without time zone not null
);


alter table public.owners
owner to postgres;

--
-- Name: ownersessions; Type: TABLE; Schema: public; Owner: postgres
--

create table public.ownersessions (
    ownertoken text                        not null,
    ownerhex   text                        not null,
    logindate  timestamp without time zone not null
);


alter table public.ownersessions
owner to postgres;

--
-- Name: pages; Type: TABLE; Schema: public; Owner: postgres
--

create table public.pages (
    domain           text                         not null,
    path             text                         not null,
    islocked         boolean default false        not null,
    commentcount     integer default 0            not null,
    stickycommenthex text    default 'none'::text not null,
    title            text    default ''::text
);


alter table public.pages
owner to postgres;

--
-- Name: resethexes; Type: TABLE; Schema: public; Owner: postgres
--

create table public.resethexes (
    resethex text                       not null,
    hex      text                       not null,
    senddate text                       not null,
    entity   text default 'owner'::text not null
);


alter table public.resethexes
owner to postgres;

--
-- Name: ssotokens; Type: TABLE; Schema: public; Owner: postgres
--

create table public.ssotokens (
    token          text                        not null,
    domain         text                        not null,
    commentertoken text                        not null,
    creationdate   timestamp without time zone not null
);


alter table public.ssotokens
owner to postgres;

--
-- Name: views; Type: TABLE; Schema: public; Owner: postgres
--

create table public.views (
    domain       text                        not null,
    commenterhex text                        not null,
    viewdate     timestamp without time zone not null
);


alter table public.views
owner to postgres;

--
-- Name: votes; Type: TABLE; Schema: public; Owner: postgres
--

create table public.votes (
    commenthex   text                        not null,
    commenterhex text                        not null,
    direction    integer                     not null,
    votedate     timestamp without time zone not null
);


alter table public.votes
owner to postgres;

--
-- Data for Name: commenters; Type: TABLE DATA; Schema: public; Owner: postgres
--

copy public.commenters(commenterhex, email, name, link, photo, provider, joindate, state, passwordhash) from stdin;
\.


--
-- Data for Name: commentersessions; Type: TABLE DATA; Schema: public; Owner: postgres
--

copy public.commentersessions(commentertoken, commenterhex, creationdate) from stdin;
\.


--
-- Data for Name: comments; Type: TABLE DATA; Schema: public; Owner: postgres
--

copy public.comments(commenthex, domain, path, commenterhex, markdown, html, parenthex, score, state, creationdate,
                     deleted, deleterhex, deletiondate) from stdin;
\.


--
-- Data for Name: config; Type: TABLE DATA; Schema: public; Owner: postgres
--

copy public.config(version) from stdin;
v1.7.0
\.


--
-- Data for Name: domains; Type: TABLE DATA; Schema: public; Owner: postgres
--

copy public.domains(domain, ownerhex, name, creationdate, state, importedcomments, autospamfilter, requiremoderation,
                    requireidentification, viewsthismonth, moderateallanonymous, emailnotificationpolicy,
                    commentoprovider, googleprovider, twitterprovider, githubprovider, gitlabprovider, ssoprovider,
                    ssosecret, ssourl, defaultsortpolicy) from stdin;
\.


--
-- Data for Name: emails; Type: TABLE DATA; Schema: public; Owner: postgres
--

copy public.emails(email, unsubscribesecrethex, lastemailnotificationdate, pendingemails, sendreplynotifications,
                   sendmoderatornotifications) from stdin;
\.


--
-- Data for Name: exports; Type: TABLE DATA; Schema: public; Owner: postgres
--

copy public.exports(exporthex, bindata, domain, creationdate) from stdin;
\.


--
-- Data for Name: migrations; Type: TABLE DATA; Schema: public; Owner: postgres
--

copy public.migrations(filename) from stdin;
20180416163802-init-schema.sql
20180610215858-commenter-password.sql
20180620083655-session-token-renamme.sql
20180724125115-remove-config.sql
20180922181651-page-attributes.sql
20180923002745-comment-count.sql
20180923004309-comment-count-build.sql
20181007230906-store-version.sql
20181007231407-v1.1.4.sql
20181218183803-sticky-comments.sql
20181228114101-v1.4.0.sql
20181228114101-v1.4.1.sql
20190122235525-anonymous-moderation-default.sql
20190123002724-v1.4.2.sql
20190131002240-export.sql
20190204180609-v1.5.0.sql
20190213033530-email-notifications.sql
20190218173502-v1.6.0.sql
20190218183556-v1.6.1.sql
20190219001130-v1.6.2.sql
20190418210855-configurable-auth.sql
20190420181913-sso.sql
20190420231030-sso-tokens.sql
20190501201032-v1.7.0.sql
20190505191006-comment-count-decrease.sql
20190508222848-reset-count.sql
20190606000842-reset-hex.sql
20190913175445-delete-comments.sql
20191204173000-sort-method.sql
20210228122203-comment-delete-log.sql
\.


--
-- Data for Name: moderators; Type: TABLE DATA; Schema: public; Owner: postgres
--

copy public.moderators(domain, email, adddate) from stdin;
\.


--
-- Data for Name: ownerconfirmhexes; Type: TABLE DATA; Schema: public; Owner: postgres
--

copy public.ownerconfirmhexes(confirmhex, ownerhex, senddate) from stdin;
\.


--
-- Data for Name: owners; Type: TABLE DATA; Schema: public; Owner: postgres
--

copy public.owners(ownerhex, email, name, passwordhash, confirmedemail, joindate) from stdin;
\.


--
-- Data for Name: ownersessions; Type: TABLE DATA; Schema: public; Owner: postgres
--

copy public.ownersessions(ownertoken, ownerhex, logindate) from stdin;
\.


--
-- Data for Name: pages; Type: TABLE DATA; Schema: public; Owner: postgres
--

copy public.pages(domain, path, islocked, commentcount, stickycommenthex, title) from stdin;
\.


--
-- Data for Name: resethexes; Type: TABLE DATA; Schema: public; Owner: postgres
--

copy public.resethexes(resethex, hex, senddate, entity) from stdin;
\.


--
-- Data for Name: ssotokens; Type: TABLE DATA; Schema: public; Owner: postgres
--

copy public.ssotokens(token, domain, commentertoken, creationdate) from stdin;
\.


--
-- Data for Name: views; Type: TABLE DATA; Schema: public; Owner: postgres
--

copy public.views(domain, commenterhex, viewdate) from stdin;
\.


--
-- Data for Name: votes; Type: TABLE DATA; Schema: public; Owner: postgres
--

copy public.votes(commenthex, commenterhex, direction, votedate) from stdin;
\.


--
-- Name: commenters commenters_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

alter table only public.commenters
add constraint commenters_pkey primary key (commenterhex);


--
-- Name: commentersessions commentersessions_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

alter table only public.commentersessions
add constraint commentersessions_pkey primary key (commentertoken);


--
-- Name: comments comments_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

alter table only public.comments
add constraint comments_pkey primary key (commenthex);


--
-- Name: domains domains_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

alter table only public.domains
add constraint domains_pkey primary key (domain);


--
-- Name: emails emails_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

alter table only public.emails
add constraint emails_pkey primary key (email);


--
-- Name: emails emails_unsubscribesecrethex_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

alter table only public.emails
add constraint emails_unsubscribesecrethex_key unique (unsubscribesecrethex);


--
-- Name: exports exports_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

alter table only public.exports
add constraint exports_pkey primary key (exporthex);


--
-- Name: migrations migrations_filename_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

alter table only public.migrations
add constraint migrations_filename_key unique (filename);


--
-- Name: moderators moderators_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

alter table only public.moderators
add constraint moderators_pkey primary key (domain, email);


--
-- Name: ownerconfirmhexes ownerconfirmhexes_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

alter table only public.ownerconfirmhexes
add constraint ownerconfirmhexes_pkey primary key (confirmhex);


--
-- Name: resethexes ownerresethexes_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

alter table only public.resethexes
add constraint ownerresethexes_pkey primary key (resethex);


--
-- Name: owners owners_email_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

alter table only public.owners
add constraint owners_email_key unique (email);


--
-- Name: owners owners_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

alter table only public.owners
add constraint owners_pkey primary key (ownerhex);


--
-- Name: ownersessions ownersessions_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

alter table only public.ownersessions
add constraint ownersessions_pkey primary key (ownertoken);


--
-- Name: ssotokens ssotokens_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

alter table only public.ssotokens
add constraint ssotokens_pkey primary key (token);


--
-- Name: domainindex; Type: INDEX; Schema: public; Owner: postgres
--

create index domainindex on public.views using btree(domain);


--
-- Name: pagesuniqueindex; Type: INDEX; Schema: public; Owner: postgres
--

create unique index pagesuniqueindex on public.pages using btree(domain, path);


--
-- Name: unsubscribesecrethexindex; Type: INDEX; Schema: public; Owner: postgres
--

create index unsubscribesecrethexindex on public.emails using btree(unsubscribesecrethex);


--
-- Name: votesuniqueindex; Type: INDEX; Schema: public; Owner: postgres
--

create unique index votesuniqueindex on public.votes using btree(commenthex, commenterhex);


--
-- Name: comments commentsinserttrigger; Type: TRIGGER; Schema: public; Owner: postgres
--

create trigger commentsinserttrigger
    after insert
    on public.comments
    for each row
execute procedure public.commentsinserttriggerfunction();


--
-- Name: views viewsinserttrigger; Type: TRIGGER; Schema: public; Owner: postgres
--

create trigger viewsinserttrigger
    after insert
    on public.views
    for each row
execute procedure public.viewsinserttriggerfunction();


--
-- Name: votes votesinserttrigger; Type: TRIGGER; Schema: public; Owner: postgres
--

create trigger votesinserttrigger
    after insert
    on public.votes
    for each row
execute procedure public.votesinserttriggerfunction();


--
-- Name: votes votesupdatetrigger; Type: TRIGGER; Schema: public; Owner: postgres
--

create trigger votesupdatetrigger
    after update
    on public.votes
    for each row
execute procedure public.votesupdatetriggerfunction();



--**********************************************************************************************************************
------------------------------------------------------------------------------------------------------------------------
-- SEED DATA
------------------------------------------------------------------------------------------------------------------------
--**********************************************************************************************************************

-- Make admin user root@comentario.app / admin
insert into public.owners(ownerhex, email, name, passwordhash, confirmedemail, joindate)
    values
        ('05878df7449326d8ad6d2fdc5c3d703fb04c72ea1a0efaa5e02ea2c3855a42e2', 'root@comentario.app', 'Admin User',
         '$2a$10$WLeCsMc7z7vSdococ9FLF.9FdcrIsJAQCeCSYFbiqFk8qRVQ/pqRK', 'true', '2023-01-17 17:55:47.008851');

insert into public.commenters (commenterhex, email, name, link, photo, provider, joindate, state, passwordhash)
    values
        ('d668b826923228bd75c64a8b99cc3d8dfa4179dd7e8121eaeced9eee8d4e20db', 'root@comentario.app', 'Admin User', 'undefined', 'undefined',
         'commento', '2023-01-17 18:23:43.604399', 'ok', '$2a$10$WLeCsMc7z7vSdococ9FLF.9FdcrIsJAQCeCSYFbiqFk8qRVQ/pqRK');

insert into public.domains(domain, ownerhex, name, creationdate, state, importedcomments, autospamfilter,
                           requiremoderation, requireidentification, viewsthismonth, moderateallanonymous,
                           emailnotificationpolicy, commentoprovider, googleprovider, twitterprovider, githubprovider,
                           gitlabprovider, ssoprovider, ssosecret, ssourl, defaultsortpolicy)
    values
        ('localhost:8100', '05878df7449326d8ad6d2fdc5c3d703fb04c72ea1a0efaa5e02ea2c3855a42e2', 'Test Domain',
         '2023-01-17 17:56:10.966890', 'unfrozen', 'false', true, false, true, 0, true, 'pending-moderation', true, true,
         true, true, true, false, '', '', 'score-desc');

insert into public.emails (email, unsubscribesecrethex, lastemailnotificationdate, pendingemails, sendreplynotifications, sendmoderatornotifications)
    values
        ('root@comentario.app', '1dae2342c9255a4ecc78f2f54380d90508aa49761f3471e94239f178a210bcb8', '2023-01-17 17:55:46.953534', 0, false, true);

insert into public.moderators (domain, email, adddate)
    values
        ('localhost:8100', 'root@comentario.app', '2023-01-17 17:56:10.968427');

insert into public.pages (domain, path, islocked, commentcount, stickycommenthex, title)
    values
        ('localhost:8100', '/', false, 1, 'none', '');

insert into public.comments (commenthex, domain, path, commenterhex, markdown, html, parenthex, score, state, creationdate, deleted, deleterhex, deletiondate)
    values
        ('805dca5d3ff5b7131c28c7054325b8d7aac7062145422438902911d9d50bd03b', 'localhost:8100', '/',
         'd668b826923228bd75c64a8b99cc3d8dfa4179dd7e8121eaeced9eee8d4e20db', 'Hey there!', '<p>Hey there!</p>', 'root',
         0, 'approved', '2023-01-17 18:28:10.767326', false, null, null);
