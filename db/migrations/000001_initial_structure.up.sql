CREATE TABLE IF NOT EXISTS periods
(
    id           uuid PRIMARY KEY,
    name         text UNIQUE NOT NULL,
    slug         text UNIQUE NOT NULL,
    meeting_time text,
    period_start timestamptz NOT NULL,
    period_end   timestamptz NOT NULL,
    created      timestamptz NOT NULL DEFAULT (timezone('utc', now())),
    updated      timestamptz NOT NULL DEFAULT (timezone('utc', now()))
);

CREATE INDEX IF NOT EXISTS periods_period_start_idx ON periods (period_start DESC);

CREATE TABLE IF NOT EXISTS voter_revisions
(
    id        uuid PRIMARY KEY,
    period_id uuid        NOT NULL REFERENCES periods (id) ON DELETE CASCADE,
    name      text        NOT NULL,
    slug      text        NOT NULL,
    note      text        NOT NULL,
    is_active boolean     NOT NULL DEFAULT true,
    created   timestamptz NOT NULL DEFAULT (timezone('utc', now())),

    CONSTRAINT voter_revisions_period_id_name_key UNIQUE (period_id, name),
    CONSTRAINT voter_revisions_period_id_slug_key UNIQUE (period_id, slug)
);

CREATE INDEX IF NOT EXISTS voter_revisions_period_id_idx ON voter_revisions (period_id);

CREATE TABLE IF NOT EXISTS voters
(
    id          uuid PRIMARY KEY,
    revision_id uuid    NOT NULL REFERENCES voter_revisions (id) ON DELETE CASCADE,
    name        text    NOT NULL,
    slug        text    NOT NULL,
    weight      integer NOT NULL,
    CONSTRAINT voters_weight_positive_check CHECK (weight >= 0),

    CONSTRAINT voters_revision_id_name_key UNIQUE (revision_id, name),
    CONSTRAINT voters_revision_id_revision_slug_key UNIQUE (revision_id, slug)
);

CREATE INDEX IF NOT EXISTS voters_revision_id_idx ON voters (revision_id);
CREATE INDEX IF NOT EXISTS voters_name_idx ON voters (name);

CREATE TABLE IF NOT EXISTS poll_collections
(
    id           uuid PRIMARY KEY,
    revision_id  uuid        NOT NULL REFERENCES voter_revisions (id) ON DELETE CASCADE,
    name         text        NOT NULL,
    slug         text        NOT NULL,
    meeting_time timestamptz NOT NULL,
    online_start timestamptz,
    online_end   timestamptz,

    CONSTRAINT poll_collections_revision_id_name_key UNIQUE (revision_id, name),
    CONSTRAINT poll_collections_revision_id_slug_key UNIQUE (revision_id, slug)
);

CREATE INDEX IF NOT EXISTS poll_collections_revision_id_idx ON poll_collections (revision_id);
CREATE INDEX IF NOT EXISTS poll_collections_meeting_time_idx ON poll_collections (meeting_time);
CREATE INDEX IF NOT EXISTS poll_collections_online_start_online_end_idx ON poll_collections (online_start, online_end);

CREATE TABLE IF NOT EXISTS poll_groups
(
    id            uuid PRIMARY KEY,
    collection_id uuid    NOT NULL REFERENCES poll_collections (id) ON DELETE CASCADE,
    name          text    NOT NULL,
    slug          text    NOT NULL,
    group_num     integer NOT NULL,

    CONSTRAINT poll_groups_group_num_positive_check CHECK (group_num >= 0),
    CONSTRAINT poll_groups_collection_id_name_key UNIQUE (collection_id, name),
    CONSTRAINT poll_groups_collection_id_slug_key UNIQUE (collection_id, slug),
    CONSTRAINT poll_groups_collection_id_group_num_key UNIQUE (collection_id, group_num)
);

CREATE INDEX IF NOT EXISTS poll_groups_collection_id_idx ON poll_groups (collection_id);

CREATE TABLE IF NOT EXISTS polls_base
(
    id                uuid PRIMARY KEY,
    group_id          uuid    NOT NULL REFERENCES poll_groups (id) ON DELETE CASCADE,
    poll_num          integer NOT NULL,
    name              text    NOT NULL,
    slug              text    NOT NULL,
    majority          text    NOT NULL,
    absolute_majority boolean NOT NULL,
    CONSTRAINT poll_base_group_id_poll_num_key UNIQUE (group_id, poll_num),
    CONSTRAINT poll_base_poll_num_positive_check CHECK (poll_num >= 0),
    CONSTRAINT poll_base_group_id_name_key UNIQUE (group_id, name),
    CONSTRAINT poll_base_group_id_slug_key UNIQUE (group_id, slug)
);

