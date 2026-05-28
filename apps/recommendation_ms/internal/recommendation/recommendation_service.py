import logging
from recommendation.v1 import recommendation_pb2

logger = logging.getLogger("recommendation_ms")


def get_recommendations(
    profile_id: str, limit: int, top_rated_contents: list, db_pool, history_client
) -> list:

    recommendations_list = []

    # Exemplo temporário:
    # recommendations_list.append({
    #     "content_id": "00000000-0000-0000-0000-000000000000",
    #     "content_type": recommendation_pb2.ContentType.MOVIE,
    #     "score": 9.8,
    #     "reason": "Recomendação baseada em IA"
    # })

    return recommendations_list
