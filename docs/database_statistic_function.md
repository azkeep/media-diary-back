```sql
CREATE OR REPLACE FUNCTION public.get_title_ratings_p95(
    p_start_date date DEFAULT NULL::date,
    p_end_date date DEFAULT NULL::date,
    p_rating_min numeric DEFAULT NULL::numeric,
    p_rating_max numeric DEFAULT NULL::numeric,
    p_min_count integer DEFAULT 1,
    p_lower_pct double precision DEFAULT 0.05,
    p_upper_pct double precision DEFAULT 0.95)
    RETURNS TABLE
            (
                title             character varying,
                media_type        character varying,
                raw_rating        numeric,
                normalized_rating numeric
            )
    LANGUAGE plpgsql
AS
$function$
BEGIN
    RETURN QUERY
        WITH filtered_titles AS (SELECT t.id,
                                        t.title,
                                        t.media_type,
                                        t.is_finished
                                 FROM titles AS t
                                 WHERE t.is_dropped IS FALSE
                                   AND (p_start_date IS NULL OR t.date_actual >= p_start_date)
                                   AND (p_end_date IS NULL OR t.date_actual <= p_end_date)),
             global_stats AS (SELECT CASE
                                         WHEN COUNT(ft.id) = 0 THEN 0::NUMERIC
                                         ELSE (COUNT(ft.id) FILTER (WHERE ft.is_finished))::NUMERIC / COUNT(ft.id)
                                         END AS f_coef
                              FROM filtered_titles AS ft),
             base_ratings AS (SELECT ft.title,
                                     ft.media_type,
                                     (
                                         COUNT(ft.id)::NUMERIC +
                                         (COUNT(ft.id) FILTER (WHERE ft.is_finished))::NUMERIC *
                                         (SELECT gs.f_coef FROM global_stats gs)
                                         ) AS raw_rating
                              FROM filtered_titles AS ft
                              GROUP BY ft.title,
                                       ft.media_type
                              HAVING p_min_count IS NULL
                                  OR COUNT(ft.id) >= p_min_count),
             percentile_bounds
                 AS (SELECT PERCENTILE_CONT(p_lower_pct) WITHIN GROUP (ORDER BY br.raw_rating)::NUMERIC AS p_low,
                            PERCENTILE_CONT(p_upper_pct) WITHIN GROUP (ORDER BY br.raw_rating)::NUMERIC AS p_high
                     FROM base_ratings AS br),
             clamped_ratings AS (SELECT br.title,
                                        br.media_type,
                                        br.raw_rating,
                                        CASE
                                            WHEN pb.p_low IS NULL OR pb.p_high IS NULL THEN NULL::NUMERIC
                                            ELSE GREATEST(pb.p_low, LEAST(pb.p_high, br.raw_rating))
                                            END AS clamped_raw_rating
                                 FROM base_ratings AS br
                                          CROSS JOIN percentile_bounds AS pb),
             min_max_bounds AS (SELECT MIN(cr.clamped_raw_rating) AS min_val,
                                       MAX(cr.clamped_raw_rating) AS max_val
                                FROM clamped_ratings AS cr),
             scored_titles AS (SELECT cr.title,
                                      cr.media_type,
                                      ROUND(cr.raw_rating, 2) AS raw_rating,
                                      CASE
                                          WHEN mm.min_val IS NULL THEN NULL::NUMERIC
                                          WHEN mm.min_val = mm.max_val THEN 5.5::NUMERIC
                                          ELSE ROUND(
                                                  1.0 + (cr.clamped_raw_rating - mm.min_val) * 9.0 /
                                                        (mm.max_val - mm.min_val),
                                                  2
                                               )
                                          END                 AS normalized_rating
                               FROM clamped_ratings AS cr
                                        CROSS JOIN min_max_bounds AS mm)
        SELECT st.title,
               st.media_type,
               st.raw_rating,
               st.normalized_rating
        FROM scored_titles AS st
        WHERE (p_rating_min IS NULL OR st.raw_rating >= p_rating_min)
          AND (p_rating_max IS NULL OR st.raw_rating <= p_rating_max)
        ORDER BY st.media_type ASC,
                 st.raw_rating DESC,
                 st.title ASC;
END;
$function$
;

```

Here is a detailed, step-by-step walkthrough of what happens inside the **refactored** `public.get_title_ratings_p95`
function—highlighting what stayed the same and what changed with the new parameters and CTEs.

