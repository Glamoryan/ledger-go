"""psycopg extensions to the DBAPI-2.0

This module holds all the extensions to the DBAPI-2.0 provided by psycopg.

- `connection` -- the new-type inheritable connection class
- `cursor` -- the new-type inheritable cursor class
- `lobject` -- the new-type inheritable large object class
- `adapt()` -- exposes the PEP-246_ compatible adapting mechanism used
  by psycopg to adapt Python types to PostgreSQL ones

.. _PEP-246: https://www.python.org/dev/peps/pep-0246/
"""

import re as _re

from psycopg2._psycopg import (                             # noqa
    BINARYARRAY, BOOLEAN, BOOLEANARRAY, BYTES, BYTESARRAY, DATE, DATEARRAY,
    DATETIMEARRAY, DECIMAL, DECIMALARRAY, FLOAT, FLOATARRAY, INTEGER,
    INTEGERARRAY, INTERVAL, INTERVALARRAY, LONGINTEGER, LONGINTEGERARRAY,
    ROWIDARRAY, STRINGARRAY, TIME, TIMEARRAY, UNICODE, UNICODEARRAY,
    AsIs, Binary, Boolean, Float, Int, QuotedString, )

from psycopg2._psycopg import (                         # noqa
    PYDATE, PYDATETIME, PYDATETIMETZ, PYINTERVAL, PYTIME, PYDATEARRAY,
    PYDATETIMEARRAY, PYDATETIMETZARRAY, PYINTERVALARRAY, PYTIMEARRAY,
    DateFromPy, TimeFromPy, TimestampFromPy, IntervalFromPy, )

from psycopg2._psycopg import (                             # noqa
    adapt, adapters, encodings, connection, cursor,
    lobject, Xid, libpq_version, parse_dsn, quote_ident,
    string_types, binary_types, new_type, new_array_type, register_type,
    ISQLQuote, Notify, Diagnostics, Column, ConnectionInfo,
    QueryCanceledError, TransactionRollbackError,
    set_wait_callback, get_wait_callback, encrypt_password, )


"""Isolation level values."""
ISOLATION_LEVEL_AUTOCOMMIT = 0
ISOLATION_LEVEL_READ_UNCOMMITTED = 4
ISOLATION_LEVEL_READ_COMMITTED = 1
ISOLATION_LEVEL_REPEATABLE_READ = 2
ISOLATION_LEVEL_SERIALIZABLE = 3
ISOLATION_LEVEL_DEFAULT = None


"""psycopg connection status values."""
STATUS_SETUP = 0
STATUS_READY = 1
STATUS_BEGIN = 2
STATUS_SYNC = 3  # currently unused
STATUS_ASYNC = 4  # currently unused
STATUS_PREPARED = 5

STATUS_IN_TRANSACTION = STATUS_BEGIN


"""psycopg asynchronous connection polling values"""
POLL_OK = 0
POLL_READ = 1
POLL_WRITE = 2
POLL_ERROR = 3


"""Backend transaction status values."""
TRANSACTION_STATUS_IDLE = 0
TRANSACTION_STATUS_ACTIVE = 1
TRANSACTION_STATUS_INTRANS = 2
TRANSACTION_STATUS_INERROR = 3
TRANSACTION_STATUS_UNKNOWN = 4


def register_adapter(typ, callable):
    """Register 'callable' as an ISQLQuote adapter for type 'typ'."""
    adapters[(typ, ISQLQuote)] = callable


class SQL_IN:
    """Adapt any iterable to an SQL quotable object."""
    def __init__(self, seq):
        self._seq = seq
        self._conn = None

    def prepare(self, conn):
        self._conn = conn

    def getquoted(self):
        pobjs = [adapt(o) for o in self._seq]
        if self._conn is not None:
            for obj in pobjs:
                if hasattr(obj, 'prepare'):
                    obj.prepare(self._conn)
        qobjs = [o.getquoted() for o in pobjs]
        return b'(' + b', '.join(qobjs) + b')'

    def __str__(self):
        return str(self.getquoted())


class NoneAdapter:
    """Adapt None to NULL.

    This adapter is not used normally as a fast path in mogrify uses NULL,
    but it makes easier to adapt composite types.
    """
    def __init__(self, obj):
        pass

    def getquoted(self, _null=b"NULL"):
        return _null


def make_dsn(dsn=None, **kwargs):
    """Convert a set of keywords into a connection strings."""
    if dsn is None and not kwargs:
        return ''

    if not kwargs:
        parse_dsn(dsn)
        return dsn

    if 'database' in kwargs:
        if 'dbname' in kwargs:
            raise TypeError(
                "you can't specify both 'database' and 'dbname' arguments")
        kwargs['dbname'] = kwargs.pop('database')

    kwargs = {k: v for (k, v) in kwargs.items() if v is not None}

    if dsn is not None:
        tmp = parse_dsn(dsn)
        tmp.update(kwargs)
        kwargs = tmp

    dsn = " ".join(["{}={}".format(k, _param_escape(str(v)))
        for (k, v) in kwargs.items()])

    parse_dsn(dsn)

    return dsn


def _param_escape(s,
        re_escape=_re.compile(r"([\\'])"),
        re_space=_re.compile(r'\s')):
    """
    Apply the escaping rule required by PQconnectdb
    """
    if not s:
        return "''"

    s = re_escape.sub(r'\\\1', s)
    if re_space.search(s):
        s = "'" + s + "'"

    return s


from psycopg2._json import register_default_json, register_default_jsonb    # noqa

try:
    JSON, JSONARRAY = register_default_json()
    JSONB, JSONBARRAY = register_default_jsonb()
except ImportError:
    pass

del register_default_json, register_default_jsonb


from psycopg2. _range import Range                              # noqa
del Range


for k, v in list(encodings.items()):
    k = k.replace('_', '').replace('-', '').upper()
    encodings[k] = v

del k, v
