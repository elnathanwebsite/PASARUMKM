import os
import requests
from fastapi import APIRouter, HTTPException
from pydantic import BaseModel
from typing import Optional
from database import execute_query

router = APIRouter()

LOUVIN_API_KEY = os.getenv("LOUVIN_API_KEY", "lv_e837b09fd975494bb25ab6eaeffef179")
LOUVIN_BASE_URL = "https://api.louvin.dev"

class SubscriptionRequest(BaseModel):
    uid: str
    customer_email: str
    customer_name: str

@router.post("/create")
def create_subscription(req: SubscriptionRequest):
    try:
        # 1. Hit Louvin API to create transaction for 60.000 IDR
        payload = {
            "amount": 60000,
            "payment_type": "qris",
            "customer_name": req.customer_name,
            "customer_email": req.customer_email,
            "description": "Upgrade Akun Pro - Pasar UMKM",
            "reference": f"PRO-{req.uid}"
        }
        
        headers = {
            "Content-Type": "application/json",
            "x-api-key": LOUVIN_API_KEY
        }
        
        response = requests.post(f"{LOUVIN_BASE_URL}/create-transaction", json=payload, headers=headers)
        data = response.json()
        
        if not data.get("success"):
            raise Exception(data.get("error", "Gagal menghubungi Louvin API"))
            
        return {"status": "success", "data": data}
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

@router.get("/check-status")
def check_subscription_status(transaction_id: str, uid: str):
    try:
        # Check status with Louvin
        headers = {"x-api-key": LOUVIN_API_KEY}
        response = requests.get(f"{LOUVIN_BASE_URL}/check-status?id={transaction_id}", headers=headers)
        data = response.json()
        
        if not data.get("success"):
            raise Exception(data.get("error", "Gagal cek status transaksi"))
            
        status = data.get("transaction", {}).get("status")
        
        if status == "settled":
            # Update user in database to PRO and add 10 credits limit
            query = """
                UPDATE users 
                SET isSubscribed = 1, aiLimit = aiLimit + 10, aiSearchLimit = aiSearchLimit + 10, aiDescLimit = aiDescLimit + 10, aiAnalysisLimit = aiAnalysisLimit + 10
                WHERE uid = ?
            """
            execute_query(query, (uid,))
            return {"status": "success", "payment_status": "settled", "message": "Akun berhasil diupgrade ke PRO"}
            
        return {"status": "success", "payment_status": status}
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))