---

### 1. Function Signature & Parameters

The function is written in `plpgsql` and returns a `TABLE` with four columns: `title`, `media_type`, `raw_rating`,
and `normalized_rating`. It now takes **7 optional parameters** (all with defaults):

* **Date filters:** `p_start_date` and `p_end_date` (default `NULL`) constrain `titles.date_actual`.
* **Result rating filters:** `p_rating_min` and `p_rating_max` (default `NULL`) constrain which rows are **returned at
  the end** based on `raw_rating`.
* **Noise & outlier controls:**
* `p_min_count` (default `1`): Minimum row count required for a `(title, media_type)` group to survive aggregation.
* `p_lower_pct` (default `0.05`) & `p_upper_pct` (default `0.95`): Lower and upper quantile thresholds (`0.0` to `1.0`)
  used for clamping `raw_rating` before normalization.

Because the body is a single `RETURN QUERY SELECT ...`, PostgreSQL plans and executes all CTEs in one query pass.

---

### 2. CTE 1: `filtered_titles` (Base Row & Date Filtering)

```sql
filtered_titles AS (
    SELECT t.id, t.title, t.media_type, t.is_finished
    FROM titles AS t
    WHERE t.is_dropped IS FALSE
      AND (p_start_date IS NULL OR t.date_actual >= p_start_date)
      AND (p_end_date   IS NULL OR t.date_actual <= p_end_date)
)

```

* Reads rows from `titles` (`t`).
* Keeps only non-dropped rows (`t.is_dropped IS FALSE`; `NULL` values are excluded).
* Applies `p_start_date` and `p_end_date` against `t.date_actual` whenever either parameter is not `NULL`.

---

### 3. CTE 2: `global_stats` (Completion Weight `f_coef`)

```sql
global_stats AS (
    SELECT
        CASE
            WHEN COUNT(ft.id) = 0 THEN 0::NUMERIC
            ELSE (COUNT(ft.id) FILTER (WHERE ft.is_finished))::NUMERIC / COUNT(ft.id)
        END AS f_coef
    FROM filtered_titles AS ft
)

```

* Aggregates across **all** rows in `filtered_titles` to produce a single completion ratio (`f_coef`):

$$\text{f_coef} = \frac{\text{COUNT} (ft.id)\ \text{FILTER (WHERE } ft.is\_finished\text{)}}{\text{COUNT} (ft.id)}$$

* Returns `0` if `filtered_titles` is empty (`COUNT = 0`), avoiding division by zero.

---

### 4. CTE 3: `base_ratings` (Grouping, `p_min_count`, and `raw_rating`)

```sql
base_ratings AS (
    SELECT
        ft.title,
        ft.media_type,
        (
            COUNT(ft.id)::NUMERIC +
            (COUNT(ft.id) FILTER (WHERE ft.is_finished))::NUMERIC * (SELECT gs.f_coef FROM global_stats gs)
        ) AS raw_rating
    FROM filtered_titles AS ft
    GROUP BY ft.title, ft.media_type
    HAVING p_min_count IS NULL OR COUNT(ft.id) >= p_min_count
)

```

* Groups `filtered_titles` by `(ft.title, ft.media_type)`.
* **New (`p_min_count`):** Filters out groups with fewer than `p_min_count` entries before calculating percentiles or
  bounds.
* Calculates the **unrounded** `raw_rating` for each surviving group:

$$\text{raw\_rating} = N_{\text{group}} + \bigl (N_{\text{finished, group}} \times \text{f\_coef}\bigr)$$

where $N_{\text{group}}$ is `COUNT(ft.id)` and $N_{\text{finished, group}}$ is the count of finished entries in that
group.

* Note that `p_rating_min` and `p_rating_max` are **no longer** in `HAVING` here—so filtering out low- or high-scoring
  items will not alter the scale bounds.

---

### 5. CTE 4: `percentile_bounds` (Quantile Thresholds `p_low` & `p_high`)

```sql
percentile_bounds AS (
    SELECT
        PERCENTILE_CONT(p_lower_pct) WITHIN GROUP (ORDER BY br.raw_rating)::NUMERIC AS p_low,
        PERCENTILE_CONT(p_upper_pct) WITHIN GROUP (ORDER BY br.raw_rating)::NUMERIC AS p_high
    FROM base_ratings AS br
)

```

