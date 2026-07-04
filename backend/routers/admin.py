from fastapi import APIRouter, HTTPException
from database import execute_query

router = APIRouter()

@router.get("/transactions")
def get_all_transactions():
    try:
        # We fetch all transaction records from riwayat_pemesanan
        # To get the seller's business name, we join with users table on seller_uid
        query = """
            SELECT 
                r.id,
                r.buyer_name,
                u.businessName as seller_name,
                r.product_name,
                r.qty,
                r.total_price,
                r.status,
                r.payment_method,
                r.created_at
            FROM riwayat_pemesanan r
            LEFT JOIN users u ON r.seller_uid = u.uid
            ORDER BY r.created_at DESC
        """
        transactions = execute_query(query)
        return {"status": "success", "data": transactions}
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))
