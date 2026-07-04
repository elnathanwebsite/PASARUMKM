from fastapi import APIRouter, HTTPException
from pydantic import BaseModel
from typing import List, Optional
from database import execute_query

router = APIRouter()

class ProductModel(BaseModel):
    id: str
    uid: str
    name: str
    price: float
    description: str
    photoUrl: str
    stock: int
    category: str
    isQrisActive: int = 1
    isPaypalActive: int = 1
    shippingJNE: int = 1
    shippingJNT: int = 1
    shippingSicepat: int = 1
    shippingPos: int = 1
    shippingInstant: int = 1
    shippingPickup: int = 1

@router.get("/")
def get_products(uid: Optional[str] = None):
    try:
        if uid:
            query = "SELECT * FROM products WHERE uid = ?"
            products = execute_query(query, (uid,))
        else:
            query = "SELECT * FROM products"
            products = execute_query(query)
        return {"status": "success", "data": products}
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

@router.post("/")
def create_product(product: ProductModel):
    try:
        query = """
            INSERT INTO products (
                id, uid, name, price, description, photoUrl, stock, category,
                isQrisActive, isPaypalActive, shippingJNE, shippingJNT,
                shippingSicepat, shippingPos, shippingInstant, shippingPickup
            ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
        """
        params = (
            product.id, product.uid, product.name, product.price, product.description,
            product.photoUrl, product.stock, product.category, product.isQrisActive,
            product.isPaypalActive, product.shippingJNE, product.shippingJNT,
            product.shippingSicepat, product.shippingPos, product.shippingInstant, product.shippingPickup
        )
        execute_query(query, params)
        return {"status": "success", "message": "Product created successfully"}
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

@router.delete("/{product_id}")
def delete_product(product_id: str):
    try:
        query = "DELETE FROM products WHERE id = ?"
        execute_query(query, (product_id,))
        return {"status": "success", "message": "Product deleted successfully"}
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))
