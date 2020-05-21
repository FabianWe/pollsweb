-- Copyright 2020 Fabian Wenzelmann <fabianwen@posteo.eu>
--
-- Licensed under the Apache License, Version 2.0 (the "License");
-- you may not use this file except in compliance with the License.
-- You may obtain a copy of the License at
--
-- http://www.apache.org/licenses/LICENSE-2.0
--
-- Unless required by applicable law or agreed to in writing, software
-- distributed under the License is distributed on an "AS IS" BASIS,
-- WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
-- See the License for the specific language governing permissions and
-- limitations under the License.

BEGIN;

DO $$ BEGIN
IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'gopolls_fraction') THEN
        CREATE TYPE gopolls_fraction AS (
            numerator integer,
            denominator integer
);
END IF;
END $$;

CREATE TABLE IF NOT EXISTS period (
    id uuid PRIMARY KEY,
    name varchar(150) UNIQUE NOT NULL,
    slug varchar(150) UNIQUE NOT NULL,
    meeting_time timestamptz,
    period_start timestamptz NOT NULL,
    period_end timestamptz NOT NULL,
    created timestamptz NOT NULL DEFAULT (timezone('utc', now()))
);

CREATE INDEX IF NOT EXISTS period_period_start_idx ON period(period_start DESC);

CREATE TABLE IF NOT EXISTS voters_revision (
    id uuid PRIMARY KEY,
    period_id uuid NOT NULL REFERENCES period(id) ON DELETE CASCADE,
    name varchar(150) NOT NULL,
    slug varchar(150) NOT NULL,
    note text NOT NULL,
    is_active boolean NOT NULL DEFAULT true,
    created timestamptz NOT NULL DEFAULT (timezone('utc', now())),

    CONSTRAINT voters_revision_period_id_name_key UNIQUE(period_id, name),
    CONSTRAINT voters_revision_period_id_slug_key UNIQUE(period_id, slug)
);

CREATE TABLE IF NOT EXISTS voter (
    id uuid PRIMARY KEY,
    revision_id uuid NOT NULL REFERENCES voters_revision(id) ON DELETE CASCADE,
    name varchar(150) NOT NULL,
    slug varchar(150) NOT NULL,
    weight smallint NOT NULL CONSTRAINT voter_weight_positive_check CHECK (weight >= 0),

    CONSTRAINT voter_revision_id_name_key UNIQUE(revision_id, name),
    CONSTRAINT voter_revision_id_revision_slug_key UNIQUE(revision_id, slug)
);

CREATE INDEX IF NOT EXISTS voter_name_idx ON voter(name);

CREATE TABLE IF NOT EXISTS poll_collection (
    id uuid PRIMARY KEY,
    revision_id uuid NOT NULL REFERENCES voters_revision(id) ON DELETE CASCADE,
    name varchar(150) NOT NULL UNIQUE,
    slug varchar(150) NOT NULL UNIQUE,
    meeting_time timestamptz NOT NULL,
    online_start timestamptz,
    online_end timestamptz
);

CREATE INDEX IF NOT EXISTS poll_collection_revision_id_idx ON poll_collection(revision_id);
CREATE INDEX IF NOT EXISTS poll_collection_meeting_time_idx ON poll_collection(meeting_time);
CREATE INDEX IF NOT EXISTS poll_collection_online_start_online_end_idx ON poll_collection(online_start, online_end);

CREATE TABLE IF NOT EXISTS poll_group (
    id uuid PRIMARY KEY,
    collection_id uuid NOT NULL REFERENCES poll_collection(id) ON DELETE CASCADE,
    name varchar(150) NOT NULL,
    slug varchar(150) NOT NULL,
    group_num integer NOT NULL CONSTRAINT poll_group_group_num_positive_check CHECK (group_num >= 0),
    CONSTRAINT poll_group_collection_id_name_key UNIQUE(collection_id, name),
    CONSTRAINT poll_group_collection_id_slug_key UNIQUE(collection_id, slug),
    CONSTRAINT poll_group_collection_id_group_num_key UNIQUE(collection_id, group_num)
);

CREATE TABLE IF NOT EXISTS poll_base (
    id uuid PRIMARY KEY,
    group_id uuid NOT NULL REFERENCES poll_group(id) ON DELETE CASCADE,
    poll_num integer NOT NULL CONSTRAINT poll_base_poll_num_positive_check CHECK (poll_num >= 0),
    name varchar(150) NOT NULL,
    slug varchar(150) NOT NULL,
    majority gopolls_fraction NOT NULL,
    absolute_majority boolean NOT NULL,
    CONSTRAINT poll_base_group_id_poll_num_key UNIQUE(group_id, poll_num),
    CONSTRAINT poll_base_group_id_name_key UNIQUE(group_id, name),
    CONSTRAINT poll_base_group_id_slug_key UNIQUE(group_id, slug)
);

CREATE INDEX IF NOT EXISTS poll_base_name_idx ON poll_base(name);


CREATE TABLE IF NOT EXISTS median_poll (
    id uuid PRIMARY KEY REFERENCES poll_base(id) ON DELETE CASCADE,
    value bigint NOT NULL CONSTRAINT median_poll_value_positive_check CHECK (value >= 0),
    currency varchar(10) NOT NULL
);

