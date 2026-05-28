from common.v1 import common_pb2 as _common_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from collections.abc import Iterable as _Iterable, Mapping as _Mapping
from typing import ClassVar as _ClassVar, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class RecordWatchHistoryRequest(_message.Message):
    __slots__ = ("profile_id", "genre_id", "movie_id", "episode_id", "last_position_seconds", "is_completed")
    PROFILE_ID_FIELD_NUMBER: _ClassVar[int]
    GENRE_ID_FIELD_NUMBER: _ClassVar[int]
    MOVIE_ID_FIELD_NUMBER: _ClassVar[int]
    EPISODE_ID_FIELD_NUMBER: _ClassVar[int]
    LAST_POSITION_SECONDS_FIELD_NUMBER: _ClassVar[int]
    IS_COMPLETED_FIELD_NUMBER: _ClassVar[int]
    profile_id: str
    genre_id: int
    movie_id: str
    episode_id: str
    last_position_seconds: int
    is_completed: bool
    def __init__(self, profile_id: _Optional[str] = ..., genre_id: _Optional[int] = ..., movie_id: _Optional[str] = ..., episode_id: _Optional[str] = ..., last_position_seconds: _Optional[int] = ..., is_completed: _Optional[bool] = ...) -> None: ...

class GetWatchHistoryRequest(_message.Message):
    __slots__ = ("id", "profile_id")
    ID_FIELD_NUMBER: _ClassVar[int]
    PROFILE_ID_FIELD_NUMBER: _ClassVar[int]
    id: str
    profile_id: str
    def __init__(self, id: _Optional[str] = ..., profile_id: _Optional[str] = ...) -> None: ...

class ListWatchHistoryRequest(_message.Message):
    __slots__ = ("profile_id",)
    PROFILE_ID_FIELD_NUMBER: _ClassVar[int]
    profile_id: str
    def __init__(self, profile_id: _Optional[str] = ...) -> None: ...

class UpdateWatchProgressRequest(_message.Message):
    __slots__ = ("id", "profile_id", "last_position_seconds", "is_completed")
    ID_FIELD_NUMBER: _ClassVar[int]
    PROFILE_ID_FIELD_NUMBER: _ClassVar[int]
    LAST_POSITION_SECONDS_FIELD_NUMBER: _ClassVar[int]
    IS_COMPLETED_FIELD_NUMBER: _ClassVar[int]
    id: str
    profile_id: str
    last_position_seconds: int
    is_completed: bool
    def __init__(self, id: _Optional[str] = ..., profile_id: _Optional[str] = ..., last_position_seconds: _Optional[int] = ..., is_completed: _Optional[bool] = ...) -> None: ...

class DeleteWatchHistoryRequest(_message.Message):
    __slots__ = ("id", "profile_id")
    ID_FIELD_NUMBER: _ClassVar[int]
    PROFILE_ID_FIELD_NUMBER: _ClassVar[int]
    id: str
    profile_id: str
    def __init__(self, id: _Optional[str] = ..., profile_id: _Optional[str] = ...) -> None: ...

class GetMostWatchedRequest(_message.Message):
    __slots__ = ("limit", "offset")
    LIMIT_FIELD_NUMBER: _ClassVar[int]
    OFFSET_FIELD_NUMBER: _ClassVar[int]
    limit: int
    offset: int
    def __init__(self, limit: _Optional[int] = ..., offset: _Optional[int] = ...) -> None: ...

class GetRecentlyWatchedRequest(_message.Message):
    __slots__ = ("profile_id", "limit", "offset")
    PROFILE_ID_FIELD_NUMBER: _ClassVar[int]
    LIMIT_FIELD_NUMBER: _ClassVar[int]
    OFFSET_FIELD_NUMBER: _ClassVar[int]
    profile_id: str
    limit: int
    offset: int
    def __init__(self, profile_id: _Optional[str] = ..., limit: _Optional[int] = ..., offset: _Optional[int] = ...) -> None: ...

class DeleteWatchHistoryResponse(_message.Message):
    __slots__ = ("success",)
    SUCCESS_FIELD_NUMBER: _ClassVar[int]
    success: bool
    def __init__(self, success: _Optional[bool] = ...) -> None: ...

