import os
import sys
import logging
from concurrent import futures
import grpc
import psycopg2
from psycopg2 import pool
from dotenv import load_dotenv

sys.path.append(
    os.path.abspath(os.path.join(os.path.dirname(__file__), "../../gen/python"))
)

from recommendation.v1 import recommendation_pb2  # type: ignore
from recommendation.v1.recommendation_pb2_grpc import (  # type: ignore
    RecommendationServiceServicer,
    add_RecommendationServiceServicer_to_server,
)
from history.v1.history_pb2_grpc import HistoryServiceStub  # type: ignore

from recommendation_service import get_recommendations  # type: ignore

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s - %(levelname)s - %(message)s",
    handlers=[logging.StreamHandler(sys.stdout)],
)
logger = logging.getLogger("recommendation_ms")


class RecommendationService(RecommendationServiceServicer):
    def __init__(self, db_pool, history_client):
        self.db_pool = db_pool
        self.history_client = history_client

    def GetRecommendations(self, request, context):
        try:
            results = get_recommendations(
                profile_id=request.profile_id,
                limit=request.limit,
                top_rated_contents=request.top_rated_contents,
                db_pool=self.db_pool,
                history_client=self.history_client,
            )

            recommendations = []
            for item in results:
                rec = recommendation_pb2.RecommendedContent(
                    content_id=item.get("content_id"),
                    content_type=item.get("content_type"),
                    score=item.get("score", 0.0),
                    reason=item.get("reason", ""),
                )
                recommendations.append(rec)

            return recommendation_pb2.GetRecommendationsResponse(
                recommendations=recommendations
            )
        except Exception as e:
            logger.error(f"Error in GetRecommendations: {e}")
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(str(e))
            return recommendation_pb2.GetRecommendationsResponse()


def initialize_database_connection():
    db_host = os.getenv("DB_HOST")
    db_port = os.getenv("DB_PORT")
    db_user = os.getenv("DB_USER")
    db_pass = os.getenv("DB_PASS")
    db_name = os.getenv("DB_NAME")

    logger.info(
        f"Connecting to database: host={db_host} port={db_port} user={db_user} dbname={db_name}"
    )
    try:
        connection_pool = pool.SimpleConnectionPool(
            1,
            10,
            host=db_host,
            port=db_port,
            user=db_user,
            password=db_pass,
            database=db_name,
        )
        # Test connection
        conn = connection_pool.getconn()
        with conn.cursor() as cursor:
            cursor.execute("SELECT 1")
        connection_pool.putconn(conn)
        logger.info("Successfully connected to database")
        return connection_pool
    except Exception as e:
        logger.error(f"failed to initialize db pool: {e}")
        sys.exit(1)


def main():
    load_dotenv()

    grpc_port = os.getenv("GRPC_PORT", "50052")

    db_pool = initialize_database_connection()

    history_host = os.getenv("HISTORY_GRPC_HOST", "history_ms")
    history_port = os.getenv("HISTORY_GRPC_PORT", "50051")
    history_addr = f"{history_host}:{history_port}"

    logger.info(f"Connecting to history service at {history_addr}")
    history_channel = grpc.insecure_channel(history_addr)
    history_client = HistoryServiceStub(history_channel)

    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    add_RecommendationServiceServicer_to_server(
        RecommendationService(db_pool, history_client), server
    )

    if os.getenv("ENV") == "development":
        from grpc_reflection.v1alpha import reflection  # type: ignore

        SERVICE_NAMES = (
            recommendation_pb2.DESCRIPTOR.services_by_name[
                "RecommendationService"
            ].full_name,
            reflection.SERVICE_NAME,
        )
        reflection.enable_server_reflection(SERVICE_NAMES, server)
        logger.info("gRPC reflection enabled")

    server.add_insecure_port(f"[::]:{grpc_port}")
    logger.info(f"recommendation microservice started on port {grpc_port}")

    try:
        server.start()
        server.wait_for_termination()
    except KeyboardInterrupt:
        logger.info("Stopping recommendation microservice...")
        server.stop(0)
        db_pool.closeall()


if __name__ == "__main__":
    main()
