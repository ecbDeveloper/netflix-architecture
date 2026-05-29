import logging
import uuid
from datetime import datetime, timezone

from common.v1 import common_pb2  # type: ignore
from history.v1 import history_pb2  # type: ignore

logger = logging.getLogger("recommendation_ms")

W_GENRE = 0.55
W_RATING = 0.30
W_RECENCY = 0.15

RECENCY_PENALTY_DAYS = 30
COMPLETED_GENRE_BOOST = 1.5
HISTORY_FETCH_LIMIT = 200
HISTORY_FETCH_OFFSET = 0


def get_recommendations(
    profile_id: str,
    limit: int,
    top_rated_contents: list,
    db_pool,
    history_client,
) -> list:
    if not top_rated_contents:
        logger.info("no candidates provided", extra={"profile_id": profile_id})
        return []

    watch_history = _fetch_watch_history(profile_id, history_client)
    genre_profile = _build_genre_profile(watch_history)
    watched_content_ids = _extract_watched_content_ids(watch_history)
    recently_watched_ids = _extract_recently_watched_ids(
        watch_history, days=RECENCY_PENALTY_DAYS
    )
    rating_map = _build_normalized_rating_map(top_rated_contents)

    scored = []
    for content in top_rated_contents:
        content_id = content.content_id

        if content_id in recently_watched_ids:
            continue

        genre_score = _genre_affinity_score(content.genre_id, genre_profile)
        rating_score = rating_map.get(content_id, 0.0)
        recency_penalty = 1.0 if content_id not in watched_content_ids else 0.5

        final_score = (
            W_GENRE * genre_score
            + W_RATING * rating_score
            + W_RECENCY * recency_penalty
        )

        reason = _build_reason(
            genre_score, rating_score, content.genre_id, genre_profile
        )

        scored.append(
            {
                "content_id": content_id,
                "content_type": content.content_type,
                "score": round(final_score, 4),
                "reason": reason,
            }
        )

    scored.sort(key=lambda x: x["score"], reverse=True)
    recommendations = scored[:limit]

    _persist_to_cache(profile_id, recommendations, db_pool)

    logger.info(
        "recommendations generated",
        extra={
            "profile_id": profile_id,
            "count": len(recommendations),
            "history_size": len(watch_history),
            "unique_genres": len(genre_profile),
        },
    )

    return recommendations


def _fetch_watch_history(profile_id: str, history_client) -> list:
    try:
        request = history_pb2.GetRecentlyWatchedRequest(
            profile_id=profile_id,
            limit=HISTORY_FETCH_LIMIT,
            offset=HISTORY_FETCH_OFFSET,
        )
        response = history_client.GetRecentlyWatched(request)
        return list(response.histories)
    except Exception as exc:
        logger.warning(
            "failed to fetch watch history, falling back to rating-only",
            extra={"profile_id": profile_id, "error": str(exc)},
        )
        return []


def _build_genre_profile(watch_history: list) -> dict[int, float]:
    raw: dict[int, float] = {}

    for item in watch_history:
        genre_id = item.genre_id
        if genre_id is None:
            continue
        weight = COMPLETED_GENRE_BOOST if item.is_completed else 1.0
        raw[genre_id] = raw.get(genre_id, 0.0) + weight

    if not raw:
        return {}

    max_count = max(raw.values())
    return {genre_id: count / max_count for genre_id, count in raw.items()}


def _extract_watched_content_ids(watch_history: list) -> set[str]:
    ids: set[str] = set()
    for item in watch_history:
        if item.movie_id:
            ids.add(item.movie_id)
        if item.episode_id:
            ids.add(item.episode_id)
    return ids


def _extract_recently_watched_ids(watch_history: list, days: int) -> set[str]:
    cutoff = datetime.now(timezone.utc).timestamp() - days * 86400
    recent: set[str] = set()

    for item in watch_history:
        try:
            watched_ts = _parse_timestamp(item.watched_at)
        except Exception:
            continue

        if watched_ts >= cutoff:
            if item.movie_id:
                recent.add(item.movie_id)
            if item.episode_id:
                recent.add(item.episode_id)

    return recent


def _parse_timestamp(watched_at: str) -> float:
    formats = [
        "%Y-%m-%dT%H:%M:%SZ",
        "%Y-%m-%dT%H:%M:%S.%fZ",
        "%Y-%m-%dT%H:%M:%S%z",
        "%Y-%m-%dT%H:%M:%S.%f%z",
        "%Y-%m-%d %H:%M:%S",
    ]
    for fmt in formats:
        try:
            dt = datetime.strptime(watched_at, fmt)
            if dt.tzinfo is None:
                dt = dt.replace(tzinfo=timezone.utc)
            return dt.timestamp()
        except ValueError:
            continue
    raise ValueError(f"unrecognized date format: {watched_at!r}")


def _build_normalized_rating_map(top_rated_contents: list) -> dict[str, float]:
    if not top_rated_contents:
        return {}

    ratings = [c.avg_rating for c in top_rated_contents]
    min_r, max_r = min(ratings), max(ratings)

    if max_r == min_r:
        return {c.content_id: 1.0 for c in top_rated_contents}

    span = max_r - min_r
    return {c.content_id: (c.avg_rating - min_r) / span for c in top_rated_contents}


def _genre_affinity_score(genre_id: int, genre_profile: dict[int, float]) -> float:
    if not genre_profile:
        return 0.5
    return genre_profile.get(genre_id, 0.1)


def _build_reason(
    genre_score: float,
    rating_score: float,
    genre_id: int,
    genre_profile: dict[int, float],
) -> str:
    is_top_genre = genre_id in genre_profile and genre_profile[genre_id] >= 0.7
    is_high_rated = rating_score >= 0.8

    if is_top_genre and is_high_rated:
        reason = "Top no seu gênero favorito"
    elif is_top_genre:
        reason = "Baseado no seu histórico"
    elif is_high_rated:
        reason = "Muito bem avaliado"
    else:
        reason = "Recomendado para você"

    return reason[:50]


def _persist_to_cache(profile_id: str, recommendations: list, db_pool) -> None:
    if not recommendations:
        return

    conn = None
    try:
        conn = db_pool.getconn()
        with conn.cursor() as cur:
            cur.execute(
                "DELETE FROM recommendations WHERE profile_id = %s",
                (profile_id,),
            )

            type_map = {
                common_pb2.ContentType.CONTENT_TYPE_MOVIE: "MOVIE",
                common_pb2.ContentType.CONTENT_TYPE_SERIES: "SERIES",
            }

            for rec in recommendations:
                content_type_str = type_map.get(rec["content_type"], "MOVIE")
                cur.execute(
                    """
                    INSERT INTO recommendations (id, profile_id, content_id, content_type, score, reason)
                    VALUES (%s, %s, %s, %s::content_type_enum, %s, %s)
                    """,
                    (
                        str(uuid.uuid4()),
                        profile_id,
                        rec["content_id"],
                        content_type_str,
                        rec["score"],
                        rec["reason"],
                    ),
                )
        conn.commit()
    except Exception as exc:
        logger.error(
            "failed to persist recommendations cache",
            extra={"profile_id": profile_id, "error": str(exc)},
        )
        if conn:
            try:
                conn.rollback()
            except Exception:
                pass
    finally:
        if conn:
            db_pool.putconn(conn)
