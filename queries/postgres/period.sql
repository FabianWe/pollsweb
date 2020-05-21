-- period.sql

-- Copyright 2020 Fabian Wenzelmann <fabianwen@posteo.eu>
-- Licensed under the Apache License, Version 2.0 (the "License");
-- you may not use this file except in compliance with the License.
-- You may obtain a copy of the License at
-- |
-- http://www.apache.org/licenses/LICENSE-2.0
-- |
-- Unless required by applicable law or agreed to in writing, software
-- distributed under the License is distributed on an "AS IS" BASIS,
-- WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
-- See the License for the specific language governing permissions and
-- limitations under the License.

-- name: period_add
INSERT INTO period (id, name, slug, meeting_time, period_start, period_end)
    VALUES ($1, $2, $3, $4, $5, $6);


-- name: period_get_by_id
SELECT * FROM period WHERE id = $1;


-- name: period_get_by_slug
SELECT * FROM period WHERE slug = $1;


-- name: period_get_latest
SELECT * FROM period ORDER BY period_start DESC LIMIT 1;


-- name: period_get_latest_n
SELECT * FROM period ORDER BY period_start DESC LIMIT $1;

--name: period_delete_by_id
DELETE FROM period WHERE id = $1;

--name: period_delete_by_slug
DELETE FROM period WHERE slug = $1;
