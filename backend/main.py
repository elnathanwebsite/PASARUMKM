from fastapi import FastAPI, HTTPException, Depends
from fastapi.middleware.cors import CORSMiddleware
from pydantic import BaseModel
from database import execute_query
import os

app = FastAPI(
    title="Pasar UMKM API",
    description="Backend API for Pasar UMKM",
    version="1.0.0"
)

# Configure CORS for frontend access
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],  # Allows all origins for development
    allow_credentials=True,
    allow_methods=["*"],  # Allows all methods
    allow_headers=["*"],  # Allows all headers
)

@app.get("/")
def read_root():
    return {"message": "Welcome to Pasar UMKM API"}

@app.get("/api/health")
def health_check():
    try:
        # Simple query to check DB connection
        result = execute_query("SELECT 1 as test")
        return {"status": "healthy", "database": "connected"}
    except Exception as e:
        return {"status": "unhealthy", "error": str(e)}

from routers import products

# --- Placeholder Routers (to be implemented in separate files) ---
app.include_router(products.router, prefix="/api/products", tags=["products"])
# @app.include_router(users_router, prefix="/api/users", tags=["users"])
# @app.include_router(orders_router, prefix="/api/orders", tags=["orders"])

if __name__ == "__main__":
    import uvicorn
    uvicorn.run("main:app", host="0.0.0.0", port=8000, reload=True)
