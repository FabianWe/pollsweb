-- voters_revision.sql

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

-- name: voters_revision_add
INSERT INTO voters_revision(id, period_id, note, is_active)
    VALUES ($1, $2, $3, $4);


-- name: voters_revision_get_by_id
SELECT * FROM voters_revision WHERE id = $1;


-- name: voters_revision_get_by_slug
SELECT * FROM voters_revision WHERE slug = $1;


-- name: voters_revision_get_for_period_id
SELECT * FROM voters_revision WHERE period_id = $1 ORDER BY created DESC;


-- name: voters_revision_get_for_period_slug
SELECT * FROM voters_revision INNER JOIN period
    ON (period.id = voters_revision.period_id)
    WHERE period.slug = $1;
