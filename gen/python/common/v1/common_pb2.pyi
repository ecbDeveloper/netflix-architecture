from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from typing import ClassVar as _ClassVar

DESCRIPTOR: _descriptor.FileDescriptor

class ContentType(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    CONTENT_TYPE_UNSPECIFIED: _ClassVar[ContentType]
    CONTENT_TYPE_MOVIE: _ClassVar[ContentType]
    CONTENT_TYPE_SERIES: _ClassVar[ContentType]
CONTENT_TYPE_UNSPECIFIED: ContentType
CONTENT_TYPE_MOVIE: ContentType
CONTENT_TYPE_SERIES: ContentType
