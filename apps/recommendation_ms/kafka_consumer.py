import json
import logging
import os
from confluent_kafka import Consumer, KafkaError  # type: ignore

logger = logging.getLogger("recommendation_ms.kafka")


def start_kafka_consumer(db_pool):
    kafka_broker = os.getenv("KAFKA_BROKER", "kafka:9092")
    topic = os.getenv("KAFKA_TOPIC", "netflix_history.public.watch_histories")

    conf = {
        "bootstrap.servers": kafka_broker,
        "group.id": "recommendation_group",
        "auto.offset.reset": "earliest",
        "enable.auto.commit": False,
    }

    consumer = Consumer(conf)
    consumer.subscribe([topic])
    logger.info("Started Kafka consumer on topic %s", topic)

    try:
        while True:
            msg = consumer.poll(timeout=1.0)
            if msg is None:
                continue
            if msg.error():
                if msg.error().code() == KafkaError._PARTITION_EOF:
                    continue
                else:
                    logger.error("Kafka error: %s", msg.error())
                    break

            try:
                val = json.loads(msg.value().decode("utf-8"))

                if not val or "payload" not in val:
                    continue

                payload = val["payload"]
                op = payload.get("op")

                if op in ("c", "u", "r"):
                    after = payload.get("after")
                    if after:
                        record_id = after.get("id")
                        profile_id = after.get("profile_id")
                        movie_id = after.get("movie_id")
                        episode_id = after.get("episode_id")
                        genre_id = after.get("genre_id")
                        watched_at = after.get("watched_at")
                        is_completed = after.get("is_completed")

                        if profile_id and record_id:
                            with db_pool.getconn() as conn:
                                with conn.cursor() as cur:
                                    cur.execute(
                                        """
                                        INSERT INTO recommendation_history 
                                        (id, profile_id, movie_id, episode_id, genre_id, watched_at, is_completed)
                                        VALUES (%s, %s, %s, %s, %s, %s, %s)
                                        ON CONFLICT (id) DO UPDATE SET
                                            movie_id = EXCLUDED.movie_id,
                                            episode_id = EXCLUDED.episode_id,
                                            genre_id = EXCLUDED.genre_id,
                                            watched_at = EXCLUDED.watched_at,
                                            is_completed = EXCLUDED.is_completed
                                    """,
                                        (
                                            record_id,
                                            profile_id,
                                            movie_id,
                                            episode_id,
                                            genre_id,
                                            watched_at,
                                            is_completed,
                                        ),
                                    )
                                conn.commit()
                                db_pool.putconn(conn)
                            logger.info(
                                f"Processed upsert for profile {profile_id}, record {record_id}"
                            )

                elif op == "d":
                    before = payload.get("before")
                    if before:
                        record_id = before.get("id")
                        if record_id:
                            with db_pool.getconn() as conn:
                                with conn.cursor() as cur:
                                    cur.execute(
                                        "DELETE FROM recommendation_history WHERE id = %s",
                                        (record_id,),
                                    )
                                conn.commit()
                                db_pool.putconn(conn)
                            logger.info(f"Processed delete for record {record_id}")

                consumer.commit(msg)

            except Exception as e:
                logger.error("Failed to process message: %s", e)

    except KeyboardInterrupt:
        pass
    finally:
        consumer.close()
