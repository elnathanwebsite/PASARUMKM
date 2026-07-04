// push-notification.js
// Client-side helper for registering Service Worker and managing Web Push subscriptions

window.initPushNotification = async function(userId, tursoClient) {
    if (!('serviceWorker' in navigator) || !('PushManager' in window)) {
        console.warn('Web Push Notifications are not supported in this browser.');
        return;
    }

    try {
        // 1. Register Service Worker (sw.js)
        const registration = await navigator.serviceWorker.register('/sw.js');
        console.log('Service Worker registered successfully:', registration);

        // 2. Request Notification Permission
        let permission = Notification.permission;
        if (permission === 'default') {
            permission = await Notification.requestPermission();
        }

        if (permission !== 'granted') {
            console.log('Notification permission not granted.');
            return;
        }

        // 3. Fetch VAPID Public Key from Backend
        const vapidResponse = await fetch('http://localhost:8000/api/push/vapid-public-key');
        const vapidData = await vapidResponse.json();
        if (!vapidData.public_key) {
            throw new Error('VAPID public key not found from backend');
        }

        // Helper to convert base64 url-safe to Uint8Array
        const urlBase64ToUint8Array = (base64String) => {
            const padding = '='.repeat((4 - base64String.length % 4) % 4);
            const base64 = (base64String + padding)
                .replace(/\-/g, '+')
                .replace(/_/g, '/');
            const rawData = window.atob(base64);
            const outputArray = new Uint8Array(rawData.length);
            for (let i = 0; i < rawData.length; ++i) {
                outputArray[i] = rawData.charCodeAt(i);
            }
            return outputArray;
        };

        const convertedVapidKey = urlBase64ToUint8Array(vapidData.public_key);

        // 4. Check existing push subscription
        let subscription = await registration.pushManager.getSubscription();

        if (!subscription) {
            // Subscribe the user
            subscription = await registration.pushManager.subscribe({
                userVisibleOnly: true,
                applicationServerKey: convertedVapidKey
            });
            console.log('New push subscription created:', subscription);
        }

        // 5. Save/update subscription in Turso DB
        const subJson = subscription.toJSON();
        
        if (subJson.endpoint && subJson.keys && subJson.keys.p256dh && subJson.keys.auth) {
            await tursoClient.execute({
                sql: "INSERT OR REPLACE INTO push_subscriptions (user_id, endpoint, p256dh, auth) VALUES (?, ?, ?, ?)",
                args: [userId, subJson.endpoint, subJson.keys.p256dh, subJson.keys.auth]
            });
            console.log('Push subscription successfully synchronized with database for user:', userId);
        } else {
            console.warn('Subscription keys are incomplete:', subJson);
        }

    } catch (error) {
        console.error('Error initializing Web Push Notifications:', error);
    }
};

// Helper function to send push notification from frontend by calling the backend api
window.sendPushNotification = async function(userId, title, body, url = "/dashboard.html") {
    try {
        const response = await fetch('http://localhost:8000/api/push/send', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                user_id: userId,
                title: title,
                body: body,
                url: url
            })
        });
        const result = await response.json();
        console.log('Push notification send result:', result);
        return result;
    } catch (err) {
        console.error('Failed to send push notification via backend API:', err);
        return null;
    }
};
