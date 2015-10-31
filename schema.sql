--
-- PostgreSQL database dump
--

SET statement_timeout = 0;
SET lock_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SET check_function_bodies = false;
SET client_min_messages = warning;

--
-- Name: plpgsql; Type: EXTENSION; Schema: -;
--

CREATE EXTENSION IF NOT EXISTS plpgsql WITH SCHEMA pg_catalog;


--
-- Name: EXTENSION plpgsql; Type: COMMENT; Schema: -;
--

COMMENT ON EXTENSION plpgsql IS 'PL/pgSQL procedural language';


SET search_path = public, pg_catalog;

--
-- Name: authorized_scopes_id_seq; Type: SEQUENCE; Schema: public;
--

CREATE SEQUENCE authorized_scopes_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: authorized_scopes; Type: TABLE; Schema: public;
--

CREATE TABLE authorized_scopes (
    id integer DEFAULT nextval('authorized_scopes_id_seq'::regclass) NOT NULL,
    oauth_token_id integer NOT NULL,
    scope_id integer NOT NULL
);


--
-- Name: applications_id_seq; Type: SEQUENCE; Schema: public;
--

CREATE SEQUENCE applications_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


SET default_tablespace = '';

SET default_with_oids = false;

--
-- Name: applications; Type: TABLE; Schema: public;
--

CREATE TABLE applications (
    id integer DEFAULT nextval('applications_id_seq'::regclass) NOT NULL,
    name character varying(510) NOT NULL,
    description text,
    website character varying(510) DEFAULT NULL::character varying,
    logo character varying(510) DEFAULT NULL::character varying,
    client_type character varying(510) NOT NULL,
    client_id character varying(510) NOT NULL,
    client_secret character varying(510) NOT NULL
);


--
-- Name: oauth_tokens_id_seq; Type: SEQUENCE; Schema: public;
--

CREATE SEQUENCE oauth_tokens_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: oauth_tokens; Type: TABLE; Schema: public;
--

CREATE TABLE oauth_tokens (
    id integer DEFAULT nextval('oauth_tokens_id_seq'::regclass) NOT NULL,
    client_id integer NOT NULL,
    user_id integer NOT NULL,
    type character varying(510) NOT NULL,
    token character varying(510) NOT NULL,
    expires_at timestamp with time zone
);


--
-- Name: permitted_scopes_id_seq; Type: SEQUENCE; Schema: public;
--

CREATE SEQUENCE permitted_scopes_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: permitted_scopes; Type: TABLE; Schema: public;
--

CREATE TABLE permitted_scopes (
    id integer DEFAULT nextval('permitted_scopes_id_seq'::regclass) NOT NULL,
    client_id integer NOT NULL,
    scope_id integer NOT NULL
);


--
-- Name: registered_redirects_id_seq; Type: SEQUENCE; Schema: public;
--

CREATE SEQUENCE registered_redirects_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: registered_redirects; Type: TABLE; Schema: public;
--

CREATE TABLE registered_redirects (
    id integer DEFAULT nextval('registered_redirects_id_seq'::regclass) NOT NULL,
    client_id integer NOT NULL,
    url character varying(510) NOT NULL,
    response_type character varying(510) NOT NULL
);


--
-- Name: scopes_id_seq; Type: SEQUENCE; Schema: public;
--

CREATE SEQUENCE scopes_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: scopes; Type: TABLE; Schema: public;
--

CREATE TABLE scopes (
    id integer DEFAULT nextval('scopes_id_seq'::regclass) NOT NULL,
    name character varying(510) NOT NULL,
    friendly_name character varying(510) NOT NULL,
    description character varying(510) NOT NULL
);


--
-- Name: users_id_seq; Type: SEQUENCE; Schema: public;
--

CREATE SEQUENCE users_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: users; Type: TABLE; Schema: public;
--

CREATE TABLE users (
    id integer DEFAULT nextval('users_id_seq'::regclass) NOT NULL,
    email character varying(510) NOT NULL,
    password character varying(510) NOT NULL
);


--
-- Data for Name: applications; Type: TABLE DATA; Schema: public;
--

COPY applications (id, name, description, website, logo, client_type, client_id, client_secret) FROM stdin;
1	Ibex				secret	2ddc784a1d3b452155647cac1becced6d41fcb9924ae92b2771481e90d1767c1	$2a$12$s5qwHNvwnmD2nafr9aTGge5eJgIsPyv7zIABZUxsBRbItLiOr3kS2
\.


--
-- Name: applications_id_seq; Type: SEQUENCE SET; Schema: public
--

SELECT pg_catalog.setval('applications_id_seq', 1, true);


--
-- Data for Name: oauth_tokens; Type: TABLE DATA; Schema: public
--

COPY oauth_tokens (id, client_id, user_id, type, token, scope, expires_at) FROM stdin;
\.


--
-- Name: oauth_tokens_id_seq; Type: SEQUENCE SET; Schema: public
--

SELECT pg_catalog.setval('oauth_tokens_id_seq', 1, true);


--
-- Data for Name: registered_redirects; Type: TABLE DATA; Schema: public
--

COPY registered_redirects (id, client_id, url, response_type) FROM stdin;
\.


--
-- Name: registered_redirects_id_seq; Type: SEQUENCE SET; Schema: public
--

SELECT pg_catalog.setval('registered_redirects_id_seq', 1, false);


--
-- Data for Name: scopes; Type: TABLE DATA; Schema: public;
--

INSERT INTO scopes (name, friendly_name, description) VALUES
  ('openid', 'OpenID', 'Your OpenID identity information'),
  ('email', 'email', 'Your email address'),
  ('profile', 'profile', 'Your profile information'),
  ('admin', 'administration', 'Access to administration endpoints');

-- primary client can request all scopes
INSERT INTO permitted_scopes (scope_id, client_id) SELECT id, 1 FROM scopes;

--
-- Data for Name: users; Type: TABLE DATA; Schema: public
--

COPY users (id, email, password) FROM stdin;
\.


--
-- Name: users_id_seq; Type: SEQUENCE SET; Schema: public
--

SELECT pg_catalog.setval('users_id_seq', 1, true);


--
-- Name: applications_pkey; Type: CONSTRAINT; Schema: public
--

ALTER TABLE ONLY applications
    ADD CONSTRAINT applications_pkey PRIMARY KEY (id);


--
-- Name: oauth_tokens_pkey; Type: CONSTRAINT; Schema: public
--

ALTER TABLE ONLY oauth_tokens
    ADD CONSTRAINT oauth_tokens_pkey PRIMARY KEY (id);


--
-- Name: registered_redirects_pkey; Type: CONSTRAINT; Schema: public
--

ALTER TABLE ONLY registered_redirects
    ADD CONSTRAINT registered_redirects_pkey PRIMARY KEY (id);


--
-- Name: users_email_key; Type: CONSTRAINT; Schema: public
--

ALTER TABLE ONLY users
    ADD CONSTRAINT users_email_key UNIQUE (email);


--
-- Name: users_pkey; Type: CONSTRAINT; Schema: public
--

ALTER TABLE ONLY users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- PostgreSQL database dump complete
--
