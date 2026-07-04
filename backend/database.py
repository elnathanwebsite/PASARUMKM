import os
from dotenv import load_dotenv
# pyrefly: ignore [missing-import]
import libsql_client

load_dotenv()

# Get database credentials from environment variables
DB_URL = os.getenv("TURSO_DATABASE_URL")
AUTH_TOKEN = os.getenv("TURSO_AUTH_TOKEN")

if not DB_URL or not AUTH_TOKEN:
    raise ValueError("Missing TURSO_DATABASE_URL or TURSO_AUTH_TOKEN in .env")

# Initialize connection client
# Create a global client instance for sync execution
client = libsql_client.create_client_sync(
    url=DB_URL,
    auth_token=AUTH_TOKEN
)

def execute_query(query: str, params: tuple = ()):
    """
    Executes a query using the libsql-client.
    """
    # Replace ? placeholders with regular DB-API style ? or ?1, ?2, etc. if needed,
    # but libsql_client supports positional arguments (list/tuple).
    try:
        # We pass params as args
        result = client.execute(query, list(params))
        
        # If it's a SELECT query (or returning rows), return a list of dicts
        if result.rows:
            columns = result.columns
            # Map each row's values to the corresponding column names
            return [dict(zip(columns, row)) for row in result.rows]
        
        return {"status": "success"}
    except Exception as e:
        print(f"Database error: {e}")
        raise