class WatchHistory(_message.Message):
    __slots__ = ("id", "profile_id", "movie_id", "episode_id", "watched_at", "last_position_seconds", "is_completed", "genre_id")
    ID_FIELD_NUMBER: _ClassVar[int]
    PROFILE_ID_FIELD_NUMBER: _ClassVar[int]
    MOVIE_ID_FIELD_NUMBER: _ClassVar[int]
    EPISODE_ID_FIELD_NUMBER: _ClassVar[int]
    WATCHED_AT_FIELD_NUMBER: _ClassVar[int]
    LAST_POSITION_SECONDS_FIELD_NUMBER: _ClassVar[int]
    IS_COMPLETED_FIELD_NUMBER: _ClassVar[int]
    GENRE_ID_FIELD_NUMBER: _ClassVar[int]
    id: str
    profile_id: str
    movie_id: str
    episode_id: str
    watched_at: str
    last_position_seconds: int
    is_completed: bool
    genre_id: int
    def __init__(self, id: _Optional[str] = ..., profile_id: _Optional[str] = ..., movie_id: _Optional[str] = ..., episode_id: _Optional[str] = ..., watched_at: _Optional[str] = ..., last_position_seconds: _Optional[int] = ..., is_completed: _Optional[bool] = ..., genre_id: _Optional[int] = ...) -> None: ...

class ListWatchHistoryResponse(_message.Message):
    __slots__ = ("histories",)
    HISTORIES_FIELD_NUMBER: _ClassVar[int]
    histories: _containers.RepeatedCompositeFieldContainer[WatchHistory]
    def __init__(self, histories: _Optional[_Iterable[_Union[WatchHistory, _Mapping]]] = ...) -> None: ...

class GetRecentlyWatchedResponse(_message.Message):
    __slots__ = ("histories",)
    HISTORIES_FIELD_NUMBER: _ClassVar[int]
    histories: _containers.RepeatedCompositeFieldContainer[WatchHistory]
    def __init__(self, histories: _Optional[_Iterable[_Union[WatchHistory, _Mapping]]] = ...) -> None: ...

class RecordWatchHistoryResponse(_message.Message):
    __slots__ = ("watch_history",)
    WATCH_HISTORY_FIELD_NUMBER: _ClassVar[int]
    watch_history: WatchHistory
    def __init__(self, watch_history: _Optional[_Union[WatchHistory, _Mapping]] = ...) -> None: ...

class GetWatchHistoryResponse(_message.Message):
    __slots__ = ("watch_history",)
    WATCH_HISTORY_FIELD_NUMBER: _ClassVar[int]
    watch_history: WatchHistory
    def __init__(self, watch_history: _Optional[_Union[WatchHistory, _Mapping]] = ...) -> None: ...

class UpdateWatchProgressResponse(_message.Message):
    __slots__ = ("watch_history",)
    WATCH_HISTORY_FIELD_NUMBER: _ClassVar[int]
    watch_history: WatchHistory
    def __init__(self, watch_history: _Optional[_Union[WatchHistory, _Mapping]] = ...) -> None: ...

class MostWatchedItem(_message.Message):
    __slots__ = ("content_id", "content_type", "genre_id", "watch_count")
    CONTENT_ID_FIELD_NUMBER: _ClassVar[int]
    CONTENT_TYPE_FIELD_NUMBER: _ClassVar[int]
    GENRE_ID_FIELD_NUMBER: _ClassVar[int]
    WATCH_COUNT_FIELD_NUMBER: _ClassVar[int]
    content_id: str
    content_type: _common_pb2.ContentType
    genre_id: int
    watch_count: int
    def __init__(self, content_id: _Optional[str] = ..., content_type: _Optional[_Union[_common_pb2.ContentType, str]] = ..., genre_id: _Optional[int] = ..., watch_count: _Optional[int] = ...) -> None: ...

class GetMostWatchedResponse(_message.Message):
    __slots__ = ("items",)
    ITEMS_FIELD_NUMBER: _ClassVar[int]
    items: _containers.RepeatedCompositeFieldContainer[MostWatchedItem]
    def __init__(self, items: _Optional[_Iterable[_Union[MostWatchedItem, _Mapping]]] = ...) -> None: ...
