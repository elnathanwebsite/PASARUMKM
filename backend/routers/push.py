from fastapi import APIRouter, HTTPException
from pydantic import BaseModel
from database import execute_query
from py_vapid import Vapid  # type: ignore
from cryptography.hazmat.primitives import serialization  # type: ignore
import base64
import json
from pywebpush import webpush, WebPushException  # type: ignore

router = APIRouter()

class PushPayload(BaseModel):
    user_id: str
    title: str
    body: str
    url: str = "/dashboard.html"

def get_or_create_vapid_keys():
    try:
        # Check if keys exist in DB settings
        res = execute_query("SELECT value FROM settings WHERE key = 'vapid_public_key'")
        if res and len(res) > 0:
            public_key = res[0]["value"]
            res_priv = execute_query("SELECT value FROM settings WHERE key = 'vapid_private_key'")
            private_key = res_priv[0]["value"]
            return public_key, private_key
        
        # If not, generate new keys
        v = Vapid()
        v.generate_keys()
        
        # Serialize public key to url-safe base64
        public_bytes = v.public_key.public_bytes(
            encoding=serialization.Encoding.X962,
            format=serialization.PublicFormat.UncompressedPoint
        )
        public_key = base64.urlsafe_b64encode(public_bytes).decode('utf-8').rstrip('=')
        
        # Serialize private key to PEM string
        private_key = v.private_pem().decode('utf-8')
        
        # Save to DB
        execute_query("INSERT OR REPLACE INTO settings (key, value) VALUES ('vapid_public_key', ?)", (public_key,))
        execute_query("INSERT OR REPLACE INTO settings (key, value) VALUES ('vapid_private_key', ?)", (private_key,))
        
        return public_key, private_key
    except Exception as e:
        print(f"Error in VAPID keys generation: {e}")
        # Return fallback or raise
        raise

@router.get("/vapid-public-key")
def get_vapid_public_key():
    try:
        pub, _ = get_or_create_vapid_keys()
        return {"public_key": pub}
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

@router.post("/send")
def send_push_notification(payload: PushPayload):
    try:
        # Get VAPID keys
        _, private_key = get_or_create_vapid_keys()
        
        # Fetch all active push subscriptions for this user_id
        subs = execute_query(
            "SELECT endpoint, p256dh, auth FROM push_subscriptions WHERE user_id = ?",
            (payload.user_id,)
        )
        
        if not subs:
            return {"status": "success", "sent_count": 0, "message": "No active subscriptions found for this user"}
        
        sent_count = 0
        deleted_count = 0
        
        # Payload message to send to the Service Worker
        notification_data = json.dumps({
            "title": payload.title,
            "body": payload.body,
            "url": payload.url
        })
        
        for sub in subs:
            subscription_info = {
                "endpoint": sub["endpoint"],
                "keys": {
                    "p256dh": sub["p256dh"],
                    "auth": sub["auth"]
                }
            }
            
            try:
                webpush(
                    subscription_info=subscription_info,
                    data=notification_data,
                    vapid_private_key=private_key,
                    vapid_claims={"sub": "mailto:admin@pasarumkm.com"}
                )
                sent_count += 1
            except WebPushException as ex:
                # If subscription is gone or invalid (404/410), delete it
                if ex.response is not None and ex.response.status_code in [404, 410]:
                    try:
                        execute_query(
                            "DELETE FROM push_subscriptions WHERE endpoint = ?",
                            (sub["endpoint"],)
                        )
                        deleted_count += 1
                    except Exception as del_err:
                        print(f"Failed to delete expired subscription: {del_err}")
                else:
                    print(f"Failed to send WebPush to endpoint {sub['endpoint']}: {ex}")
            except Exception as other_ex:
                print(f"Unexpected error sending WebPush: {other_ex}")
                
        return {
            "status": "success", 
            "sent_count": sent_count, 
            "deleted_expired_count": deleted_count
        }
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))
