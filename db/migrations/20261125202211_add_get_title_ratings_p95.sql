-- +goose Up
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION public.get_title_ratings_p95(
    p_start_date date DEFAULT NULL::date,
    p_end_date date DEFAULT NULL::date,
    p_rating_min numeric DEFAULT NULL::numeric,
    p_rating_max numeric DEFAULT NULL::numeric,
    p_min_count integer DEFAULT 1,
    p_lower_pct double precision DEFAULT 0.05,
    p_upper_pct double precision DEFAULT 0.95
)
    RETURNS TABLE(title character varying, media_type character varying, raw_rating numeric, normalized_rating numeric)
    LANGUAGE plpgsql
AS $function$
BEGIN
    RETURN QUERY
        WITH filtered_titles AS (
            SELECT
                t.id,
                t.title,
                t.media_type,
                t.is_finished
            FROM titles AS t
            WHERE t.is_dropped IS FALSE
              AND (p_start_date IS NULL OR t.date_actual >= p_start_date)
              AND (p_end_date   IS NULL OR t.date_actual <= p_end_date)
        ),
             global_stats AS (
                 SELECT
                     CASE
                         WHEN COUNT(ft.id) = 0 THEN 0::NUMERIC
                         ELSE (COUNT(ft.id) FILTER (WHERE ft.is_finished))::NUMERIC / COUNT(ft.id)
                         END AS f_coef
                 FROM filtered_titles AS ft
             ),
             base_ratings AS (
                 SELECT
                     ft.title,
                     ft.media_type,
                     (
                         COUNT(ft.id)::NUMERIC +
                         (COUNT(ft.id) FILTER (WHERE ft.is_finished))::NUMERIC * (SELECT gs.f_coef FROM global_stats gs)
                         ) AS raw_rating
                 FROM filtered_titles AS ft
                 GROUP BY
                     ft.title,
                     ft.media_type
                 HAVING
                     p_min_count IS NULL OR COUNT(ft.id) >= p_min_count
             ),
             percentile_bounds AS (
                 SELECT
                             PERCENTILE_CONT(p_lower_pct) WITHIN GROUP (ORDER BY br.raw_rating)::NUMERIC AS p_low,
                             PERCENTILE_CONT(p_upper_pct) WITHIN GROUP (ORDER BY br.raw_rating)::NUMERIC AS p_high
                 FROM base_ratings AS br
             ),
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
             ),
             min_max_bounds AS (
                 SELECT
                     MIN(cr.clamped_raw_rating) AS min_val,
                     MAX(cr.clamped_raw_rating) AS max_val
                 FROM clamped_ratings AS cr
             ),
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
        SELECT
            st.title,
            st.media_type,
            st.raw_rating,
            st.normalized_rating
        FROM scored_titles AS st
        WHERE (p_rating_min IS NULL OR st.raw_rating >= p_rating_min)
          AND (p_rating_max IS NULL OR st.raw_rating <= p_rating_max)
        ORDER BY
            st.media_type ASC,
            st.raw_rating DESC,
            st.title ASC;
END;
$function$;
-- +goose StatementEnd

-- +goose Down
DROP FUNCTION public.get_title_ratings_p95(date, date, numeric, numeric, int4, float8, float8);
