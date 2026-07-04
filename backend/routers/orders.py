from fastapi import APIRouter, HTTPException
from pydantic import BaseModel
from typing import Optional
from database import execute_query

router = APIRouter()

class OrderModel(BaseModel):
    id: str
    user_id: str
    product_id: str
    store_id: str
    quantity: int
    total_price: float
    status: str = "pending"
    # Add other fields based on your DB schema

@router.get("/history/{user_id}")
def get_order_history(user_id: str):
    try:
        query = "SELECT * FROM riwayat_pemesanan WHERE user_id = ?"
        orders = execute_query(query, (user_id,))
        return {"status": "success", "data": orders}
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

@router.post("/checkout")
def place_order(order: OrderModel):
    try:
        query = """
            INSERT INTO riwayat_pemesanan (
                id, user_id, product_id, store_id, quantity, total_price, status
            ) VALUES (?, ?, ?, ?, ?, ?, ?)
        """
        params = (
            order.id, order.user_id, order.product_id, order.store_id, 
            order.quantity, order.total_price, order.status
        )
        execute_query(query, params)
        return {"status": "success", "message": "Order placed successfully"}
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))