CREATE TABLE IF NOT EXISTS basic_poll (
    id uuid PRIMARY KEY REFERENCES poll_base(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS schulze_poll (
    id uuid PRIMARY KEY REFERENCES poll_base(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS schulze_option (
    id uuid PRIMARY KEY,
    schulze_poll_id uuid NOT NULL REFERENCES schulze_poll(id) ON DELETE CASCADE,
    option varchar(150) NOT NULL,
    option_num integer NOT NULL CONSTRAINT schulze_option_option_num_positive_check CHECK (option_num >= 0),
    CONSTRAINT schulze_option_schulze_poll_id_option_key UNIQUE(schulze_poll_id, option),
    CONSTRAINT schulze_option_schulze_poll_id_option_num_key UNIQUE(schulze_poll_id, option_num)
);

CREATE TABLE IF NOT EXISTS median_vote (
    id uuid PRIMARY KEY,
    poll_id uuid NOT NULL REFERENCES median_poll(id) ON DELETE CASCADE,
    voter_id uuid NOT NULL REFERENCES voter(id) ON DELETE CASCADE,
    approved boolean NOT NULL,
    value bigint NOT NULL CONSTRAINT median_vote_value_positive_check CHECK (value >= 0),
    CONSTRAINT median_vote_poll_id_voter_id_key UNIQUE(poll_id, voter_id)
);

CREATE INDEX IF NOT EXISTS median_vote_voter_id_idx ON median_vote(voter_id);


CREATE TABLE IF NOT EXISTS basic_poll_vote (
    id uuid PRIMARY KEY,
    poll_id uuid NOT NULL REFERENCES basic_poll(id) ON DELETE CASCADE,
    voter_id uuid NOT NULL REFERENCES voter(id) ON DELETE CASCADE,
    approved boolean NOT NULL,
    option smallint NOT NULL CONSTRAINT basic_poll_vote_option_valid_check CHECK (option >= 0 AND option <= 2),
    CONSTRAINT basic_poll_vote_poll_id_voter_id_key UNIQUE(poll_id, voter_id)
);

CREATE INDEX IF NOT EXISTS basic_poll_vote_voter_id_idx ON basic_poll_vote(voter_id);


CREATE TABLE IF NOT EXISTS schulze_vote (
    id uuid PRIMARY KEY,
    poll_id uuid NOT NULL REFERENCES schulze_poll(id) ON DELETE CASCADE,
    voter_id uuid NOT NULL REFERENCES voter(id) ON DELETE CASCADE,
    approved boolean NOT NULL,
    CONSTRAINT schulze_vote_poll_id_voter_id_key UNIQUE(poll_id, voter_id)
);

CREATE INDEX IF NOT EXISTS schulze_vote_voter_id_idx ON schulze_vote(voter_id);


CREATE TABLE IF NOT EXISTS schulze_option_vote (
    id uuid PRIMARY KEY,
    vote_id uuid NOT NULL REFERENCES schulze_vote(id) ON DELETE CASCADE,
    option_id uuid NOT NULL REFERENCES schulze_option(id) ON DELETE CASCADE,
    sorting_position integer NOT NULL CONSTRAINT schulze_option_vote_sorting_position_positive_check CHECK (sorting_position >= 0),
    CONSTRAINT schulze_option_vote_vote_id_option_id_key UNIQUE(vote_id, option_id)
);

CREATE INDEX IF NOT EXISTS schulze_option_vote_vote_id_idx ON schulze_option_vote(vote_id);
CREATE INDEX IF NOT EXISTS schulze_option_vote_option_id_idx ON schulze_option_vote(option_id);


CREATE TABLE IF NOT EXISTS archived_median_result (
    id uuid PRIMARY KEY REFERENCES median_poll(id) ON DELETE CASCADE,
    value bigint NOT NULL CONSTRAINT archived_median_result_value_positive_check CHECK (value >= 0),
    archived_on timestamptz NOT NULL DEFAULT (timezone('utc', now()))
);

CREATE TABLE IF NOT EXISTS archived_basic_poll_result (
    id uuid PRIMARY KEY REFERENCES basic_poll(id) ON DELETE CASCADE,
    option smallint NOT NULL CONSTRAINT archived_basic_poll_result_option_valid_check CHECK (option >= 0 AND option <= 2),
    archived_on timestamptz NOT NULL DEFAULT (timezone('utc', now()))
);

CREATE TABLE IF NOT EXISTS archived_schulze_result (
    id uuid PRIMARY KEY REFERENCES schulze_poll(id) ON DELETE CASCADE,
    archived_on timestamptz NOT NULL DEFAULT (timezone('utc', now()))
);

CREATE TABLE IF NOT EXISTS archived_schulze_option_result (
    id uuid PRIMARY KEY,
    archived_schulze_result_id uuid NOT NULL REFERENCES archived_schulze_result(id) ON DELETE CASCADE,
    option_id uuid NOT NULL UNIQUE REFERENCES schulze_option(id) ON DELETE CASCADE,
    group_num integer NOT NULL CONSTRAINT archived_schulze_option_result_group_num_positive_check CHECK (group_num >= 0)
);

CREATE INDEX IF NOT EXISTS archived_schulze_option_result_archived_schulze_result_id_idx ON archived_schulze_option_result(archived_schulze_result_id);

COMMIT;
