from fastapi import APIRouter, HTTPException
from pydantic import BaseModel
from typing import Optional
from database import execute_query

router = APIRouter()

class UserModel(BaseModel):
    uid: str
    email: str
    displayName: Optional[str] = None
    photoURL: Optional[str] = None
    phoneNumber: Optional[str] = None
    # Add other fields based on your DB schema

@router.get("/{uid}")
def get_user(uid: str):
    try:
        query = "SELECT * FROM users WHERE uid = ?"
        users = execute_query(query, (uid,))
        if not users:
            return {"status": "error", "message": "User not found"}
        return {"status": "success", "data": users[0]}
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

@router.post("/")
def create_or_update_user(user: UserModel):
    try:
        # Check if user exists
        check_query = "SELECT uid FROM users WHERE uid = ?"
        existing = execute_query(check_query, (user.uid,))
        
        if existing:
            query = """
                UPDATE users 
                SET email = ?, displayName = ?, photoURL = ?, phoneNumber = ?
                WHERE uid = ?
            """
            params = (user.email, user.displayName, user.photoURL, user.phoneNumber, user.uid)
        else:
            query = """
                INSERT INTO users (uid, email, displayName, photoURL, phoneNumber) 
                VALUES (?, ?, ?, ?, ?)
            """
            params = (user.uid, user.email, user.displayName, user.photoURL, user.phoneNumber)
            
        execute_query(query, params)
        return {"status": "success", "message": "User saved successfully"}
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))
