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

DROP TABLE IF EXISTS archived_schulze_option_result;
DROP TABLE IF EXISTS archived_schulze_result;
DROP TABLE IF EXISTS archived_basic_poll_result;
DROP TABLE IF EXISTS archived_median_result;
DROP TABLE IF EXISTS schulze_option_vote;
DROP TABLE IF EXISTS schulze_vote;
DROP TABLE IF EXISTS basic_poll_vote;
DROP TABLE IF EXISTS median_vote;
DROP TABLE IF EXISTS schulze_option;
DROP TABLE IF EXISTS schulze_poll;
DROP TABLE IF EXISTS basic_poll;
DROP TABLE IF EXISTS median_poll;
DROP TABLE IF EXISTS poll_base;
DROP TABLE IF EXISTS poll_group;
DROP TABLE IF EXISTS poll_collection;
DROP TABLE IF EXISTS voter;
DROP TABLE IF EXISTS voters_revision;
DROP TABLE IF EXISTS period;
DROP TYPE IF EXISTS gopolls_fraction;

COMMIT;
