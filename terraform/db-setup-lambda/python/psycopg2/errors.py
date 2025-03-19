"""Error classes for PostgreSQL error codes
"""




def lookup(code):
    """Lookup an error code and return its exception class.

    Raise `!KeyError` if the code is not found.
    """
    from psycopg2._psycopg import sqlstate_errors   # avoid circular import
    return sqlstate_errors[code]