* **New:** Computes continuous interpolated percentiles (`PERCENTILE_CONT`) across `br.raw_rating` for all groups in
  `base_ratings`.
* Produces a single row with `p_low` (at `p_lower_pct`, default 5th percentile) and `p_high` (at `p_upper_pct`, default
  95th percentile), cast to `NUMERIC`.
* If `base_ratings` has no rows, both `p_low` and `p_high` evaluate to `NULL`.

---

### 6. CTE 5: `clamped_ratings` (Winsorizing / Clamping `raw_rating`)

```sql
clamped_ratings AS (
    SELECT
        br.title,
        br.media_type,
        br.raw_rating,
        CASE
            WHEN pb.p_low IS NULL OR pb.p_high IS NULL THEN NULL::NUMERIC
            ELSE GREATEST(pb.p_low, LEAST(pb.p_high, br.raw_rating))
        END AS clamped_raw_rating
    FROM base_ratings AS br
    CROSS JOIN percentile_bounds AS pb
)

```

* **New:** Cross-joins each row in `base_ratings` with the single row from `percentile_bounds`.
* Keeps the original `br.raw_rating` intact, while creating `clamped_raw_rating`:
* `LEAST(pb.p_high, br.raw_rating)` caps top outliers at `p_high`.
* `GREATEST(pb.p_low, ...)` lifts bottom values up to `p_low`.


* Because outliers are capped rather than deleted, an extreme spike still sits at the top of the clamped range without
  forcing the runner-up to take over the ceiling.

---

### 7. CTE 6: `min_max_bounds` (Scale Range from Clamped Values)

```sql
min_max_bounds AS (
    SELECT
        MIN(cr.clamped_raw_rating) AS min_val,
        MAX(cr.clamped_raw_rating) AS max_val
    FROM clamped_ratings AS cr
)

```

* Computes `min_val` and `max_val` over `cr.clamped_raw_rating` (which equal `p_low` and `p_high` whenever at least one
  row exists).
* Because this happens **before** `p_rating_min` / `p_rating_max` are applied, the normalization scale reflects the
  broader clamped distribution rather than a narrowed result slice.

---

### 8. CTE 7: `scored_titles` (Rounding & `1.00–10.00` Normalization)

```sql
scored_titles AS (
    SELECT
        cr.title,
        cr.media_type,
        ROUND(cr.raw_rating, 2) AS raw_rating,
        CASE
            WHEN mm.min_val IS NULL THEN NULL::NUMERIC
            WHEN mm.min_val = mm.max_val THEN 5.5::NUMERIC
            ELSE ROUND(
                1.0 + (cr.clamped_raw_rating - mm.min_val) * 9.0 / (mm.max_val - mm.min_val),
                2
            )
        END AS normalized_rating
    FROM clamped_ratings AS cr
    CROSS JOIN min_max_bounds AS mm
)

```

* Rounds the true (unclamped) `cr.raw_rating` to 2 decimal places for display.
* Computes `normalized_rating` from `cr.clamped_raw_rating` onto a `1.00` to `10.00` scale:
* `NULL` if `mm.min_val IS NULL` (no matching groups).
* `5.5` (midpoint) if `mm.min_val = mm.max_val` (zero spread / identical clamped ratings, avoiding division by zero).
* Otherwise:

$$\text{normalized\_rating} = \text{ROUND}\!\left (1.0 + \frac{ (\text{clamped\_raw\_rating} - \text{min\_val}) \times 9.0}{\text{max\_val} - \text{min\_val}},\, 2\right)$$





---

### 9. Final `SELECT`, Post-Scale Rating Filter & `ORDER BY`

```sql
SELECT st.title,
       st.media_type,
       st.raw_rating,
       st.normalized_rating
FROM scored_titles AS st
WHERE (p_rating_min IS NULL OR st.raw_rating >= p_rating_min)
  AND (p_rating_max IS NULL OR st.raw_rating <= p_rating_max)
ORDER BY st.media_type ASC,
         st.normalized_rating DESC,
         st.title ASC;

```

* **Changed:** Applies `p_rating_min` and `p_rating_max` here—after normalization—so filtering returned rows narrows the
  output without rescaling the remaining items to `1.00–10.00`. *(Note: `st.raw_rating` is already rounded to 2 decimal
  places at this point.)*
* Sorts the final rows by `media_type` ascending, `normalized_rating` descending, and `title` ascending.