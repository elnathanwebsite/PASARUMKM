// Service Worker for Web Push Notifications

self.addEventListener('push', function(event) {
    let payload = {
        title: 'Pasar UMKM',
        body: 'Ada aktivitas terbaru di toko Anda!',
        icon: '/favicon/favicon-96x96.png',
        badge: '/favicon/favicon-96x96.png',
        url: '/dashboard.html'
    };

    if (event.data) {
        try {
            const parsed = event.data.json();
            payload = { ...payload, ...parsed };
        } catch (e) {
            payload.body = event.data.text();
        }
    }

    const options = {
        body: payload.body,
        icon: payload.icon,
        badge: payload.badge,
        data: {
            url: payload.url
        },
        vibrate: [100, 50, 100],
        actions: [
            { action: 'open', title: 'Buka' }
        ]
    };

    event.waitUntil(
        self.registration.showNotification(payload.title, options)
    );
});

self.addEventListener('notificationclick', function(event) {
    event.notification.close();
    
    let urlToOpen = '/dashboard.html';
    if (event.notification.data && event.notification.data.url) {
        urlToOpen = event.notification.data.url;
    }
    
    // Resolve relative path to absolute URL
    const targetUrl = new URL(urlToOpen, self.location.origin).href;

    event.waitUntil(
        clients.matchAll({ type: 'window', includeUncontrolled: true }).then(function(windowClients) {
            // Check if there is already a window open with this URL and focus it
            for (let i = 0; i < windowClients.length; i++) {
                const client = windowClients[i];
                if (client.url === targetUrl && 'focus' in client) {
                    return client.focus();
                }
            }
            // Otherwise open a new window
            if (clients.openWindow) {
                return clients.openWindow(targetUrl);
            }
        })
    );
});
