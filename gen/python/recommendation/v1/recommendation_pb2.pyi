from common.v1 import common_pb2 as _common_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from collections.abc import Iterable as _Iterable, Mapping as _Mapping
from typing import ClassVar as _ClassVar, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class GetRecommendationsRequest(_message.Message):
    __slots__ = ("profile_id", "limit", "top_rated_contents")
    PROFILE_ID_FIELD_NUMBER: _ClassVar[int]
    LIMIT_FIELD_NUMBER: _ClassVar[int]
    TOP_RATED_CONTENTS_FIELD_NUMBER: _ClassVar[int]
    profile_id: str
    limit: int
    top_rated_contents: _containers.RepeatedCompositeFieldContainer[TopRatedContent]
    def __init__(self, profile_id: _Optional[str] = ..., limit: _Optional[int] = ..., top_rated_contents: _Optional[_Iterable[_Union[TopRatedContent, _Mapping]]] = ...) -> None: ...

class TopRatedContent(_message.Message):
    __slots__ = ("content_id", "content_type", "genre_id", "avg_rating")
    CONTENT_ID_FIELD_NUMBER: _ClassVar[int]
    CONTENT_TYPE_FIELD_NUMBER: _ClassVar[int]
    GENRE_ID_FIELD_NUMBER: _ClassVar[int]
    AVG_RATING_FIELD_NUMBER: _ClassVar[int]
    content_id: str
    content_type: _common_pb2.ContentType
    genre_id: int
    avg_rating: float
    def __init__(self, content_id: _Optional[str] = ..., content_type: _Optional[_Union[_common_pb2.ContentType, str]] = ..., genre_id: _Optional[int] = ..., avg_rating: _Optional[float] = ...) -> None: ...

class RecommendedContent(_message.Message):
    __slots__ = ("content_id", "content_type", "score", "reason")
    CONTENT_ID_FIELD_NUMBER: _ClassVar[int]
    CONTENT_TYPE_FIELD_NUMBER: _ClassVar[int]
    SCORE_FIELD_NUMBER: _ClassVar[int]
    REASON_FIELD_NUMBER: _ClassVar[int]
    content_id: str
    content_type: _common_pb2.ContentType
    score: float
    reason: str
    def __init__(self, content_id: _Optional[str] = ..., content_type: _Optional[_Union[_common_pb2.ContentType, str]] = ..., score: _Optional[float] = ..., reason: _Optional[str] = ...) -> None: ...

class GetRecommendationsResponse(_message.Message):
    __slots__ = ("recommendations",)
    RECOMMENDATIONS_FIELD_NUMBER: _ClassVar[int]
    recommendations: _containers.RepeatedCompositeFieldContainer[RecommendedContent]
    def __init__(self, recommendations: _Optional[_Iterable[_Union[RecommendedContent, _Mapping]]] = ...) -> None: ...