CREATE INDEX IF NOT EXISTS polls_base_group_id_idx ON polls_base (group_id);
CREATE INDEX IF NOT EXISTS polls_base_name_idx ON polls_base (name);


CREATE TABLE IF NOT EXISTS median_polls
(
    id       uuid PRIMARY KEY REFERENCES polls_base (id) ON DELETE CASCADE,
    value    integer NOT NULL,
    currency text    NOT NULL,
    CONSTRAINT median_polls_value_positive_check CHECK (value >= 0)
);

CREATE TABLE IF NOT EXISTS basic_polls
(
    id uuid PRIMARY KEY REFERENCES polls_base (id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS schulze_polls
(
    id uuid PRIMARY KEY REFERENCES polls_base (id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS schulze_options
(
    id              uuid PRIMARY KEY,
    schulze_poll_id uuid    NOT NULL REFERENCES schulze_polls (id) ON DELETE CASCADE,
    option          text    NOT NULL,
    option_num      integer NOT NULL,
    CONSTRAINT schulze_options_option_num_positive_check CHECK (option_num >= 0),
    CONSTRAINT schulze_options_schulze_poll_id_option_key UNIQUE (schulze_poll_id, option),
    CONSTRAINT schulze_options_schulze_poll_id_option_num_key UNIQUE (schulze_poll_id, option_num)
);

CREATE INDEX IF NOT EXISTS schulze_options_schulze_poll_id_idx ON schulze_options (schulze_poll_id);

CREATE TABLE IF NOT EXISTS median_votes
(
    id       uuid PRIMARY KEY,
    poll_id  uuid    NOT NULL REFERENCES median_polls (id) ON DELETE CASCADE,
    voter_id uuid    NOT NULL REFERENCES voters (id) ON DELETE CASCADE,
    approved boolean NOT NULL,
    value    integer NOT NULL,
    CONSTRAINT median_votes_value_positive_check CHECK (value >= 0),
    CONSTRAINT median_votes_poll_id_voter_id_key UNIQUE (poll_id, voter_id)
);

CREATE INDEX IF NOT EXISTS median_votes_poll_id_idx ON median_votes (poll_id);
CREATE INDEX IF NOT EXISTS median_votes_voter_id_idx ON median_votes (voter_id);

CREATE TABLE IF NOT EXISTS basic_poll_votes
(
    id       uuid PRIMARY KEY,
    poll_id  uuid       NOT NULL REFERENCES basic_polls (id) ON DELETE CASCADE,
    voter_id uuid       NOT NULL REFERENCES voters (id) ON DELETE CASCADE,
    approved boolean    NOT NULL,
    option   varchar(1) NOT NULL,

    CONSTRAINT basic_poll_votes_option_valid_check CHECK (option = 'y' OR option = 'n' OR option = 'a'),
    CONSTRAINT basic_poll_votes_poll_id_voter_id_key UNIQUE (poll_id, voter_id)
);

CREATE INDEX IF NOT EXISTS basic_poll_votes_poll_id_idx ON basic_poll_votes (poll_id);
CREATE INDEX IF NOT EXISTS basic_poll_votes_voter_id_idx ON basic_poll_votes (voter_id);

CREATE TABLE IF NOT EXISTS schulze_votes
(
    id       uuid PRIMARY KEY,
    poll_id  uuid    NOT NULL REFERENCES schulze_polls (id) ON DELETE CASCADE,
    voter_id uuid    NOT NULL REFERENCES voters (id) ON DELETE CASCADE,
    approved boolean NOT NULL,
    CONSTRAINT schulze_votes_poll_id_voter_id_key UNIQUE (poll_id, voter_id)
);

CREATE INDEX IF NOT EXISTS schulze_votes_poll_id_idx ON schulze_votes (voter_id);
CREATE INDEX IF NOT EXISTS schulze_votes_voter_id_idx ON schulze_votes (voter_id);

CREATE TABLE IF NOT EXISTS schulze_option_votes
(
    id               uuid PRIMARY KEY,
    vote_id          uuid    NOT NULL REFERENCES schulze_votes (id) ON DELETE CASCADE,
    option_id        uuid    NOT NULL REFERENCES schulze_options (id) ON DELETE CASCADE,
    sorting_position integer NOT NULL,

    CONSTRAINT schulze_option_votes_sorting_position_positive_check CHECK (sorting_position >= 0),
    CONSTRAINT schulze_option_votes_vote_id_option_id_key UNIQUE (vote_id, option_id)
);

CREATE INDEX IF NOT EXISTS schulze_option_votes_vote_id_idx ON schulze_option_votes (vote_id);
CREATE INDEX IF NOT EXISTS schulze_option_votes_option_id_idx ON schulze_option_votes (option_id);
